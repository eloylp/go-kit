package filesys

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Copy will copy directories and files recursively
// from sourceRoot to destRoot.
// It supports both, relative and absolute paths. It
// also uses streams for reducing memory consumption.
//
// This function will abort the test if any operation fails.
// In such case, it will not do any type of clean up operation.
//
// As an example, given the following filesystem tree:
// /home/user/data/note1.txt
// /home/user/data/note2.txt
// /home/user2
//
// Executing Copy with the following parameters:
// sourceRoot: /home/user/data
// destRoot:   /home/user2
//
// The copied data will be:
// /home/user2/data/note1.txt
// /home/user2/data/note2.txt
func Copy(sourceRoot, destRoot string) error {

	// Ensure we work with absolute paths from here.
	sourceRoot, err := filepath.Abs(sourceRoot)
	if err != nil {
		return err
	}
	destRoot, err = filepath.Abs(destRoot)
	if err != nil {
		return err
	}

	err = filepath.WalkDir(sourceRoot, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		destFirstElement := filepath.Base(sourceRoot)
		var destPath string
		if sourceRoot == path {
			// The current walked path is the same as source root.
			// So we just join with the root destination element.
			destPath = filepath.Join(destRoot, destFirstElement)
		} else {
			rel, err := filepath.Rel(sourceRoot, path)
			if err != nil {
				return err
			}
			destPath = filepath.Join(destRoot, destFirstElement, rel)
		}

		if info.IsDir() {
			err = os.MkdirAll(destPath, 0775)
			if err != nil {
				return err
			}
			return nil
		}
		fileFrom, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fileFrom.Close()
		fileTo, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer fileTo.Close()
		if _, err := io.Copy(fileTo, fileFrom); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
