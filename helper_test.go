package magic

import (
	"bytes"
	"path"
)

var (
	// Directory containing test fixtures, etc.
	testDirectory = "test"

	// Auxiliary files for use in tests ...
	fixturesDirectory = path.Clean(path.Join(testDirectory, "fixtures"))

	// Default directory containing files using old-style Magic format.
	formatDirectory = "old-format"

	// PNG image data, 1634 x 2224, 8-bit/color RGBA, non-interlaced
	sampleImageFile = path.Clean(path.Join(fixturesDirectory, "gopher.png"))

	// Magic file for testing only ...
	shellMagicFile = path.Clean(path.Join(fixturesDirectory, "shell.magic"))
)

func compareStrings(this, other string) bool {
	if this == "" || other == "" {
		return false
	}
	return bytes.Equal([]byte(this), []byte(other))
}
