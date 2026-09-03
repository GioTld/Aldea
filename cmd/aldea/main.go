package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgPath string

var root = &cobra.Command{
	Use:   "aldea",
	Short: "Aldea P2P storage client",
}

func main() {
	root.PersistentFlags().StringVar(&cfgPath, "config", "client.yaml", "path to client config file")

	root.AddCommand(cmdInit, cmdPut, cmdGet, cmdStatus)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
