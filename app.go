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
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	agent, err := agents.NewScreenshotAgent(ctx)
	if err != nil {
		log.Printf("Erro ao iniciar agente: %v", err)
		return
	}
	a.screenshotAgent = agent

	go a.startWatching()
}

func (a *App) startWatching() {
	userHomeDir, _ := os.UserHomeDir()

	possibleDirs := []string{
		filepath.Join(userHomeDir, "OneDrive", "Imagens", "Screenshots"),
		filepath.Join(userHomeDir, "OneDrive", "Pictures", "Screenshots"),
		filepath.Join(userHomeDir, "Pictures", "Screenshots"),
		filepath.Join(userHomeDir, "Imagens", "Screenshots"),
		"screenshots",
	}

	var screenshotDir string
	for _, dir := range possibleDirs {
		if _, err := os.Stat(dir); err == nil {
			screenshotDir = dir
			break
		}
	}

	if screenshotDir == "" {
		screenshotDir = "screenshots"
		os.MkdirAll(screenshotDir, 0755)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer watcher.Close()

	watcher.Add(screenshotDir)

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
