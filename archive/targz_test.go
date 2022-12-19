//nolint:gosec
package archive_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eloylp.dev/kit/archive"
)

func TestCreateTARGZ(t *testing.T) {
	tmpDir := t.TempDir()
	tarGzFilePath := fmt.Sprintf("%s/test.tar.gz", tmpDir)

	wBytes, err := archive.TARGZ(tarGzFilePath, Root+"/gnu.png", Root+"/tux.png", Root+"/notes")
	require.NoError(t, err)
	assert.Equal(t, RootSize, wBytes)

	tarGzFile, err := os.Open(tarGzFilePath)
	require.NoError(t, err)
	defer tarGzFile.Close()
	AssertTARGZMD5Sums(t, tarGzFile, map[string]string{
		".":                        "",
		"gnu.png":                  GnuTestFileMD5,
		"tux.png":                  TuxTestFileMD5,
		"notes.txt":          NotesTestFileMD5,
		"subnotes":           "",
		"subnotes/notes.txt": SubNotesTestFileMD5,
	})
}

func TestExtractTARGZ(t *testing.T) {
	tmpDir := t.TempDir()

	wBytes, err := archive.ExtractTARGZ(tmpDir, RootTARGZ)
	require.NoError(t, err)
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
	require.NoError(t, err)
}

func TestExtractTAGHeaderPathEscalationIsForbidden(t *testing.T) {
	rootDir := t.TempDir()
	targetDir := filepath.Join(rootDir, "sub")

	// Prepare a tar.gz test fixture, that will include a header name trying to scale
	// to other dirs.
	buff := bytes.NewBuffer(nil)
	gw := gzip.NewWriter(buff)
	tw := tar.NewWriter(gw)
	fileHeaderName := "../scalated-to-root"
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     fileHeaderName,
		Size:     70,
	}))
	_, err := tw.Write([]byte("Hello, im the content of a file that will be placed in the wrong place"))
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())

	_, err = archive.ExtractTARGZStream(buff, targetDir)
	expected := fmt.Sprintf("at ExtractTARGZ(): the path you provided %s is not a suitable one",
		filepath.Join(rootDir, "scalated-to-root"))
	assert.EqualError(t, err, expected)
}

func TestExtractTARGZDoesNotAcceptRelativePaths(t *testing.T) {
	buffer := bytes.NewReader(nil)
	_, err := archive.ExtractTARGZStream(buffer, "relative/one")
	assert.EqualError(t, err, "at ExtractTARGZ(): the extraction path must be absolute")
}
