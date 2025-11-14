package main

import (
	"log"

	"github.com/Mockird31/avito_tech/config"
	"github.com/Mockird31/avito_tech/internal/app"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.Run(cfg)
}
