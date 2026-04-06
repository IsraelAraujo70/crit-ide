package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/israelcorrea/crit-ide/internal/app"
	"github.com/israelcorrea/crit-ide/internal/logger"
)

func main() {
	var filePath string
	debug := false

	for _, arg := range os.Args[1:] {
		switch arg {
		case "--debug", "-d":
			debug = true
		case "--help", "-h":
			fmt.Println("Usage: crit-ide [--debug] [file]")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  --debug, -d    Enable debug logging to /tmp/crit-ide/")
			fmt.Println("  --help,  -h    Show this help")
			os.Exit(0)
		default:
			filePath = arg
		}
	}

	if debug {
		logDir := filepath.Join(os.TempDir(), "crit-ide")
		if err := logger.Init(logDir); err != nil {
			fmt.Fprintf(os.Stderr, "crit-ide: failed to init logger: %v\n", err)
			os.Exit(1)
		}
		defer logger.Close()
	}

	a := app.New(filePath)
	if err := a.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "crit-ide: %v\n", err)
		os.Exit(1)
	}
}
