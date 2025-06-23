package handler

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
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

// Task status indicators
var spinnerChars = []string{"[|]", "[/]", "[-]", "[\\]"}
var TaskCompleted = "[*]"

// taskState holds shared state for the currently running task
type taskState struct {
	cancel  context.CancelFunc
	message string
}

// printTask prints a task with a spinning animation
func printTask(c *cobra.Command, message string) {
	// Create a cancellable context for this spinner
	ctx, cancel := context.WithCancel(c.Context())

	// Store cancel function so we can stop the spinner later
	if taskCtx, ok := c.Context().Value("taskState").(*taskState); ok {
		// Cancel any previously running spinner first
		if taskCtx.cancel != nil {
			taskCtx.cancel()
			// Small delay to ensure previous spinner is stopped
			time.Sleep(10 * time.Millisecond)
		}
		taskCtx.message = message
		taskCtx.cancel = cancel
	} else {
		// First task, create the state and store it
		state := &taskState{
			message: message,
			cancel:  cancel,
		}
		c.SetContext(context.WithValue(c.Context(), "taskState", state))
	}

	// Start spinner in background
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Clear line and print spinner with current character
				spinner := styles.Muted.Render(spinnerChars[i%len(spinnerChars)])
				c.Printf("\r\033[K%s %s", spinner, message)
				i++
			}
		}
	}()

	// Add a small random delay between 200-400ms to make spinner animation visible
	randomDelay := 200 + time.Duration(rand.Intn(201)) // 300-500ms
	time.Sleep(randomDelay * time.Millisecond)
}

// completeTask marks a task as completed
func completeTask(c *cobra.Command, message string) {
	// Cancel spinner
	if state, ok := c.Context().Value("taskState").(*taskState); ok && state.cancel != nil {
		state.cancel()
		// Small delay to ensure spinner is stopped
		time.Sleep(10 * time.Millisecond)
	}

	// Clear line and display success message
	c.Printf("\r\033[K%s %s\n", styles.Success.Render(TaskCompleted), message)
}

// errorTask marks a task as failed
func errorTask(c *cobra.Command, message string) {
	// Cancel spinner
	if state, ok := c.Context().Value("taskState").(*taskState); ok && state.cancel != nil {
		state.cancel()
		// Small delay to ensure spinner is stopped
		time.Sleep(10 * time.Millisecond)
	}

	// Clear line and display error message
	c.Printf("\r\033[K%s %s\n", styles.Bad.Render("[ ! ]"), message)
}

// warnTask marks a task as a warning
func warnTask(c *cobra.Command, message string) {
	// Cancel spinner
	if state, ok := c.Context().Value("taskState").(*taskState); ok && state.cancel != nil {
		state.cancel()
		// Small delay to ensure spinner is stopped
		time.Sleep(10 * time.Millisecond)
	}

	// Clear line and display warning message
	c.Printf("\r\033[K%s %s\n", styles.Warn.Render("[?]"), message)
}

var user_dir, err = os.UserHomeDir()

var testHeartbeat = wakatime.Heartbeat{
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
}

