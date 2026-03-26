package agents

import (
	"log"
	"os"
	"path/filepath"

	"google.golang.org/adk/tool"
)

type ReadDirectoriesNameArgs struct {
}

type ReadDirectoriesNameResult struct {
	FoldersNames []string `json:"foldersNames" jsonschema:"The names of the folders in the Screenshot directory."`
}

func readDirectoriesNameTool(ctx tool.Context, args ReadDirectoriesNameArgs) (ReadDirectoriesNameResult, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	directory := filepath.Join(userHomeDir, "OneDrive", "Imagens", "Screenshots")

	folders, err := os.ReadDir(directory)

	if err != nil {
		log.Fatal(err)
	}

	foldersNames := []string{}
	for _, folder := range folders {
		if folder.IsDir() {
			foldersNames = append(foldersNames, folder.Name())
		}
	}

	return ReadDirectoriesNameResult{foldersNames}, nil

}

type MoveScreenshotToDirectoryArgs struct {
	FolderName string `json:"folderName" jsonschema:"The name of the folder to move the screenshot to."`
	FilePath   string `json:"filePath" jsonschema:"The path of the screenshot to move."`
}

type MoveScreenshotToDirectoryResult struct {
	FileMoved bool `json:"fileMoved"`
}

func moveScreenshotToDirectoryTool(ctx tool.Context, args MoveScreenshotToDirectoryArgs) (MoveScreenshotToDirectoryResult, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	baseDir := filepath.Join(userHomeDir, "OneDrive", "Imagens", "Screenshots")

	newDir := filepath.Join(baseDir, args.FolderName)

	info, err := os.Stat(newDir)

	if os.IsNotExist(err) {
		err = os.Mkdir(newDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	if info != nil && !info.IsDir() {
		log.Fatal(err)
	}

	os.Rename(args.FilePath, filepath.Join(newDir, filepath.Base(args.FilePath)))

	return MoveScreenshotToDirectoryResult{true}, nil

}
