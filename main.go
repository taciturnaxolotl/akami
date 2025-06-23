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
		Long: `
 █████╗ ██╗  ██╗ █████╗ ███╗   ███╗██╗
██╔══██╗██║ ██╔╝██╔══██╗████╗ ████║██║
███████║█████╔╝ ███████║██╔████╔██║██║
██╔══██║██╔═██╗ ██╔══██║██║╚██╔╝██║██║
██║  ██║██║  ██╗██║  ██║██║ ╚═╝ ██║██║
╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝

🌷 Akami — The cutsie hackatime helper`,
	}

	// diagnose command
	cmd.AddCommand(&cobra.Command{
		Use:   "doc",
		Short: "diagnose potential hackatime issues",
		RunE:  handler.Doctor,
		Args:  cobra.NoArgs,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "test",
		Short: "send a test heartbeat to hackatime or whatever api url you provide",
		RunE:  handler.TestHeartbeat,
		Args:  cobra.NoArgs,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "get your hackatime stats",
		RunE:  handler.Status,
		Args:  cobra.NoArgs,
	})

	cmd.PersistentFlags().StringP("url", "u", "", "The base url for the hackatime client")
	cmd.PersistentFlags().StringP("key", "k", "", "API key to use for authentication")

	// this is where we get the fancy fang magic ✨
	if err := fang.Execute(
		context.Background(),
		cmd,
	); err != nil {
		os.Exit(1)
	}
}