func Doctor(c *cobra.Command, _ []string) error {
	// Initialize a new context with task state
	c.SetContext(context.WithValue(context.Background(), "taskState", &taskState{}))

	// check our os
	printTask(c, "Checking operating system")

	os_name := runtime.GOOS

	user_dir, err := os.UserHomeDir()
	if err != nil {
		errorTask(c, "Checking operating system")
		return errors.New("somehow your user doesn't exist? fairly sure this should never happen; plz report this to @krn on slack or via email at me@dunkirk.sh")
	}
	hackatime_path := filepath.Join(user_dir, ".wakatime.cfg")

	if os_name != "linux" && os_name != "darwin" && os_name != "windows" {
		errorTask(c, "Checking operating system")
		return errors.New("hmm you don't seem to be running a recognized os? you are listed as running " + styles.Fancy.Render(os_name) + "; can you plz report this to @krn on slack or via email at me@dunkirk.sh?")
	}
	completeTask(c, "Checking operating system")

	c.Printf("Looks like you are running %s so lets take a look at %s for your config\n\n", styles.Fancy.Render(os_name), styles.Muted.Render(hackatime_path))

	printTask(c, "Checking wakatime config file")

	rawCfg, err := os.ReadFile(hackatime_path)
	if errors.Is(err, os.ErrNotExist) {
		errorTask(c, "Checking wakatime config file")
		return errors.New("you don't have a wakatime config file! go check " + styles.Muted.Render("https://hackatime.hackclub.com/my/wakatime_setup") + " for the instructions and then try this again")
	}

	cfg, err := ini.Load(rawCfg)
	if err != nil {
		errorTask(c, "Checking wakatime config file")
		return errors.New(err.Error())
	}

	settings, err := cfg.GetSection("settings")
	if err != nil {
		errorTask(c, "Checking wakatime config file")
		return errors.New("wow! your config file seems to be messed up and doesn't have a settings heading; can you follow the instructions at " + styles.Muted.Render("https://hackatime.hackclub.com/my/wakatime_setup") + " to regenerate it?\n\nThe raw error we got was: " + err.Error())
	}
	completeTask(c, "Checking wakatime config file")

	printTask(c, "Verifying API credentials")

	api_key := settings.Key("api_key").String()
	api_url := settings.Key("api_url").String()
	if api_key == "" {
		errorTask(c, "Verifying API credentials")
		return errors.New("hmm 🤔 looks like you don't have an api_key in your config file? are you sure you have followed the setup instructions at " + styles.Muted.Render("https://hackatime.hackclub.com/my/wakatime_setup") + " correctly?")
	}
	if api_url == "" {
		errorTask(c, "Verifying API credentials")
		return errors.New("hmm 🤔 looks like you don't have an api_url in your config file? are you sure you have followed the setup instructions at " + styles.Muted.Render("https://hackatime.hackclub.com/my/wakatime_setup") + " correctly?")
	}
	completeTask(c, "Verifying API credentials")

	printTask(c, "Validating API URL")

	correctApiUrl := "https://hackatime.hackclub.com/api/hackatime/v1"
	if api_url != correctApiUrl {
		if api_url == "https://api.wakatime.com/api/v1" {
			client := wakatime.NewClient(api_key)
			_, err := client.GetStatusBar()

			if !errors.Is(err, wakatime.ErrUnauthorized) {
				errorTask(c, "Validating API URL")
				return errors.New("turns out you were connected to wakatime.com instead of hackatime; since your key seems to work if you would like to keep syncing data to wakatime.com as well as to hackatime you can either setup a realy serve like " + styles.Muted.Render("https://github.com/JasonLovesDoggo/multitime") + " or you can wait for " + styles.Muted.Render("https://github.com/hackclub/hackatime/issues/85") + " to get merged in hackatime and have it synced there :)\n\nIf you want to import your wakatime.com data into hackatime then you can use hackatime v1 temporarily to connect your wakatime account and import (in settings under integrations at " + styles.Muted.Render("https://waka.hackclub.com") + ") and then click the import from hackatime v1 button at " + styles.Muted.Render("https://hackatime.hackclub.com/my/settings") + ".\n\n If you have more questions feel free to reach out to me (hackatime v1 creator) on slack (at @krn) or via email at me@dunkirk.sh")
			} else {
				errorTask(c, "Validating API URL")
				return errors.New("turns out your config is connected to the wrong api url and is trying to use wakatime.com to sync time but you don't have a working api key from them. Go to " + styles.Muted.Render("https://hackatime.hackclub.com/my/wakatime_setup") + " to run the setup script and fix your config file")
			}
		}
		warnTask(c, "Validating API URL")
		c.Printf("\nYour api url %s doesn't match the expected url of %s however if you are using a custom forwarder or are sure you know what you are doing then you are probably fine\n\n", styles.Muted.Render(api_url), styles.Muted.Render(correctApiUrl))
	} else {
		completeTask(c, "Validating API URL")
	}

	client := wakatime.NewClientWithOptions(api_key, api_url)
	printTask(c, "Checking your coding stats for today")

	duration, err := client.GetStatusBar()
	if err != nil {
		errorTask(c, "Checking your coding stats for today")
		if errors.Is(err, wakatime.ErrUnauthorized) {
			return errors.New("Your config file looks mostly correct and you have the correct api url but when we tested your api_key it looks like it is invalid? Can you double check if the key in your config file is the same as at " + styles.Muted.Render("https://hackatime.hackclub.com/my/wakatime_setup") + "?")
		}

		return errors.New("Something weird happened with the hackatime api; if the error doesn't make sense then please contact @krn on slack or via email at me@dunkirk.sh\n\n" + styles.Bad.Render("Full error: "+err.Error()))
	}
	completeTask(c, "Checking your coding stats for today")

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

	c.Printf("Sweet!!! Looks like your hackatime is configured properly! Looks like you have coded today for %s\n\n", styles.Fancy.Render(formattedTime))

	printTask(c, "Sending test heartbeat")

	err = client.SendHeartbeat(testHeartbeat)
	if err != nil {
		errorTask(c, "Sending test heartbeat")
		return errors.New("oh dear; looks like something went wrong when sending that heartbeat. " + styles.Bad.Render("Full error: \""+strings.TrimSpace(err.Error())+"\""))
	}
	completeTask(c, "Sending test heartbeat")

	c.Println("🥳 it worked! you are good to go! Happy coding 👋")

	return nil
}

