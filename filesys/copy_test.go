package filesys_test

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.eloylp.dev/kit/filesys"
)

func TestRecursiveCopy(t *testing.T) {
	tmp := t.TempDir()

	filesys.Copy("test", tmp)

	filePaths := FilesInDest(t, tmp)

	assert.Equal(t, "test", RelPath(t, tmp, filePaths[0]))
	assert.Equal(t, "test/note1.txt", RelPath(t, tmp, filePaths[1]))
	assert.Equal(t, "test/note2.txt", RelPath(t, tmp, filePaths[2]))
}

func TestCopySingleFile(t *testing.T) {
	tmp := t.TempDir()

	filesys.Copy("test/note1.txt", tmp)

	filePaths := FilesInDest(t, tmp)
	assert.Equal(t, "note1.txt", RelPath(t, tmp, filePaths[0]))
}

func FilesInDest(t *testing.T, root string) (entries []string) {
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		if path == root {
			return nil
		}
		entries = append(entries, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return
}

func RelPath(t *testing.T, basepath, targetPath string) string {
	rel, err := filepath.Rel(basepath, targetPath)
	if err != nil {
		t.Fatal(err)
	}
	return rel
}
