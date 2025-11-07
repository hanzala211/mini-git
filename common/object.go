package common

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func WriteObject(repoRoot string,content []byte, fileType string, filePath string) (string, error) {
	header := fmt.Sprintf("%s %d\x00", fileType, len(content))
	fullData := append([]byte(header), content...)

	hash := sha1.Sum(fullData)
	stringHash := fmt.Sprintf("%x", hash)
	objFolder := filepath.Join(repoRoot , RootDir , ObjectDir , stringHash[:2])
	objFile := filepath.Join(objFolder , stringHash[2:])
	if _, err := os.Stat(objFile); err == nil { // if same object already exists dont add it 
		return stringHash, nil
	}
	if err := os.MkdirAll(objFolder, 0744); err != nil {
		return "", err
	}
	var bytes bytes.Buffer
	writer := zlib.NewWriter(&bytes)
	if _, err := writer.Write(fullData); err != nil {
		return "", fmt.Errorf("failed to compress object data: %w", err) 
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close zlib writer: %w", err)
	}
	if err := os.WriteFile(objFile, bytes.Bytes(), 0644); err != nil {
		return "", err
	}
	if fileType == BlobFile && filePath != "" {
		fmt.Print(fmt.Sprintln("adding file", filePath))
	}
	return stringHash, nil
}

func ReadObject(repoRoot string, objectSha string) ([]byte, error) {
	folderDir := filepath.Join(repoRoot, RootDir, ObjectDir, objectSha[:2])
	fileDir := filepath.Join(folderDir, objectSha[2:])
	content, err := os.ReadFile(fileDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}
	zr, err := zlib.NewReader(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to start reader: %w", err)
	}
	decompressedData, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("failed to read decompressed data: %w", err)
	}
	defer zr.Close()
	// Strip the header (format: "commit <size>\x00" or "tree <size>\x00" etc.)
	nullIndex := bytes.IndexByte(decompressedData, '\x00')
	if nullIndex == -1 {
		return nil, fmt.Errorf("invalid object format: missing null byte")
	}
	return decompressedData[nullIndex+1:], nil // +1 because we want to skip the \x00
}