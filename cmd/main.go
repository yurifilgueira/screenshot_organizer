package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/yurifilgueira/screenshot_organizer/agents"
)

func main() {

	ctx := context.Background()
	screenshotAgent, err := agents.NewScreenshotAgent(ctx)

	if err != nil {
		log.Fatal(err)
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	screnshotDir := filepath.Join(userHomeDir, "OneDrive", "Imagens", "Screenshots")

	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Directory does not exist")
		} else {
			log.Fatal(err)
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	watcher.Add(screnshotDir)
	go watchLoop(ctx, watcher, screenshotAgent)

	<-make(chan bool)

}

func watchLoop(ctx context.Context, watcher *fsnotify.Watcher, screenshotAgent *agents.ScreenshotAgent) {
	for {
		select {
		case event := <-watcher.Events:
			if event.Op == fsnotify.Write {

				info, err := os.Stat(event.Name)
				if err != nil {
					log.Fatal(err)
				}

				if info.IsDir() {
					continue
				}

				log.Printf("Modified file: %s\n", event.Name)
				response, _ := screenshotAgent.Organize(ctx, event.Name)

				fmt.Printf("Response: %s\n", response)
			}
		case err := <-watcher.Errors:
			log.Printf("Error: %v\n", err)
		}
	}
}
