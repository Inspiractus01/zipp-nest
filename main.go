package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
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
		case "uninstall":
			if err := uninstallService(); err != nil {
				fmt.Fprintln(os.Stderr, "error stopping service:", err)
			}
			if err := os.Remove("/usr/local/bin/zipp-nest"); err != nil && !os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, "error removing binary:", err)
				os.Exit(1)
			}
			fmt.Println("zipp-nest uninstalled")
			return
		}
	}

	if result := checkForUpdate(); result.hasUpdate {
		fmt.Printf("\n  zipp-nest v%s → v%s  updating...\n\n", version, result.latest)
		cmd := exec.Command("bash", "-c",
			"curl -sL https://raw.githubusercontent.com/Inspiractus01/zipp-nest/main/install.sh | bash",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "\n  update failed: %v\n\n", err)
		} else {
			self, err := os.Executable()
			if err == nil {
				syscall.Exec(self, os.Args, os.Environ())
			}
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
