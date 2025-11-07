package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hanzalaoc211/mini-git/common"
	"github.com/spf13/cobra"
)

func InitCommand(cmd *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cwd, common.RootDir)); err == nil {
		log.Fatal("already a mini-git repository")
	}
	os.MkdirAll(filepath.Join(cwd, common.RootDir), 0755)
	os.MkdirAll(filepath.Join(cwd, common.RootDir, common.ObjectDir), 0755)
	os.MkdirAll(filepath.Join(cwd, common.RootDir, common.RefsDir), 0755)
	os.MkdirAll(filepath.Join(cwd, common.RootDir, common.RefsDir, common.HeadDir), 0755)
	os.WriteFile(filepath.Join(cwd, common.RootDir, common.RefsDir, common.HeadDir, "master"), []byte(""), 0644)
	os.WriteFile(filepath.Join(cwd, common.RootDir, common.IndexFile), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(cwd, common.RootDir, common.HEAD), []byte("ref: refs/heads/master\n"), 0644)
	fmt.Println("Initialized mini-git repository in", cwd)
}