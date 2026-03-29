package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yurifilgueira/screenshot_organizer/agents"
)

type App struct {
	ctx             context.Context
	screenshotAgent *agents.ScreenshotAgent
	isWatching      bool
	dirPath         string
	watcher         *fsnotify.Watcher
}

func NewApp() *App {
	return &App{isWatching: false}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	agent, err := agents.NewScreenshotAgent(ctx)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	a.screenshotAgent = agent
}

func (a *App) startWatching() {

	screenshotDir := a.dirPath

	if screenshotDir == "" || !filepath.IsAbs(screenshotDir) {
		log.Fatal("Screenshot directory path is not set or not absolute.")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()

	watcher.Add(screenshotDir)

	a.isWatching = true
	a.watcher = watcher

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				info, err := os.Stat(event.Name)
				if err != nil || info.IsDir() {
					continue
				}

				runtime.EventsEmit(a.ctx, "processing-start", event.Name)

				response, err := a.screenshotAgent.Organize(a.ctx, event.Name)
				if err != nil {
					runtime.EventsEmit(a.ctx, "processing-error", err.Error())
					continue
				}

				runtime.EventsEmit(a.ctx, "new-result", map[string]string{
					"filename": filepath.Base(event.Name),
					"category": response,
				})
			}
		case err := <-watcher.Errors:
			log.Printf("Error: %v", err)
		}
	}
}

func (a *App) SaveConfig(path string, _key string) {

	oldPath := a.dirPath
	a.dirPath = path

	if !a.isWatching {
		go a.startWatching()
		return
	}

	a.watcher.Add(a.dirPath)
	a.watcher.Remove(oldPath)
}

func (a *App) SelectDirectory() string {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select the Screenshot Directory",
	})
	if err != nil {
		log.Printf("Error opening directory: %v", err)
		return ""
	}
	return selection
}
