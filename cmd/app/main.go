package main

import (
	"flag"
	"github.com/s02190058/spa/internal/app"
	"github.com/s02190058/spa/internal/config"
	"log"
)

var configPath = flag.String("config", "configs/main.yml", "config path")

func main() {
	flag.Parse()
	cfg, err := config.New(*configPath)
	if err != nil {
		log.Fatalf("unable to read config: %v", err)
	}

	app.Run(cfg)
}
