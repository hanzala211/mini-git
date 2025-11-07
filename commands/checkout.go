package commands

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hanzalaoc211/mini-git/common"
	"github.com/spf13/cobra"
)

func restoreTreeFiles(repoRoot string, treeSha string) {
	treeData, err := common.ReadObject(repoRoot, treeSha)
	if err != nil{
		log.Fatalf("failed to read tree: %v", err)
	}
	treeDataStr := string(treeData)
	fmt.Println(treeDataStr)
}

func switchBranch(repoRoot string, branchPath string, currentBranch string) {
	content, err := os.ReadFile(branchPath)
	if err != nil {
		log.Fatalf("failed to read branch: %v", err)
	}	
	contentStr := string(content)
	contentStr = strings.TrimSpace(contentStr)
	if contentStr == "" {
		currentBranchContent, err := os.ReadFile(filepath.Join(repoRoot, common.RootDir, common.RefsDir, common.HeadDir, currentBranch))
		if err != nil {
			log.Fatalf("failed to read current branch: %v", err)
		}
		os.WriteFile(branchPath, currentBranchContent, 0644)
	}
	commitData, err := common.ReadObject(repoRoot, contentStr)
	if err != nil {
		log.Fatalf("failed to read object: %v", err)
	}
	commitDataStr := string(commitData)
	commitDataStr = strings.TrimSpace(commitDataStr)
	if commitDataStr == "" {
		log.Fatalf("branch %s is not a valid branch", branchPath)
	}
	treeSha := strings.Split(commitDataStr, "\n")[0]
	treeSha = strings.Split(treeSha, " ")[1]
	treeSha = strings.TrimSpace(treeSha)
	restoreTreeFiles(repoRoot, treeSha)
}


func CheckoutCommand(cmd *cobra.Command, args []string) {
	repoPath, err := common.FindRepoRoot()
	if err != nil {
		log.Fatal(err)
	}
	headRef, _ := common.GetHeadRef(repoPath)
	currentBranch := strings.Split(headRef, "/")[2]
	branchName := args[0]
	if currentBranch == branchName {
		fmt.Printf("Already on '%s'\n", branchName)
		return
	}
	branchPath := filepath.Join(repoPath, common.RootDir, common.RefsDir, common.HeadDir, branchName)
	if _, err := os.Stat(branchPath); err != nil {
		log.Fatalf("branch %s does not exist", branchName)
	}
	switchBranch(repoPath, branchPath, currentBranch)
	err = os.WriteFile(filepath.Join(repoPath, common.RootDir, common.HEAD), []byte("ref: refs/heads/" + branchName), 0644)
	if err != nil {
		log.Fatalf("failed to update head: %v", err)
	}
	fmt.Printf("Switched to branch %s\n", branchName)
}