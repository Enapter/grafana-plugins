package main

import (
	"fmt"
	"os"

	"github.com/Enapter/grafana-plugins/pkg/plugin"
)

func main() {
	if err := plugin.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}
}
