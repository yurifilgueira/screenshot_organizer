package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func main() {

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
	go watchLoop(watcher)

	<-make(chan bool)

}

func watchLoop(watcher *fsnotify.Watcher) {
	for {
		select {
		case event := <-watcher.Events:
			if event.Op == fsnotify.Write {
				log.Printf("Modified file: %s\n", event.Name)
			}
		case err := <-watcher.Errors:
			log.Printf("Error: %v\n", err)
		}
	}
}
