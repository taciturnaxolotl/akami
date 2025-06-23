package main

import (
	"context"
	"errors"
	"os"
	"runtime"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

func main() {
	// init our cobra command with a name and description
	cmd := &cobra.Command{
		Use:   "akami",
		Short: "🌷 the cutsie hackatime helper",
	}

	// add our lipgloss styles
	fancy := lipgloss.NewStyle().Foreground(lipgloss.Magenta).Bold(true).Italic(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.BrightBlue).Italic(true)

	// root diagnose command
	cmd.AddCommand(&cobra.Command{
		Use:   "doc",
		Short: "diagnose potential hackatime issues",
		RunE: func(c *cobra.Command, _ []string) error {
			// check our os
			os_name := runtime.GOOS

			user_dir, err := os.UserHomeDir()
			if err != nil {
				return errors.New("somehow your user doesn't exist? fairly sure this should never happen; plz report this to @krn on slack or via email at me@dunkirk.sh")
			}
			hackatime_path := user_dir + "/.wakatime.cfg"

			switch os_name {
			case "linux":
			case "darwin":
			case "windows":
			default:
				return errors.New("hmm you don't seem to be running a recognized os? you are listed as running " + fancy.Render(os_name) + "; can you plz report this to @krn on slack or via email at me@dunkirk.sh?")
			}

			c.Println("Looks like you are running", fancy.Render(os_name), "so lets take a look at", muted.Render(hackatime_path), "for your config")

			rawCfg, err := os.ReadFile(hackatime_path)
			if errors.Is(err, os.ErrNotExist) {
				return errors.New("you don't have a wakatime config file! go check https://hackatime.hackclub.com/my/wakatime_setup for the instructions and then try this again")
			}

			cfg, err := ini.Load(rawCfg)
			if err != nil {
				return errors.New(err.Error())
			}

			settings, err := cfg.GetSection("settings")
			if err != nil {
				return errors.New("wow! your config file seems to be messed up and doesn't have a settings heading; can you follow the instructions at https://hackatime.hackclub.com/my/wakatime_setup to regenerate it?\n\nThe raw error we got was: " + err.Error())
			}

			api_key := settings.Key("api_key").String()
			api_url := settings.Key("api_url").String()
			if api_key == "" {
				return errors.New("hmm 🤔 looks like you don't have an api_key in your config file? are you sure you have followed the setup instructions at https://hackatime.hackclub.com/my/wakatime_setup correctly?")
			}
			if api_url == "" {
				return errors.New("hmm 🤔 looks like you don't have an api_url in your config file? are you sure you have followed the setup instructions at https://hackatime.hackclub.com/my/wakatime_setup correctly?")
			}

			if api_url != "https://hackatime.hackclub.com/api/hackatime/v1" {
				c.Println("\nYour api url", muted.Render(api_url), "doesn't match the expected url of", muted.Render("https://hackatime.hackclub.com/api/hackatime/v1"), "however if you are using a custom forwarder or are sure you know what you are doing then you are probably fine")
			}

			return nil
		},
	})

	// this is where we get the fancy fang magic ✨
	if err := fang.Execute(
		context.Background(),
		cmd,
	); err != nil {
		os.Exit(1)
	}
}
