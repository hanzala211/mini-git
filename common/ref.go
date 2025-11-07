package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetHeadRef(repoRoot string) (string, error) {
	filePath := filepath.Join(repoRoot, RootDir, HEAD)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	headRef := strings.TrimSpace(strings.TrimPrefix(string(content), "ref:"))
	return headRef, nil
}

func GetParentSha(repoRoot string) (string, error) {
	headRef, err := GetHeadRef(repoRoot)
	if err != nil {
		return "", err
	}
	filePath := filepath.Join(repoRoot, RootDir, headRef)
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err){
			return "", nil
		}
		return "", fmt.Errorf("failed to read ref %s: %w", headRef, err)
	}
	return strings.TrimSpace(string(content)), nil
}

func UpdateHead(repoRoot string, newSha string) error {
	headRef, err := GetHeadRef(repoRoot)
	if err != nil {
		return err
	}
	filePath := filepath.Join(repoRoot, RootDir, headRef)
	if err := os.WriteFile(filePath, []byte(newSha + "\n"), 0644); err != nil {
		return fmt.Errorf("failed to update head ref %s: %w", headRef, err)
	}
	return nil
}
