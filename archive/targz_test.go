//nolint:gosec
package archive_test

import (
	"crypto/md5"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.eloylp.dev/kit/archive"
)

func TestCreateTARGZFromDir(t *testing.T) {
	tmpDir := t.TempDir()
	tarGzFilePath := fmt.Sprintf("%s/test.tar.gz", tmpDir)
	tarGzFile, err := os.Create(tarGzFilePath)
	assert.NoError(t, err)
	wBytes, err := archive.CreateTARGZ(tarGzFile, Root)
	assert.NoError(t, err)
	err = tarGzFile.Close()
	assert.NoError(t, err)
	assert.Equal(t, RootSize, wBytes)
	tarGzFile, err = os.Open(tarGzFilePath)
	assert.NoError(t, err)
	defer tarGzFile.Close()
	AssertTARGZMD5Sums(t, tarGzFile, map[string]string{
		".":                        "",
		"gnu.png":                  GnuTestFileMD5,
		"notes":                    "",
		"notes/notes.txt":          NotesTestFileMD5,
		"notes/subnotes":           "",
		"notes/subnotes/notes.txt": SubNotesTestFileMD5,
		"tux.png":                  TuxTestFileMD5,
	})
}

func TestCreateTARGZUniqueFile(t *testing.T) {
	tmpDir := t.TempDir()
	tarGzFilePath := fmt.Sprintf("%s/test.tar.gz", tmpDir)
	tarGzFile, err := os.Create(tarGzFilePath)
	assert.NoError(t, err)
	_, err = archive.CreateTARGZ(tarGzFile, fmt.Sprintf("%s/notes/subnotes/notes.txt", Root))
	assert.NoError(t, err)
	err = tarGzFile.Close()
	assert.NoError(t, err)
	tarGzFile, err = os.Open(tarGzFilePath)
	assert.NoError(t, err)
	defer tarGzFile.Close()
	AssertTARGZMD5Sums(t, tarGzFile, map[string]string{
		"notes.txt": SubNotesTestFileMD5,
	})
}

func TestExtractTARGZ(t *testing.T) {
	tmpDir := t.TempDir()
	tarGz, err := os.Open(RootTARGZ)
	assert.NoError(t, err)
	wBytes, err := archive.ExtractTARGZ(tarGz, tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, RootSize, wBytes)

	assertMap := map[string]string{
		"gnu.png":                  GnuTestFileMD5,
		"notes":                    "",
		"notes/notes.txt":          NotesTestFileMD5,
		"notes/subnotes":           "",
		"notes/subnotes/notes.txt": SubNotesTestFileMD5,
		"tux.png":                  TuxTestFileMD5,
	}
	err = filepath.Walk(tmpDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == tmpDir {
			return nil
		}
		relPath, err := filepath.Rel(tmpDir, path)
		if err != nil {
			return err
		}
		expectedMd5, ok := assertMap[relPath]
		if !ok {
			assert.Fail(t, "is expected that %q was present in assertMap", relPath)
		}
		if expectedMd5 == "" && info.IsDir() {
			return nil
		}
		fileContent, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		md5Hash := fmt.Sprintf("%x", md5.Sum(fileContent)) //nolint:gosec
		assert.Equal(t, expectedMd5, md5Hash)
		return nil
	})
	assert.NoError(t, err)
}
