package utils

import (
	// "fmt"
	"io"
	"os"
	"path/filepath"
)


func CopyFiles(sourceDir, targetDir string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, relPath)

		if !fileExists(targetPath) {
			if info.IsDir() {
				return os.MkdirAll(targetPath, info.Mode())
			} else {
				if err := copyFile(path, targetPath); err != nil {
					return err
				}
			}
		}
		

		return nil
	})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	if err := destFile.Sync(); err != nil {
		return err
	}

	return nil
}