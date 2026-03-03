package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/awaketai/issue2md/internal/cli"
)

func main() {
	args, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := cli.Run(args, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		var inputErr *cli.InputError
		if errors.As(err, &inputErr) {
			os.Exit(1)
		}
		os.Exit(2)
	}
}
