package main

import (
	"fmt"
	"os"
)

func createFolderIfNotExists(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to create folder: %v", err)
		}
	}

	return nil
}
