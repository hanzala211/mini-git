package commands

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hanzalaoc211/mini-git/common"
	"github.com/spf13/cobra"
)

func AddFileToIndex(repoRoot string, filePath string, index common.Index) {
	absFilePath, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absFilePath, repoRoot) {
		log.Fatalf("file %s is outside repository", filePath)
	}
	content, err := os.ReadFile(absFilePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	hash, err := common.WriteObject(repoRoot, content, common.BlobFile, filePath)
	if err != nil {
		log.Fatal(err)
	}
	relativeFilePath := filepath.ToSlash(filePath)
	index[relativeFilePath] = hash
}

func AddCommand(cmd *cobra.Command, args []string) {
	repoPath, err := common.FindRepoRoot()
	if err != nil {
		log.Fatal(err)
	}
	minigitDir, err := filepath.Abs(filepath.Join(repoPath, common.RootDir))// find the absolute path to minigit dir to skip it during 
	// recursive walk
	if err != nil {
		log.Fatal(err)
	}
	gitDir, err := filepath.Abs(filepath.Join(repoPath, ".git"))// find the absolute path to git dir to skip it during 
	// recursive walk
	if err != nil {
		log.Fatal(err)
	}
	indexBytes, err := os.ReadFile(filepath.Join(repoPath, common.RootDir, common.IndexFile))
	if err != nil {
		log.Fatal("Failed to read index")
	}
	var index common.Index
	json.Unmarshal(indexBytes, &index)
	
	for _, pathArg := range args {
		stat, err := os.Stat(pathArg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error stating path %s: %v\n", pathArg, err)
			continue // Skip to the next argument
		}
		if stat.IsDir() {
			err := filepath.Walk(pathArg, func(path string, info fs.FileInfo, err error) error { // filepath.Walk to recursively walk the directory
				if err != nil {
					return err // Handle error from walking
				}
				absPath, _ := filepath.Abs(path)
				if info.IsDir() && (absPath == minigitDir || absPath == gitDir) {
					return filepath.SkipDir
				}

				if !info.IsDir() {
					AddFileToIndex(repoPath, path, index)
				}
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}
		}else { 
			AddFileToIndex(repoPath, pathArg, index)
		}
	}
	newIndex, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		log.Fatal("Failed to marshal index")
	}
	os.WriteFile(filepath.Join(repoPath, common.RootDir, common.IndexFile), newIndex, 0644)
}