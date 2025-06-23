package handler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/taciturnaxolotl/akami/styles"
	"github.com/taciturnaxolotl/akami/wakatime"
	"gopkg.in/ini.v1"
)

func Doctor() *cobra.Command {
	return &cobra.Command{
		Use:   "doc",
		Short: "diagnose potential hackatime issues",
		RunE: func(c *cobra.Command, _ []string) error {
			// check our os
			os_name := runtime.GOOS

			user_dir, err := os.UserHomeDir()
			if err != nil {
				return errors.New("somehow your user doesn't exist? fairly sure this should never happen; plz report this to @krn on slack or via email at me@dunkirk.sh")
			}
			hackatime_path := filepath.Join(user_dir, ".wakatime.cfg")

			if os_name != "linux" && os_name != "darwin" && os_name != "windows" {
				return errors.New("hmm you don't seem to be running a recognized os? you are listed as running " + styles.Fancy.Render(os_name) + "; can you plz report this to @krn on slack or via email at me@dunkirk.sh?")
			}

			c.Println("Looks like you are running", styles.Fancy.Render(os_name), "so lets take a look at", styles.Muted.Render(hackatime_path), "for your config")

			rawCfg, err := os.ReadFile(hackatime_path)
			if errors.Is(err, os.ErrNotExist) {
				return errors.New("you don't have a wakatime config file! go check " + styles.Muted.Render("https://hackatime.hackclub.com/my/wakatime_setup") + " for the instructions and then try this again")
			}

			cfg, err := ini.Load(rawCfg)
			if err != nil {
				return errors.New(err.Error())
			}

			settings, err := cfg.GetSection("settings")
			if err != nil {
				return errors.New("wow! your config file seems to be messed up and doesn't have a settings heading; can you follow the instructions at " + styles.Muted.Render("https://hackatime.hackclub.com/my/wakatime_setup") + " to regenerate it?\n\nThe raw error we got was: " + err.Error())
			}

			api_key := settings.Key("api_key").String()
			api_url := settings.Key("api_url").String()
			if api_key == "" {
				return errors.New("hmm 🤔 looks like you don't have an api_key in your config file? are you sure you have followed the setup instructions at " + styles.Muted.Render("https://hackatime.hackclub.com/my/wakatime_setup") + " correctly?")
			}
			if api_url == "" {
				return errors.New("hmm 🤔 looks like you don't have an api_url in your config file? are you sure you have followed the setup instructions at " + styles.Muted.Render("https://hackatime.hackclub.com/my/wakatime_setup") + " correctly?")
			}

			correctApiUrl := "https://hackatime.hackclub.com/api/hackatime/v1"
			if api_url != correctApiUrl {
				if api_url == "https://api.wakatime.com/api/v1" {
					client := wakatime.NewClient(api_key)
					_, err := client.GetStatusBar()

					if !errors.Is(err, wakatime.ErrUnauthorized) {
						return errors.New("turns out you were connected to wakatime.com instead of hackatime; since your key seems to work if you would like to keep syncing data to wakatime.com as well as to hackatime you can either setup a realy serve like " + styles.Muted.Render("https://github.com/JasonLovesDoggo/multitime") + " or you can wait for " + styles.Muted.Render("https://github.com/hackclub/hackatime/issues/85") + " to get merged in hackatime and have it synced there :)\n\nIf you want to import your wakatime.com data into hackatime then you can use hackatime v1 temporarily to connect your wakatime account and import (in settings under integrations at " + styles.Muted.Render("https://waka.hackclub.com") + ") and then click the import from hackatime v1 button at " + styles.Muted.Render("https://hackatime.hackclub.com/my/settings") + ".\n\n If you have more questions feel free to reach out to me (hackatime v1 creator) on slack (at @krn) or via email at me@dunkirk.sh")
					} else {
						return errors.New("turns out your config is connected to the wrong api url and is trying to use wakatime.com to sync time but you don't have a working api key from them. Go to " + styles.Muted.Render("https://hackatime.hackclub.com/my/wakatime_setup") + " to run the setup script and fix your config file")
					}
				}
				c.Println("\nYour api url", styles.Muted.Render(api_url), "doesn't match the expected url of", styles.Muted.Render(correctApiUrl), "however if you are using a custom forwarder or are sure you know what you are doing then you are probably fine")
			}

			client := wakatime.NewClientWithOptions(api_key, api_url)
			c.Println("\nChecking your coding stats for today...")
			duration, err := client.GetStatusBar()
			if err != nil {
				if errors.Is(err, wakatime.ErrUnauthorized) {
					return errors.New("Your config file looks mostly correct and you have the correct api url but when we tested your api_key it looks like it is invalid? Can you double check if the key in your config file is the same as at " + styles.Muted.Render("https://hackatime.hackclub.com/my/wakatime_setup") + "?")
				}

				return errors.New("Something weird happened with the hackatime api; if the error doesn't make sense then please contact @krn on slack or via email at me@dunkirk.sh\n\n" + styles.Bad.Render("Full error: "+err.Error()))
			}

			// Convert seconds to a formatted time string (hours, minutes, seconds)
			totalSeconds := duration.Data.GrandTotal.TotalSeconds
			hours := totalSeconds / 3600
			minutes := (totalSeconds % 3600) / 60
			seconds := totalSeconds % 60

			formattedTime := ""
			if hours > 0 {
				formattedTime += fmt.Sprintf("%d hours, ", hours)
			}
			if minutes > 0 || hours > 0 {
				formattedTime += fmt.Sprintf("%d minutes, ", minutes)
			}
			formattedTime += fmt.Sprintf("%d seconds", seconds)

			c.Println("\nSweet!!! Looks like your hackatime is configured properly! Looks like you have coded today for", styles.Fancy.Render(formattedTime))

			c.Println("\nSending one quick heartbeat to make sure everything is ship shape and then you should be good to go!")

			err = client.SendHeartbeat(wakatime.Heartbeat{
				Branch:           "main",
				Category:         "coding",
				CursorPos:        1,
				Entity:           filepath.Join(user_dir, "akami.txt"),
				Type:             "file",
				IsWrite:          true,
				Language:         "Go",
				LineNo:           1,
				LineCount:        4,
				Project:          "example",
				ProjectRootCount: 3,
				Time:             float64(time.Now().Unix()),
			})
			if err != nil {
				return errors.New("oh dear; looks like something went wrong when sending that heartbeat. " + styles.Bad.Render("Full error: \""+strings.TrimSpace(err.Error())+"\""))
			}

			c.Println("\n🥳 it worked! you are good to go! Happy coding 👋")

			return nil
		},
	}
}
