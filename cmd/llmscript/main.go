package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/statico/llmscript/internal/cli"
)

func main() {
	flag.Parse()

	if err := cli.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
