package magic

import (
	"bytes"
	"path"
	"reflect"
	"runtime"
	"testing"
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

func Skip(t *testing.T, message string) {
	// XXX(krzysztof): Attempt to circumvent lack of T.Skip() prior to Go version go1.1 ...
	f := reflect.ValueOf(t).MethodByName("Skip")
	if ok := f.IsValid(); !ok {
		f = reflect.ValueOf(t).MethodByName("Log")
	}

	f.Call([]reflect.Value{reflect.ValueOf(message)})
}

func CompareStrings(this, other string) bool {
	if this == "" || other == "" {
		return false
	}
	return bytes.Equal([]byte(this), []byte(other))
}

func OldGoVersion() (bool, string) {
	// Contains every release of Go prior to
	// when the `os.Unsetenv()` function was
	// added in the version 1.4.x and newer.
	versions := []string{
		"go1", "go1.0.1",
		"go1.0.2", "go1.0.3",
		"go1.1", "go1.1.1", "go1.1.2",
		"go1.2", "go1.2.1", "go1.2.2",
		"go1.3", "go1.3.1", "go1.3.2", "go1.3.3",
	}

	version := runtime.Version()

	for _, v := range versions {
		if v == version {
			return true, version
		}
	}
	return false, version
}
