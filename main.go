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
		case "serve":
			cfg, err := loadConfig()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			printBanner(cfg)
			if err := startServer(cfg, nil); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			return
		}
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading config:", err)
		os.Exit(1)
	}

	if err := runUI(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func printBanner(cfg *Config) {
	ts := checkTailscale()
	fmt.Println()
	fmt.Println(`  ,~~~~~,`)
	fmt.Println(` (~~~~~~~)  zipp-nest v` + version)
	fmt.Println("  `~~~~~`")
	fmt.Println()
	if ts.running {
		fmt.Printf("  tailscale:  %s\n", ts.ip)
		fmt.Printf("  address:    %s:%d\n", ts.ip, cfg.Port)
	} else {
		fmt.Println("  tailscale:  not connected")
	}
	fmt.Printf("  storage:    %s\n", cfg.StoragePath)
	fmt.Println()
}
