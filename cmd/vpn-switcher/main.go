package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/config"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/logging"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/network"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/ui"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/vpn"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	configPath := flag.String("config", config.DefaultConfigPath(), "path to config file")
	debug := flag.Bool("debug", false, "enable debug logging")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("vpn-switcher %s (%s)\n", version, commit)
		return
	}

	logger, err := logging.Setup(*debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: setup logging: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load config: %v\n", err)
		os.Exit(1)
	}

	var exec system.CommandExecutor
	exec = system.NewRealExecutor()
	if *debug {
		exec = logging.NewLoggingExecutor(exec, logger)
	}

	configDirs := config.ExpandConfigDirs(cfg.General.OpenVPNConfigDirs)
	ovpn := vpn.NewOpenVPNProvider(exec, configDirs, cfg.General.OpenVPNMethod)
	ts := vpn.NewTailscaleProvider(exec)
	mgr := vpn.NewManager(exec, ovpn, ts)

	rmgr := network.NewRoutingManager(exec)
	ipf := network.NewIPFetcher(cfg.General.IPCheckURL)
	imon := network.NewInterfaceMonitor(exec)

	app := ui.NewApp(mgr, rmgr, ipf, imon, exec, cfg)

	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
