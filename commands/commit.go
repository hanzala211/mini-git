package commands

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hanzalaoc211/mini-git/common"
	"github.com/spf13/cobra"
)

func buildTree(repoPath string, index common.Index) (string, error) {
	fileTree := make(map[string]interface{})

	for fullPath, sha := range index {
		paths := strings.Split(fullPath, "/")
		currentFileTree := fileTree
		for _, path := range paths {
			if paths[len(paths) - 1] == path {
				currentFileTree[path] = sha
			}else {
				if _, ok := currentFileTree[path]; !ok {
					currentFileTree[path] = make(map[string]interface{})
				}
				currentFileTree = currentFileTree[path].(map[string]interface{})
			}
		}
	}

	return writeTreeRecursively(repoPath, fileTree)
}

func writeTreeRecursively(repoPath string, treeRoot map[string]interface{}) (string, error) {
	var treeArr []common.TreeNode

	for name, node := range treeRoot {
		if sha, ok := node.(string); ok {
			treeArr = append(treeArr, common.TreeNode{
				Name: name,
				Mode: "100644", // file mode
				Sha: sha,
			})
		}else if childNode, ok := node.(map[string]interface{}); ok {
			childNodeSha, err := writeTreeRecursively(repoPath, childNode)
			if err != nil {
				return "", err
			}
			treeArr = append(treeArr, common.TreeNode{
				Name: name,
				Sha: childNodeSha,
				Mode: "040000", // folder mode
			})
		}
	}

	sort.Slice(treeArr, func(i, j int) bool {
		return treeArr[i].Name < treeArr[j].Name
	})

	var treeContent bytes.Buffer

	for _, tree := range treeArr {
		hexBytes, _ := hex.DecodeString(tree.Sha)
		treeContent.Write([]byte(fmt.Sprintf("%s %s\x00", tree.Mode, tree.Name)))
		treeContent.Write(hexBytes)
	}

	return common.WriteObject(repoPath, treeContent.Bytes(), common.TreeFile, "")
}

func findLastCommitTreeSha(repoPath string) (string, error) {
	parentSha, err := common.GetParentSha(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to get parent commit: %w", err)
	}
	if parentSha == "" {
		return "", nil
	}
	commitObjByt, err := common.ReadObject(repoPath, parentSha)
	if err != nil {
		return "", nil
	}
	commitStr := string(commitObjByt)
	treeSha := strings.Split(commitStr, "\n")[0]
	if treeSha == "" {
		return "", nil
	}
	return strings.Split(treeSha, " ")[1], nil
}

func CommitCommand(cmd *cobra.Command, args []string) {
	repoPath, err := common.FindRepoRoot()
	if err != nil {
		log.Fatal(err)
	}
	commitMsg, _ := cmd.Flags().GetString("m")
	if commitMsg == "" {
		log.Fatal("commit message is required")
	}

	indexBytes, err := os.ReadFile(filepath.Join(repoPath, common.RootDir, common.IndexFile))
	if err != nil {
		log.Fatal("Failed to read index")
	}
	var index common.Index
	json.Unmarshal(indexBytes, &index)
	if index == nil  {
		log.Fatal("no changes to commit")
	}
	treeSha, err := buildTree(repoPath, index)
	if err != nil {
		log.Fatalf("failed to build trees: %w", err)
	}
	lastCommitTreeSha, err := findLastCommitTreeSha(repoPath)
	if lastCommitTreeSha == treeSha {
		log.Fatal("no file to push to commit")
	}
	if err != nil {
		log.Fatalf("failed to find last commit tree: %w", err)
	}
	
	parentSha, err := common.GetParentSha(repoPath)
	if err !=nil {
		log.Fatalf("failed to get parent commit: %w", err)
	}

	now := time.Now()
	timestamp := fmt.Sprintf("%d %s", now.Unix(), now.Format("-0700"))

	var commitContent bytes.Buffer

	fmt.Fprintf(&commitContent, "tree %s\n", treeSha)
	if parentSha != "" {
		fmt.Fprintf(&commitContent, "parent %s\n", parentSha) 
	}
	fmt.Fprintf(&commitContent, "\n%s\n", commitMsg)
	fmt.Fprintf(&commitContent, "%s\n", timestamp)
	commitSha, err := common.WriteObject(repoPath, commitContent.Bytes(), common.CommitFile, "")
	if err != nil {
		log.Fatalf("failed to write commit object: %w", err)
	}
	err = common.UpdateHead(repoPath, commitSha)
	if err != nil {
		log.Fatalf("failed to update head: %w", err)
	}
	fmt.Println("changes committed")
}