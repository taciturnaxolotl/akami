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
	})

	cmdTest := &cobra.Command{
		Use:   "test",
		Short: "send a test heartbeat to hackatime or whatever api url you provide",
		RunE:  handler.TestHeartbeat,
		Args:  cobra.NoArgs,
	}
	cmdTest.Flags().StringP("url", "u", "", "The base url for the hackatime client")
	cmdTest.Flags().StringP("key", "k", "", "API key to use for authentication")
	cmd.AddCommand(cmdTest)

	// this is where we get the fancy fang magic ✨
	if err := fang.Execute(
		context.Background(),
		cmd,
	); err != nil {
		os.Exit(1)
	}
}
