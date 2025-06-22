package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "akami",
		Short: "🌷 the cutsie hackatime helper",
	}
	if err := fang.Execute(context.TODO(), cmd); err != nil {
		os.Exit(1)
	}
}
