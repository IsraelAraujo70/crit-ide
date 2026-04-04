package main

import (
	"fmt"
	"os"

	"github.com/israelcorrea/crit-ide/internal/app"
)

func main() {
	var filePath string
	if len(os.Args) > 1 {
		filePath = os.Args[1]
	}

	a := app.New(filePath)
	if err := a.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "crit-ide: %v\n", err)
		os.Exit(1)
	}
}
