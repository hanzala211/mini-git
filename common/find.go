package common

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

func FindRepoRoot() (string, error) {
	cwd, err := os.Getwd() 
	if err != nil {
		log.Fatal(err)
	}

	for {
		repoPath := filepath.Join(cwd, RootDir)
		stat, err := os.Stat(repoPath)
		if err == nil && stat.IsDir() { // checks if the cwd has .minigit folder IsDir returns true if found
			return cwd, nil
		}

		if cwd == filepath.Dir(cwd){ // to check if we had reached the top level like /home == / false next time this will be / == / so that will tell us that we are at system path
			return "", errors.New("not a mini-git repository")
		}

		cwd = filepath.Dir(cwd)
	}
}