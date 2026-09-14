package main

import (
	"API-mirror/api/ollama"
	"API-mirror/internal/cli"
	"API-mirror/internal/server"
	"log"
	"os"
)

func main() {

	file, err := os.OpenFile("server.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	logger := log.New(file, "", log.Ldate|log.Ltime)
	logger.Println("Server logging started")

	instance, err := ollama.NewInstance()
	if err != nil {
		log.Fatal(err)
	}

	instances := map[string]*server.ServerInstance{"ollama": instance}
	cliInput := cli.SartCLIReader()
	serverEvents := make(chan server.ServerEvent)

	instance.Start(serverEvents)
	log.Println("listening on :11434")

	cli.RunSelectLoop(instances, cliInput, serverEvents)
}
