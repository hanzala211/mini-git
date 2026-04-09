package commands

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hanzala211/mini-git/common"
	"github.com/spf13/cobra"
)

func getAllAncestorCommits(repoPath string, branchName string) (map[string]bool, error) {
	parentSha, err := common.GetBranchCommitSha(repoPath, branchName) // get current branch's commit sha
	if err != nil {
		return map[string]bool{}, fmt.Errorf("failed to get branch %s commit: %w", branchName, err)
	}
	if parentSha == "" {
		return map[string]bool{}, nil
	}
	ancestorCommits := make(map[string]bool)
	hasAncestor := true
	ancestorCommits[parentSha] = true
	for hasAncestor {
		commitObjByt, err := common.ReadObject(repoPath, parentSha)
		if err != nil {
			break
		}
		commitStr := string(commitObjByt)
		lines := strings.Split(commitStr, "\n")
		if len(lines) < 2 || !strings.HasPrefix(lines[1], "parent") {
			hasAncestor = false
			break
		}
		ancestorCommitSha := strings.Split(lines[1], " ")[1]
		ancestorCommitSha = strings.TrimSpace(ancestorCommitSha)
		parentSha = ancestorCommitSha
		ancestorCommits[ancestorCommitSha] = true
	}
	return ancestorCommits, nil
}

func findMergeBase(repoPath string, currentBranch string, newBranch string) (string, error) {
	currentBranchCommits, err := getAllAncestorCommits(repoPath, currentBranch)
	if err != nil {
		return "", fmt.Errorf("failed to get all ancestor commits: %v", err)
	}
	targetCommitSha, err := common.GetBranchCommitSha(repoPath, newBranch)
	if err != nil {
		return "", fmt.Errorf("failed to get branch %s commit: %v", newBranch, err)
	}
	for targetCommitSha != "" {
		if currentBranchCommits[targetCommitSha] {
			return targetCommitSha, nil
		}
		commitData, err := common.ReadObject(repoPath, targetCommitSha)
		if err != nil {
			return "", fmt.Errorf("failed to read commit object: %v", err)
		}
		commitStr := strings.TrimSpace(string(commitData))
		lines := strings.Split(commitStr, "\n")
		if len(lines) < 2 || !strings.HasPrefix(lines[1], "parent") {
			return "", fmt.Errorf("invalid commit object: %v", err)
		}
		targetCommitSha = strings.Split(lines[1], " ")[1]
	}
	return "", fmt.Errorf("failed to find merge base")
}

func buildTreeFromCommit(repoPath string, commitSha string) (map[string]string, error) {
	commitData, err := common.ReadObject(repoPath, commitSha)
	if err != nil {
		return nil, fmt.Errorf("failed to read commit object: %v", err)
	}
	commitStr := strings.TrimSpace(string(commitData))
	lines := strings.Split(commitStr, "\n")
	treeSha := strings.Split(lines[0], " ")[1]
	tree := make(map[string]string)
	err = flattenTree(repoPath, treeSha, "", tree)
	if err != nil {
		return nil, fmt.Errorf("failed to flatten tree: %v", err)
	}
	return tree, nil
}

func flattenTree(repoPath string, treeSha string, prefix string, tree map[string]string) error {
	treeData, err := common.ReadObject(repoPath, treeSha)
	if err != nil {
		return fmt.Errorf("failed to read tree object: %v", err)
	}
	i := 0
	for i < len(treeData) {
		spaceIndex := bytes.IndexByte(treeData[i:], ' ')
		if spaceIndex == -1 {
			return errors.New("invalid tree data")
		}
		mode := string(treeData[i : i+spaceIndex]) // get the mode
		fileNamStart := i + spaceIndex + 1         // +1 because we want to skip the space

		nullIndex := bytes.IndexByte(treeData[fileNamStart:], '\x00')
		if nullIndex == -1 {
			return errors.New("invalid tree data")
		}
		fullNullIndex := fileNamStart + nullIndex
		fileName := string(treeData[fileNamStart:fullNullIndex]) // get the file name
		shaStart := fullNullIndex + 1
		shaEnd := shaStart + 20
		sha := treeData[shaStart:shaEnd] // get the sha
		if mode == "040000" {
			err = flattenTree(repoPath, hex.EncodeToString(sha), filepath.Join(prefix, fileName), tree)
			if err != nil {
				return err
			}
		} else {
			tree[filepath.Join(prefix, fileName)] = hex.EncodeToString(sha)
		}
		i = shaEnd
	}
	return nil
}

