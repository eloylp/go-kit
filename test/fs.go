package test

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// Copy will copy directories and files recursively from one path to another.
// This function will abort the test if any operation files. This function will
// not do any type on clean up on fail. This is recommended to use in conjunction
// to testing.TempDir() .
func Copy(t *testing.T, source, dest string) {
	t.Helper()
	err := filepath.Walk(source, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		sourceRel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			err = os.MkdirAll(filepath.Join(dest, sourceRel), 0775)
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
		fileTo, err := os.Create(filepath.Join(dest, sourceRel))
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
		t.Fatal(err)
	}
}
