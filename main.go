package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		switch args[0] {
		case "--version", "-v":
			fmt.Println("zipp-nest v" + version)
			return
		case "token":
			cfg, err := loadConfig()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			fmt.Println(cfg.Token)
			return
		case "serve", "":
			// fall through to serve
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading config:", err)
		os.Exit(1)
	}

	printBanner(cfg)

	if err := startServer(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "server error:", err)
		os.Exit(1)
	}
}

func printBanner(cfg *Config) {
	fmt.Println()
	fmt.Println(`  )()( `)
	fmt.Println(` ( ●● )  zipp-nest v` + version)
	fmt.Println(`  \──/`)
	fmt.Println(`  /||\`)
	fmt.Println()
	fmt.Printf("  token:    %s\n", cfg.Token)
	fmt.Printf("  port:     %d\n", cfg.Port)
	fmt.Printf("  storage:  %s\n", cfg.StoragePath)
	fmt.Println()
	fmt.Println("  ─────────────────────────────────────")
	fmt.Println("  time      job                  event")
	fmt.Println("  ─────────────────────────────────────")
}
