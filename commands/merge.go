package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hanzalaoc211/mini-git/common"
	"github.com/spf13/cobra"
)

func MergeCommand(cmd *cobra.Command, args []string) {
	repoPath, err := common.FindRepoRoot()
	if err != nil {
		log.Fatal(err)
	}
	currentBranch, err  := common.GetHeadRef(repoPath)
	if err != nil {
		log.Fatal(err)
	}
	currentBranch = strings.Split(currentBranch, "/")[2]
	newBranch := args[0]
	if currentBranch == newBranch {
		fmt.Println("Already on the branch you are trying to merge")
		return
	}
	newBranchCommit, err := os.ReadFile(filepath.Join(repoPath, common.RootDir, common.RefsDir, common.HeadDir, newBranch))
	if err != nil {
		log.Fatal(err)
	}
	newBranchCommitSHA := strings.TrimSpace(string(newBranchCommit))
	oldBranchCommit, err := common.GetParentSha(repoPath)
	if err != nil {
		log.Fatal(err)
	}
	if isAncestor(repoPath, oldBranchCommit, newBranchCommitSHA) {
		if err := os.WriteFile(filepath.Join(repoPath, common.RootDir, common.RefsDir, common.HeadDir, currentBranch), newBranchCommit, 0644); err != nil {
			log.Fatal(err)
		}
		newTreeShaByt, _ := common.ReadObject(repoPath, newBranchCommitSHA)
		newTreeSha := strings.Split(strings.TrimSpace(string(newTreeShaByt)), "\n")[0]
		newTreeSha = strings.Split(newTreeSha, " ")[1]
		newTreeSha = strings.TrimSpace(newTreeSha)
		oldTreeShaByt, _ := common.ReadObject(repoPath, oldBranchCommit)
		oldTreeSha := strings.Split(strings.TrimSpace(string(oldTreeShaByt)), "\n")[0]
		oldTreeSha = strings.Split(oldTreeSha, " ")[1]
		oldTreeSha = strings.TrimSpace(oldTreeSha)
		diffAndApply(repoPath, newTreeSha, oldTreeSha)
		if newBranchCommitSHA != "" {
			commitData, err := common.ReadObject(repoPath, newBranchCommitSHA)
			if err != nil {
				log.Fatalf("failed to read commit object: %v", err)
			}
			commitStr := strings.TrimSpace(string(commitData))
			treeSha := strings.Split(commitStr, "\n")[0]
			treeSha = strings.Split(treeSha, " ")[1]
			treeSha = strings.TrimSpace(treeSha)
			
			newIndex, err := buildIndexFromTree(repoPath, treeSha, "")
			if err != nil {
				log.Fatalf("failed to build index from tree: %v", err)
			}
			
			indexBytes, err := json.MarshalIndent(newIndex, "", "  ")
			if err != nil {
				log.Fatalf("failed to marshal index: %v", err)
			}
			indexPath := filepath.Join(repoPath, common.RootDir, common.IndexFile)
			if err := os.WriteFile(indexPath, indexBytes, 0644); err != nil {
				log.Fatalf("failed to write index: %v", err)
			}
		} else {
			// Empty branch - clear the index
			emptyIndex := make(common.Index)
			indexBytes, err := json.MarshalIndent(emptyIndex, "", "  ")
			if err != nil {
				log.Fatalf("failed to marshal index: %v", err)
			}
			indexPath := filepath.Join(repoPath, common.RootDir, common.IndexFile)
			if err := os.WriteFile(indexPath, indexBytes, 0644); err != nil {
				log.Fatalf("failed to write index: %v", err)
			}
		}
		return
	}
	log.Fatal("3-way merges are not yet implemented")
}

func isAncestor(repoRoot string, possibleAncestorCommit string, commit string) bool {
	for commit != "" {
		if possibleAncestorCommit == commit {
			return true
		}
		commitObj, err := common.ReadObject(repoRoot, commit)
		if err != nil {
			log.Fatal(err)
		}
		commitObjStr := strings.TrimSpace(string(commitObj))
		lines := strings.Split(commitObjStr, "\n")
		if len(lines) < 2 {
			// No parent commit (initial commit or orphan commit)
			return false
		}
		parentLine := strings.Split(lines[1], " ")
		if len(parentLine) < 2 {
			// Invalid parent line format
			return false
		}
		commit = parentLine[1]
		commit = strings.TrimSpace(commit)
	}
	return false
}