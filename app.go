package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sync"

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
	hideOnClose     bool
	appConfig       AppConfig
}

func NewApp() *App {
	return &App{isWatching: false, hideOnClose: false}
}

type AppConfig struct {
	ScreenshotsDirPath string `json:"screenshotsDirPath"`
	ApiKey             string `json:"apiKey"`
	HideOnClose        bool   `json:"hideOnClose"`
}

type ConfigManager struct {
	Data       AppConfig
	mutex      sync.Mutex
	configPath string
}

func NewConfigManager() *ConfigManager {

	configPath := getConfigFilePath()
	cm := &ConfigManager{
		configPath: configPath,
		mutex:      sync.Mutex{},
	}

	file, err := os.ReadFile(configPath)

	if err == nil {
		json.Unmarshal(file, &cm.Data)
	} else {
		cm.Data = AppConfig{HideOnClose: false}
	}

	return cm
}

func (a *App) GetConfig() *AppConfig {
	err := a.loadConfigs()

	if err != nil {
		a.appConfig = AppConfig{HideOnClose: false}
	}

	return &a.appConfig
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	err := a.loadConfigs()

	if err != nil {
		a.appConfig = AppConfig{HideOnClose: false}
	}

	if a.appConfig.ScreenshotsDirPath != "" && a.appConfig.ApiKey != "" {
		a.dirPath = a.appConfig.ScreenshotsDirPath

		newAgent, err := agents.NewScreenshotAgent(a.ctx, a.appConfig.ApiKey)
		if err != nil {
			log.Fatal(err)
		}
		a.screenshotAgent = newAgent

		go a.startWatching()
	}

}

func (a *App) loadConfigs() error {

	err := loadScreenshotDirPathConfig(&a.appConfig)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = loadApikeyConfig(&a.appConfig)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func loadApikeyConfig(config *AppConfig) error {
	username, err := user.Current()

	if err != nil {
		fmt.Println(err)
		return err
	}

	config.ApiKey, err = keyring.Get(SERVICE_NAME, username.Username)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func loadScreenshotDirPathConfig(config *AppConfig) error {
	configPath := getConfigFilePath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	return nil
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

	configData, err := json.MarshalIndent(a.appConfig, "", "  ")

	if err != nil {
		log.Println(err)
		return
	}

	os.WriteFile(getConfigFilePath(), configData, 0644)

	persistApiKeyConfig(key)

	if !a.isWatching {
		go a.startWatching()
		return
	}

	a.watcher.Add(a.dirPath)
	a.watcher.Remove(oldPath)
}

func getConfigFilePath() string {
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
	return configFilePath
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

func (a *App) beforeClose(ctx context.Context) bool {

	if a.hideOnClose {
		runtime.WindowHide(ctx)
		return true
	}

	return false
}

func (a *App) SetHideOnClose(hideOnClose bool) {
	a.hideOnClose = hideOnClose

	a.SaveConfig(a.dirPath, a.appConfig.ApiKey)
}
