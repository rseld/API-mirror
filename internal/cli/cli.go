package cli

import (
	"API-mirror/internal/server"
	"bufio"
	"log"
	"os"
	"strings"
)

func RunSelectLoop(instances map[string]*server.ServerInstance, cliInput <-chan string, serverEvents <-chan server.ServerEvent) {
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
					log.Printf(" %s: %s", name, instance.Server.Addr)
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
				log.Println("action: quit")
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

			case "log", "start", "stop":
				log.Printf("action: %s (unimplemented)", cmd)

			default:
				log.Printf("action: unknown command %q", cmd)
			}

		case ev := <-serverEvents:
			if ev.Err != nil {
				log.Printf("server %s failed: %v", ev.Name, ev.Err)
			} else {
				log.Printf("server %s stopped cleanly", ev.Name)
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
