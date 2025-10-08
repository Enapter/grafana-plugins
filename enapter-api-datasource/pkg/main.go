package main

import (
	"fmt"
	"os"

	"github.com/Enapter/grafana-plugins/pkg/grafana"
)

func main() {
	if err := grafana.NewPlugin().Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}
}
