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

	"github.com/hanzalaoc211/mini-git/common"
	"github.com/spf13/cobra"
)
type TreeEntry struct {
	Mode string
	SHA  string
}


func parseTreeToMap(treeData []byte) (map[string]TreeEntry, error) {
	entries := map[string]TreeEntry{}
	i := 0
	for i < len(treeData) {
		spaceIndex := bytes.IndexByte(treeData[i:], ' ')
		if spaceIndex == -1 {
			return nil, errors.New("invalid tree data")
		}
		mode := string(treeData[i : i + spaceIndex]) // get the mode
		fileNamStart := i + spaceIndex + 1 // +1 because we want to skip the space

		nullIndex := bytes.IndexByte(treeData[fileNamStart:], '\x00')
		if nullIndex == -1 {
			return nil, errors.New("invalid tree data")
		}
		fullNullIndex := fileNamStart + nullIndex
		fileName := string(treeData[fileNamStart : fullNullIndex]) // get the file name
		shaStart := fullNullIndex + 1
		shaEnd := shaStart + 20
		sha := treeData[shaStart : shaEnd] // get the sha
		entries[fileName] = TreeEntry{
			Mode: mode,
			SHA: hex.EncodeToString(sha),
		}
		i = shaEnd
	}
	return entries, nil
}

func restoreFile(repoRoot string, blobSha string, filePath string)  {
	blobData, err := common.ReadObject(repoRoot, blobSha)
	if err != nil {
		log.Fatalf("failed to read object in restoreFile: %v", err)
	}
	if err := os.WriteFile(filePath, blobData, 0644); err != nil {
		log.Fatalf("failed to write file: %v", err)
	}
}

func restoreFullTree(repoRoot string, treeSha string, currentPath string) {
	treeData, err := common.ReadObject(repoRoot, treeSha)
	if err != nil {
		log.Fatalf("failed to read object in restoreFullTree: %v", err)
	}
	i := 0
	for i < len(treeData) {
		spaceIndex := bytes.IndexByte(treeData[i:], ' ')
		if spaceIndex == -1 {
			log.Fatalf("invalid tree data")
		}
		mode := string(treeData[i : i + spaceIndex]) // get the mode
		fileNamStart := i + spaceIndex + 1 // +1 because we want to skip the space

		nullIndex := bytes.IndexByte(treeData[fileNamStart:], '\x00')
		if nullIndex == -1 {
			log.Fatalf("invalid tree data")
		}
		fullNullIndex := fileNamStart + nullIndex
		fileName := string(treeData[fileNamStart : fullNullIndex]) // get the file name
		shaStart := fullNullIndex + 1
		shaEnd := shaStart + 20
		sha := treeData[shaStart : shaEnd] // get the sha
		i = shaEnd
		if mode == "040000" {
			if err := os.MkdirAll(filepath.Join(currentPath, fileName), 0755); err != nil {
				log.Fatalf("failed to create directory: %v", err)
			}
			restoreFullTree(repoRoot, hex.EncodeToString(sha), filepath.Join(currentPath, fileName))
		}else {
			restoreFile(repoRoot, hex.EncodeToString(sha), filepath.Join(currentPath, fileName))
		}
	}
}

func diffAndApply(repoRoot string, newTreeSha string, oldTreeSha string) {
	var oldEntries map[string]TreeEntry
	if oldTreeSha != "" {
		oldTreeData, err := common.ReadObject(repoRoot, oldTreeSha)
		if err != nil {
			log.Fatalf("failed to read object in diffAndApply: %v", err)
		}
		oldEntries, err = parseTreeToMap(oldTreeData)
		if err != nil {
			log.Fatalf("failed to parse tree: %v", err)
		}
	}else {
		oldEntries = make(map[string]TreeEntry)
	}

	newTreeData, err := common.ReadObject(repoRoot, newTreeSha)
	if err != nil {
		log.Fatalf("failed to read object in diffAndApply second: %v", err)
	}
	newEntries, err := parseTreeToMap(newTreeData)
	if err != nil {
		log.Fatalf("failed to parse tree: %v", err)
	}

	for filename, _ := range oldEntries {
		if _, exists := newEntries[filename]; !exists {
			fullPath := filepath.Join(repoRoot, filename)
			if err := os.RemoveAll(fullPath); err != nil {
				log.Fatalf("failed to remove file: %v", err)
			}
		}
	}

	for filename, newEntry := range newEntries {
		fullPath := filepath.Join(repoRoot, filename)
		if oldEntry, exists := oldEntries[filename]; exists {
			if oldEntry.SHA == newEntry.SHA { // not modified ignore it
				continue
			}
			// SHAS are different modify the file
			if err := os.RemoveAll(fullPath); err != nil {
				log.Fatalf("failed to remove file: %v", err)
			}
			if newEntry.Mode == "040000" {
				restoreFullTree(repoRoot, newEntry.SHA, fullPath)
			}else {
				restoreFile(repoRoot, newEntry.SHA, fullPath)
			}
		}else {
			if newEntry.Mode == "040000" {
				if err := os.MkdirAll(fullPath, 0755); err != nil {
					log.Fatalf("failed to create directory: %v", err)
				}
				restoreFullTree(repoRoot, newEntry.SHA, fullPath)
			}else {
				restoreFile(repoRoot, newEntry.SHA, fullPath)
			}
		}
	}
	
	fmt.Println(oldEntries, newEntries)
}

