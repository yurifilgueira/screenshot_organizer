package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yurifilgueira/screenshot_organizer/agents"
	"github.com/zalando/go-keyring"
)

const BASE_CONFIG_FOLDER_NAME = "Screenshot_Organizer"
const CONFIG_FOLDER_NAME = "config"
const CONFIG_FILE_NAME = "app-config.json"
const CONFIG_DIRECTORY_FIELD = "screenshotsDirPath"
const SERVICE_NAME = "screenshot-organizer"

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

	screenshotsDirectory, apikey := loadConfigs()

	if screenshotsDirectory != "" {
		a.dirPath = screenshotsDirectory
	}

	if apikey != "" {
		newAgent, err := agents.NewScreenshotAgent(a.ctx, apikey)
		if err != nil {
			log.Fatal(err)
		}
		a.screenshotAgent = newAgent

		go a.startWatching()
	}

}

func loadConfigs() (screenshotsDirectory string, apikey string) {
	userConfigDir, err := os.UserConfigDir()

	if err != nil {
		log.Fatal(err)
	}
	configPath := filepath.Join(userConfigDir, BASE_CONFIG_FOLDER_NAME, CONFIG_FOLDER_NAME, CONFIG_FILE_NAME)

	data, err := os.ReadFile(configPath)

	if err != nil {
		log.Fatal(err)
	}

	result := map[string]string{}
	json.Unmarshal(data, &result)

	username, err := user.Current()

	if err != nil {
		log.Fatal(err)
	}

	apikey, err = keyring.Get(SERVICE_NAME, username.Username)

	if err != nil {
		fmt.Println(err)
	}

	return result[CONFIG_DIRECTORY_FIELD], apikey
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

func (a *App) SaveConfig(path string, key string) {

	oldPath := a.dirPath
	a.dirPath = path
	newAgent, err := agents.NewScreenshotAgent(a.ctx, key)
	if err != nil {
		log.Fatal(err)
	}
	a.screenshotAgent = newAgent

	persistScreenshotDirPathConfig(a)
	persistApiKeyConfig(key)

	if !a.isWatching {
		go a.startWatching()
		return
	}

	a.watcher.Add(a.dirPath)
	a.watcher.Remove(oldPath)
}

func persistScreenshotDirPathConfig(a *App) {
	configPath, err := os.UserConfigDir()

	if err != nil {
		log.Fatal(err)
	}

	appFolder := filepath.Join(configPath, BASE_CONFIG_FOLDER_NAME)

	_, err = os.Stat(appFolder)

	if os.IsNotExist(err) {
		os.MkdirAll(appFolder, 0755)
	}

	_, err = os.Stat(filepath.Join(appFolder, CONFIG_FOLDER_NAME))

	if os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(appFolder, CONFIG_FOLDER_NAME), 0755)
	}

	configFilePath := filepath.Join(appFolder, CONFIG_FOLDER_NAME, CONFIG_FILE_NAME)
	configData := map[string]string{CONFIG_DIRECTORY_FIELD: a.dirPath}
	jsonData, _ := json.Marshal(configData)
	os.WriteFile(configFilePath, jsonData, 0644)
}

func persistApiKeyConfig(key string) {
	service := SERVICE_NAME
	username, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	apiKey := key

	err = keyring.Set(service, username.Username, apiKey)
	if err != nil {
		log.Fatal(err)
	}

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
