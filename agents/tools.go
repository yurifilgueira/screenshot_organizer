package agents

import (
	"fmt"
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
			fmt.Println(folder.Name())
			foldersNames = append(foldersNames, folder.Name())
		}
	}

	return ReadDirectoriesNameResult{foldersNames}, nil

}
