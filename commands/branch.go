package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hanzala211/mini-git/common"
	"github.com/spf13/cobra"
)

func listBranches(repoRoot string) {
	filePath := filepath.Join(repoRoot, common.RootDir, common.RefsDir, common.HeadDir)
	headRef, _ := common.GetHeadRef(repoRoot)
	entries, err := os.ReadDir(filePath)
	if err != nil {
		log.Fatal(err)
	}
	currentBranch := strings.Split(headRef, "/")[2]
	for _, entry := range entries {
		if entry.Name() == currentBranch {
			fmt.Println("*", entry.Name())
		} else {
			fmt.Println(entry.Name())
		}
	}
}

func createBranch(repoRoot string, branchName string) {
	filePath := filepath.Join(repoRoot, common.RootDir, common.RefsDir, common.HeadDir, branchName)
	if _, err := os.Stat(filePath); err == nil {
		log.Fatalf("branch %s already exists", branchName)
	}
	parentSha, err := common.GetParentSha(repoRoot)
	if err != nil {
		log.Fatalf("failed to get parent commit: %v", err)
	}
	if err := os.WriteFile(filePath, []byte(parentSha), 0644); err != nil {
		log.Fatalf("failed to create branch %s: %v", branchName, err)
	}
	headPath := filepath.Join(repoRoot, common.RootDir, common.HEAD)
	err = os.WriteFile(headPath, []byte("ref: refs/heads/"+branchName), 0644)
	if err != nil {
		log.Fatalf("failed to update head: %v", err)
	}
	fmt.Printf("Branch %s created and switched to it.\n", branchName)
}

func BranchCommand(cmd *cobra.Command, args []string) {
	repoPath, err := common.FindRepoRoot()
	if err != nil {
		log.Fatal(err)
	}
	if len(args) == 0 {
		listBranches(repoPath)
	} else {
		createBranch(repoPath, args[0])
	}
}
