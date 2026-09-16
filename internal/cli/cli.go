package cli

import (
	"API-mirror/internal/config"
	"API-mirror/internal/logging"
	"API-mirror/internal/server"
	"bufio"
	"context"
	"log"
	"os"
	"strings"
	"time"
)

var QuitTimeout = 5 * time.Second

func RunSelectLoop(instances map[string]*server.ServerInstance, cliInput <-chan string, serverEvents chan server.ServerEvent,
	providerBuilder map[string]func(config.InstanceConfig) (*server.ServerInstance, error)) {

	for {
		select {
		case line, ok := <-cliInput:
			if !ok {
				log.Println("cli input closed, shutting down")
				return
			}

			parts := strings.Fields(line)
			cmd := parts[0]
			arg := ""
			if len(parts) > 1 {
				arg = parts[1]
			}
			log.Printf("received command: %q", line)

			switch cmd {
			case "status":
				log.Println("action: status")
				for name, instance := range instances {
					if arg != "" && name != arg {
						continue
					}
					log.Printf(" %s: port%s running: %t", name, instance.Server.Addr, instance.Running)
				}

			case "routes":
				log.Println("action: routes")
				for name, instance := range instances {
					if arg != "" && name != arg {
						continue
					}
					for _, r := range instance.Registry.Routes() {
						log.Printf(" %s : %s", name, r)
					}
				}

			case "quit":
				force := arg == "--force"
				log.Printf("action: quit force: %t", force)
				pending := map[string]bool{}
				for name, instance := range instances {
					if !instance.Running {
						continue
					}
					pending[name] = true
					go func(name string, instance *server.ServerInstance) {
						ctx, cancel := context.WithTimeout(context.Background(), server.ShutdownTimeout)
						defer cancel()
						if err := instance.Stop(ctx); err != nil {
							log.Printf("stop %s: shutdown error :%v", name, err)
						}
					}(name, instance)
				}
				if force || len(pending) == 0 {
					return
				}
				timeout := time.After(QuitTimeout)
				results := map[string]error{}
				for len(pending) > 0 {
					select {
					case ev := <-serverEvents:
						if !pending[ev.Instance.Name] {
							log.Printf("quit: unexpected event from %q", ev.Instance.Name)
							continue
						}
						results[ev.Instance.Name] = ev.Err
						delete(pending, ev.Instance.Name)
						log.Printf("quit: %s stopped (err=%v)", ev.Instance.Name, ev.Err)

					case <-timeout:
						names := make([]string, 0, len(pending))
						for name := range pending {
							names = append(names, name)
						}
						log.Printf("quit: timed out waiting on: %v", names)
						return
					}
				}
				log.Println("quit: all instances stopped cleanly")
				return

			case "reload":
				if arg == "" {
					log.Println("usage: reload <name>")
					break
				}
				log.Println("action: reload")
				instance, ok := instances[arg]
				if !ok {
					log.Printf("unknown instance %q", arg)
					break
				}
				if err := instance.ReloadAll(); err != nil {
					log.Printf("reload failed: %v", err)
				} else {
					log.Println("reload succeeded")
				}

			case "start":
				if arg == "" {
					log.Println("usage: start <name>")
					break
				}
				instance, ok := instances[arg]
				if !ok {
					log.Printf("unknown instance: %q", arg)
					break
				}
				if ok && instance.Running {
					log.Printf("start failed: %s already running", arg)
					break
				}
				var cfg config.InstanceConfig
				if ok {
					cfg = instance.Config
				}
				builder := providerBuilder[cfg.Type]
				newInstance, err := builder(cfg)
				if err != nil {
					log.Printf("start %s: rebuild failed: %v", arg, err)
					break
				}
				newInstance.Running = true
				instances[arg] = newInstance
				newInstance.Start(serverEvents)
				log.Printf("action: start %s", arg)

			case "stop":
				if arg == "" {
					log.Println("usage: stop <name>")
					break
				}
				instance, ok := instances[arg]
				if !ok {
					log.Printf("unknown instance: %q", arg)
					break
				}
				if !instance.Running {
					log.Printf("stop failed: %s not running", arg)
					break
				}
				log.Printf("action: stop %s", arg)
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), server.ShutdownTimeout)
					defer cancel()
					if err := instance.Stop(ctx); err != nil {
						log.Printf("stop %s: shutdown error: %v", instance.Name, err)
					}
				}()

			case "log":
				var ok bool
				switch arg {
				case "quiet":
					logging.Set(logging.Quiet)
					ok = true
				case "normal":
					logging.Set(logging.Normal)
					ok = true
				case "verbose":
					logging.Set(logging.Verbose)
					ok = true
				default:
					log.Println("usage: log <quiet|normal|verbose>")
				}
				if ok {
					log.Printf("action: log %s", arg)
				}

			default:
				log.Printf("action: unknown command %q", cmd)
			}

		case ev := <-serverEvents:
			current, ok := instances[ev.Instance.Name]
			if !ok || current != ev.Instance {
				log.Printf("stale event from replaced %q, ignoring", ev.Instance.Name)
				break
			}
			ev.Instance.Running = false

			if ev.Err != nil {
				log.Printf("server %s failed: %v", ev.Instance.Name, ev.Err)
			} else {
				log.Printf("server %s stopped cleanly", ev.Instance.Name)
			}
		}
	}
}

func SartCLIReader() chan string {
	cliInput := make(chan string)

	go func() {
		defer close(cliInput)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			cliInput <- line
		}
		if err := scanner.Err(); err != nil {
			log.Printf("cli reader error: %w", err)
		}
	}()
	return cliInput
}
