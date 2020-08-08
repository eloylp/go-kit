// This file serves as central point for testing
// common infrastructure.
package check_test

import (
	"strings"
)

const (
	contentTypeHeader = "Content-Type"
	locationHeader    = "Location"
	tuxFilePath       = "tux.png"
	tuxFileMD5Sum     = "f0826ff9a4a21078ff0d0ab55b4210f0"
)

func removeSpacesAndTabs(s string) string {
	s = strings.Replace(s, "\t", "", -1)
	s = strings.Replace(s, " ", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	return s
}