func TestHeartbeat(c *cobra.Command, args []string) error {
	// Initialize a new context with task state
	c.SetContext(context.WithValue(context.Background(), "taskState", &taskState{}))

	printTask(c, "Validating arguments")

	configApiKey, _ := c.Flags().GetString("key")
	configApiURL, _ := c.Flags().GetString("url")

	// If either value is missing, try to load from config file
	if configApiKey == "" || configApiURL == "" {
		userDir, err := os.UserHomeDir()
		if err != nil {
			errorTask(c, "Validating arguments")
			return err
		}
		wakatimePath := filepath.Join(userDir, ".wakatime.cfg")

		cfg, err := ini.Load(wakatimePath)
		if err != nil {
			errorTask(c, "Validating arguments")
			return errors.New("config file not found and you haven't passed all arguments")
		}

		settings, err := cfg.GetSection("settings")
		if err != nil {
			errorTask(c, "Validating arguments")
			return errors.New("no settings section in your config")
		}

		// Only load from config if not provided as parameter
		if configApiKey == "" {
			configApiKey = settings.Key("api_key").String()
			if configApiKey == "" {
				errorTask(c, "Validating arguments")
				return errors.New("couldn't find an api_key in your config")
			}
		}

		if configApiURL == "" {
			configApiURL = settings.Key("api_url").String()
			if configApiURL == "" {
				errorTask(c, "Validating arguments")
				return errors.New("couldn't find an api_url in your config")
			}
		}
	}

	completeTask(c, "Arguments look fine!")

	printTask(c, "Loading api client")

	client := wakatime.NewClientWithOptions(configApiKey, configApiURL)
	_, err := client.GetStatusBar()
	if err != nil {
		errorTask(c, "Loading api client")
		return err
	}

	completeTask(c, "Loading api client")

	c.Println("Sending a test heartbeat to", styles.Muted.Render(configApiURL))

	printTask(c, "Sending test heartbeat")

	err = client.SendHeartbeat(testHeartbeat)

	if err != nil {
		errorTask(c, "Sending test heartbeat")
		return err
	}

	completeTask(c, "Sending test heartbeat")

	c.Println("❇️ test heartbeat sent!")

	return nil
}
