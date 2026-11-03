// Command mmmbed generates source code containing binary assets.
package main

import (
	"fmt"
	"github.com/zfogg/mmmbed/internal/cli"
	"os"
)

func main() {
	if err := cli.New().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
