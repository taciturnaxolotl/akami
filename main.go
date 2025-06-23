package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/taciturnaxolotl/akami/handler"
)

func main() {
	// init our cobra command with a name and description
	cmd := &cobra.Command{
		Use:   "akami",
		Short: "🌷 the cutsie hackatime helper",
	}

	// diagnose command
	cmd.AddCommand(handler.Doctor())

	// this is where we get the fancy fang magic ✨
	if err := fang.Execute(
		context.Background(),
		cmd,
	); err != nil {
		os.Exit(1)
	}
}
