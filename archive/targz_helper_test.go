package archive_test

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	Root                = "tests/root"
	RootSize            = int64(533766)
	TuxTestFileMD5      = "a0e6e27f7e31fd0bd549ea936033bf28"
	GnuTestFileMD5      = "0073978283cb69d470ec2ea1b66f1988"
	NotesTestFileMD5    = "36d7e788e7a54109f5beb9ebe103da39"
	SubNotesTestFileMD5 = "0ff6da62cf7875cce432f7b955008953"
	RootTARGZ           = "tests/root.tar.gz"
)

// AssertMD5Sums accepts a reader, which streams the tar.gz
// content and an expected map of file names/md5Sum.
//
// This helper will read the entire tar.gz reader and assert all
// the expected content its present, by checking the md5 sums
// of each file. In case a folder is found, an empty string ""
// will be in the place of the sum.
func AssertMD5Sums(t *testing.T, r io.Reader, expected map[string]string) {
	gzipReader, err := gzip.NewReader(r)
	mustNoErr(err)
	tarReader := tar.NewReader(gzipReader)
	contents := map[string]string{}
	for {
		h, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		sum := md5.New()
		if !h.FileInfo().IsDir() {
			_, err = io.Copy(sum, tarReader)
			if err != nil {
				t.Fatal(err)
			}
			contents[h.Name] = fmt.Sprintf("%x", sum.Sum(nil))
			continue
		}
		contents[h.Name] = ""
	}
	assert.Equal(t, expected, contents)
}

func mustNoErr(err error) {
	if err != nil {
		panic(err)
	}
}
