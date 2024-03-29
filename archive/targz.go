// Package archive for grouping utilities that help us
// to deal with multiple files.
package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// TARGZ creates a new tar.gz file in the provided path
// by inspecting the provided sources.
func TARGZ(filePath string, srcPaths ...string) (int64, error) {
	f, err := os.Create(filePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return StreamTARGZ(f, srcPaths...)
}

// StreamTARGZ will write a compressed .tar.gz stream to the passed io.Writer
// from the specified paths. If the path is a directory, it will find all files
// and folder recursively and add it to the tar.gz stream. If the path passed is
// a single file, will only add that file to the stream.
// The returned written bytes does not include headers size.
func StreamTARGZ(writer io.Writer, paths ...string) (int64, error) {
	gzipReader := gzip.NewWriter(writer)
	defer gzipReader.Close()
	tarReader := tar.NewWriter(gzipReader)
	defer tarReader.Close()

	var totalBytes int64
	for _, path := range paths {
		pathInfo, err := os.Stat(path)
		if err != nil {
			return 0, err
		}
		if !pathInfo.IsDir() {
			b, err := tarFromFile(path, tarReader)
			if err != nil {
				return 0, err
			}
			totalBytes += b
			continue
		}
		b, err := tarFromDir(path, tarReader)
		if err != nil {
			return 0, err
		}
		totalBytes += b
	}
	return totalBytes, nil
}

func tarFromDir(path string, tarWriter *tar.Writer) (int64, error) {
	var totalBytes int64
	err := filepath.Walk(path, func(currentPath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name, err = relativePath(path, currentPath)
		if err != nil {
			return err
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if !info.IsDir() {
			b, err := appendToWriter(tarWriter, currentPath)
			if err != nil {
				return err
			}
			totalBytes += b
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return totalBytes, nil
}

func tarFromFile(path string, tarStream *tar.Writer) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return 0, err
	}
	header.Name = filepath.Base(path)
	if err := tarStream.WriteHeader(header); err != nil {
		return 0, err
	}
	b, err := appendToWriter(tarStream, path)
	if err != nil {
		return 0, err
	}
	return b, nil
}

func appendToWriter(w io.Writer, path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	b, err := io.Copy(w, file)
	if err != nil {
		return 0, err
	}
	return b, nil
}

// ExtractTARGZ will extract the provided tar.gz file
// into the provided path.
func ExtractTARGZ(dst, path string) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return ExtractTARGZStream(f, dst)
}

// ExtractTARGZStream will read the provided stream, that is supposed to be
// a tar.gz one and extract all the elements in the provided path. The
// provided path must be an absolute one.
//
// It will prevent directory escalation. If one of the headers contains
// a path outside the provided one, will return an error and will not
// clean operation done until that moment.
//
// The returned written bytes does not include headers.
func ExtractTARGZStream(stream io.Reader, path string) (int64, error) {
	if !filepath.IsAbs(path) {
		return 0, fmt.Errorf("the extraction path must be absolute")
	}
	gzipReader, err := gzip.NewReader(stream)
	if err != nil {
		return 0, fmt.Errorf("failed reading compressed gzip: %v", err)
	}
	tarReader := tar.NewReader(gzipReader)
	var totalBytes int64
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("failed reading next part of tar: %v", err)
		}
		extractionPath := filepath.Join(path, header.Name) //nolint:gosec
		err = pathInRoot(path, extractionPath)
		if err != nil {
			return 0, fmt.Errorf("path in root check: %v", err)
		}
		// Start processing types
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(extractionPath, 0755); err != nil {
				return 0, fmt.Errorf("failed creating dir %s part of tar: %v", path, err)
			}
		case tar.TypeReg:
			dir := filepath.Dir(extractionPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return 0, fmt.Errorf("failed creating dir %s part of tar: %v", dir, err)
			}
			outFile, err := os.Create(extractionPath)
			if err != nil {
				return 0, fmt.Errorf("failed creating file part %s of tar: %v", path, err)
			}
			b, err := io.Copy(outFile, tarReader) // nolinter: gosec (must be controlled by read/write timeouts)
			if err != nil {
				return 0, fmt.Errorf("failed copying data of file %s part of tar: %v", path, err)
			}
			totalBytes += b
			if err := outFile.Close(); err != nil {
				return totalBytes, fmt.Errorf("failed closing file %s part of tar: %v", path, err)
			}
		default:
			return 0, fmt.Errorf("unknown part of tar: type: %v in %s", header.Typeflag, header.Name)
		}
	}
	return totalBytes, nil
}
