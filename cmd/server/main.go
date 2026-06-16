package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Ddhjx-code/AgentHub/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
		os.Exit(1)
	}

	fmt.Printf("AgentHub server starting on :%d\n", cfg.Server.Port)
}