func switchBranch(repoRoot string, newBranchPath string, currentBranch string) {
	content, err := os.ReadFile(newBranchPath)
	if err != nil {
		log.Fatalf("failed to read branch: %v", err)
	}	
	contentStr := string(content)
	contentStr = strings.TrimSpace(contentStr)
	currentBranchContent, err := os.ReadFile(filepath.Join(repoRoot, common.RootDir, common.RefsDir, common.HeadDir, currentBranch))
	currentBranchContentStr := strings.TrimSpace(string(currentBranchContent))
	if currentBranchContentStr == contentStr {
		return
	}
	if contentStr == "" {
		if err != nil {
			log.Fatalf("failed to read current branch: %v", err)
		}
		os.WriteFile(newBranchPath, currentBranchContent, 0644)
		return
	}
	newCommitData, err := common.ReadObject(repoRoot, contentStr)
	if err != nil {
		log.Fatalf("failed to read object in switchBranch: %v", err)
	}
	newCommitDataStr := string(newCommitData)
	newCommitDataStr = strings.TrimSpace(newCommitDataStr)
	if newCommitDataStr == "" {
		log.Fatalf("branch %s is not a valid branch", newBranchPath)
	}
	newTreeSha := strings.Split(newCommitDataStr, "\n")[0]
	newTreeSha = strings.Split(newTreeSha, " ")[1]
	newTreeSha = strings.TrimSpace(newTreeSha)
	oldCommitData, err := common.ReadObject(repoRoot, currentBranchContentStr)
	if err != nil {
		log.Fatalf("failed to read object in switchBranch second: %v", err)
	}
	oldCommitDataStr := string(oldCommitData)
	oldCommitDataStr = strings.TrimSpace(oldCommitDataStr)
	if oldCommitDataStr == "" {
		log.Fatalf("branch %s is not a valid branch", newBranchPath)
	}
	oldTreeSha := strings.Split(oldCommitDataStr, "\n")[0]
	oldTreeSha = strings.Split(oldTreeSha, " ")[1]
	oldTreeSha = strings.TrimSpace(oldTreeSha)
	diffAndApply(repoRoot, newTreeSha, oldTreeSha)
}

func buildIndexFromTree(repoRoot string, treeSha string, prefix string) (common.Index, error) {
	index := make(common.Index)
	treeData, err := common.ReadObject(repoRoot, treeSha)
	if err != nil {
		return nil, fmt.Errorf("failed to read tree object: %w", err)
	}
	
	i := 0
	for i < len(treeData) {
		spaceIndex := bytes.IndexByte(treeData[i:], ' ') // finding space between mode and file name
		if spaceIndex == -1 {
			break
		}
		mode := string(treeData[i : i + spaceIndex]) // getting mode
		fileNamStart := i + spaceIndex + 1
		
		nullIndex := bytes.IndexByte(treeData[fileNamStart:], '\x00')
		if nullIndex == -1 {
			break
		}
		fullNullIndex := fileNamStart + nullIndex
		fileName := string(treeData[fileNamStart : fullNullIndex])
		shaStart := fullNullIndex + 1
		shaEnd := shaStart + 20
		sha := treeData[shaStart : shaEnd]
		shaHex := hex.EncodeToString(sha)
		
		var filePath string
		if prefix == "" {
			filePath = fileName
		} else {
			filePath = filepath.ToSlash(filepath.Join(prefix, fileName))
		}
		
		if mode == "040000" {
			// It's a directory, recurse into it
			subIndex, err := buildIndexFromTree(repoRoot, shaHex, filePath)
			if err != nil {
				return nil, err
			}
			// Merge the sub-index into the main index
			for k, v := range subIndex {
				index[k] = v
			}
		} else {
			// It's a file, add it to the index
			index[filePath] = shaHex
		}
		
		i = shaEnd
	}
	
	return index, nil
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
	
	branchContent, err := os.ReadFile(branchPath)
	if err != nil {
		log.Fatalf("failed to read branch: %v", err)
	}
	branchCommitSha := strings.TrimSpace(string(branchContent))
	if branchCommitSha != "" {
		commitData, err := common.ReadObject(repoPath, branchCommitSha)
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
	
	err = os.WriteFile(filepath.Join(repoPath, common.RootDir, common.HEAD), []byte("ref: refs/heads/" + branchName), 0644)
	if err != nil {
		log.Fatalf("failed to update head: %v", err)
	}
	fmt.Printf("Switched to branch %s\n", branchName)
}