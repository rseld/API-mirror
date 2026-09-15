package main

import (
	"API-mirror/api/ollama"
	"API-mirror/internal/cli"
	"API-mirror/internal/config"
	"API-mirror/internal/server"
	"flag"
	"log"
	"os"
)

func main() {

	configPath := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	cfg := config.LoadOrDefault(*configPath)

	providerBuilder := map[string]func(config.InstanceConfig) (*server.ServerInstance, error){
		"ollama": ollama.NewInstance,
	}

	file, err := os.OpenFile("server.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	logger := log.New(file, "", log.Ldate|log.Ltime)
	logger.Println("Server logging started")

	instances := map[string]*server.ServerInstance{}
	for _, icfg := range cfg.Instances {
		builder, ok := providerBuilder[icfg.Type]
		if !ok {
			log.Fatalf("unknown provider type %q for instance %q", icfg.Type, icfg.Name)
		}
		instance, err := builder(icfg)
		if err != nil {
			log.Fatalf("building instance %q: %v", icfg.Name, err)
		}
		instances[icfg.Name] = instance
	}
	if err != nil {
		log.Fatal(err)
	}

	cliInput := cli.SartCLIReader()
	serverEvents := make(chan server.ServerEvent)

	cli.RunSelectLoop(instances, cliInput, serverEvents, providerBuilder)
}
