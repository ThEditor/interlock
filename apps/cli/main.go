package main

import (
	"context"
	"log"
	"os"

	"cli/cmd"
)

func main() {
	rootCmd := cmd.NewRootCommand()

	if err := rootCmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