func mergeFiles(repoRoot string, blobSha string, filePath string) {
	blobData, err := common.ReadObject(repoRoot, blobSha)
	if err != nil {
		log.Fatalf("failed to read object in restoreFile: %v", err)
	}
	fullPath := filepath.Join(repoRoot, filePath)
	parentPath := filepath.Dir(fullPath)
	if err := os.MkdirAll(parentPath, 0755); err != nil {
		log.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(filePath, blobData, 0644); err != nil {
		log.Fatalf("failed to write file: %v", err)
	}
}

func MergeCommand(cmd *cobra.Command, args []string) {
	repoPath, err := common.FindRepoRoot()
	if err != nil {
		log.Fatal(err)
	}
	currentBranch, err := common.GetHeadRef(repoPath)
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
	} else {
		// 3-way merge
		mergeBase, err := findMergeBase(repoPath, currentBranch, newBranch)
		if err != nil {
			log.Fatal(err)
		}
		mergeBaseTree, err := buildTreeFromCommit(repoPath, mergeBase)
		if err != nil {
			log.Fatal(err)
		}
		targetTree, err := buildTreeFromCommit(repoPath, newBranchCommitSHA)
		if err != nil {
			log.Fatal(err)
		}
		headTree, err := buildTreeFromCommit(repoPath, oldBranchCommit)
		if err != nil {
			log.Fatal(err)
		}
		masterTree := make(map[string]bool)
		for path, _ := range mergeBaseTree {
			masterTree[path] = true
		}
		for path, _ := range targetTree {
			masterTree[path] = true
		}
		for path, _ := range headTree {
			masterTree[path] = true
		}
		resolvedFiles := make(map[string]string)
		conflicts := []string{}
		for path, _ := range masterTree {
			baseSha := mergeBaseTree[path]
			targetSha := targetTree[path]
			headSha := headTree[path]
			if headSha == targetSha {
				if headSha != "" {
					resolvedFiles[path] = headSha
				}
			} else if headSha == baseSha {
				if targetSha != "" {
					resolvedFiles[path] = targetSha
				}
			} else if baseSha == targetSha {
				if headSha != "" {
					resolvedFiles[path] = headSha
				}
			} else {
				conflicts = append(conflicts, path)
			}
		}
		if len(conflicts) > 0 {
			fmt.Println("Automatic merge failed; fix conflicts and then commit the result.")
			for _, file := range conflicts {
				fmt.Println("CONFLICT: ", file)
			}
			return
		}
		for path, _ := range headTree {
			if _, exists := resolvedFiles[path]; !exists {
				if err := os.Remove(filepath.Join(repoPath, path)); err != nil {
					log.Fatalf("failed to remove file: %v", err)
				}
			}
		}

		for path, _ := range resolvedFiles {
			mergeFiles(repoPath, resolvedFiles[path], path)
		}
		var index common.Index = resolvedFiles
		indexBytes, err := json.MarshalIndent(index, "", "  ")
		if err != nil {
			log.Fatalf("failed to marshal index: %v", err)
		}
		if err := os.WriteFile(filepath.Join(repoPath, common.RootDir, common.IndexFile), indexBytes, 0644); err != nil {
			log.Fatalf("failed to write index: %v", err)
		}

		now := time.Now()
		timestamp := fmt.Sprintf("%d %s", now.Unix(), now.Format("-0700"))

		var commitContent bytes.Buffer

		newTreeSha, err := buildTree(repoPath, index)
		if err != nil {
			log.Fatalf("failed to build tree: %v", err)
		}

		fmt.Fprintf(&commitContent, "tree %s\n", newTreeSha)
		if oldBranchCommit != "" {
			fmt.Fprintf(&commitContent, "parent %s\n", oldBranchCommit)
		}
		fmt.Fprintf(&commitContent, "parent %s\n", newBranchCommitSHA)
		fmt.Fprintf(&commitContent, "\n%s\n", "Merge branch '"+newBranch+"' into "+currentBranch)
		fmt.Fprintf(&commitContent, "%s\n", timestamp)
		commitSha, err := common.WriteObject(repoPath, commitContent.Bytes(), common.CommitFile, "")
		if err != nil {
			log.Fatalf("failed to write commit object: %w", err)
		}
		err = common.UpdateHead(repoPath, commitSha)
		if err != nil {
			log.Fatalf("failed to update head: %w", err)
		}
		fmt.Println("Automatic merge successful")
	}
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
