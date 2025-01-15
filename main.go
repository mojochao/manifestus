package main

import (
	"os"

	"github.com/mojochao/manifestus/cli"
)

func main() {
	if err := cli.New().Run(os.Args); err != nil {
		os.Exit(1)
	}
}
