package pathutil

import (
	"fmt"
	"path/filepath"
	"strings"
)

// RelativePath extracts the relative path part of requirePath
// argument taking in count the root one. This is an example:
// root == /var/www/html
// requiredPath == /var/www/html/docs
// The result will be "docs".
func RelativePath(root, requiredPath string) (string, error) {
	rel, err := filepath.Rel(root, requiredPath)
	if err != nil {
		return "", err
	}
	result := filepath.ToSlash(rel)
	return result, err
}

// PathInRoot will check that the provided argument
// path is a sub location of the provided argument rootPath.
// Note this only do the check by calculating the absolute
// paths and comparing them (string comparison). Other
// checks like symbolic links are not covered.
func PathInRoot(rootPath, path string) error {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(absPath, absRoot) {
		return fmt.Errorf("the path you provided %s is not a suitable one", path)
	}
	return nil
}
