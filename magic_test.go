package magic

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
)

// Reference original functions.
var (
	version       = Version
	versionString = VersionString
	versionSlice  = VersionSlice
)

func init() {
	_, err := version()
	if err != nil && err.(*Error).Errno == int(syscall.ENOSYS) {
		// Mock return values by replacing original functions.
		version = func() (int, error) {
			return 518, nil
		}
		versionString = func() (string, error) {
			return "5.18", nil
		}
		versionSlice = func() ([]int, error) {
			return []int{5, 18}, nil
		}
	}
}

func TestNew(t *testing.T) {
	var mgc *Magic

	var n int
	var err error
	var p []string
	var v string

	mgc, err = New()
	if err != nil {
		t.Fatalf("unable to create new Magic type: %s", err.Error())
	}

	func(v interface{}) {
		if _, ok := v.(*Magic); !ok {
			t.Fatalf("not a Magic type: %s", reflect.TypeOf(v).String())
		}
	}(mgc)

	n, _ = Version()

	if n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))
	brokenMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-broken.magic"))

	mgc.Close()

	_, err = New("does/not/exist")

	v = "magic: could not find any valid magic files!"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: could not find any magic files!"
	}

	if err == nil {
		t.Errorf("value given {%v}, want {%v}", err.Error(), nil)
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given {%q}, want {%q}", err.Error(), v)
	}

	mgc, err = New(genuineMagicFile)
	if err != nil {
		t.Errorf("value given \"%T\", should be empty", err)
	}

	// Current path should change accordingly ...
	p, _ = mgc.Path()

	if ok := compareStrings(p[0], genuineMagicFile); !ok {
		t.Errorf("value given %q, want %q", p[0], genuineMagicFile)
	}

	mgc.Close()

	_, err = New(brokenMagicFile)

	v = "magic: line 1: No current entry for continuation"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: No current entry for continuation"
	}

	if err == nil {
		t.Errorf("value given \"%T\", should not be empty", err)
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given {%q}, want {%q}", err.Error(), v)
	}
}

func TestMagic_Close(t *testing.T) {
	mgc, _ := New()

	var cookie reflect.Value

	magic := reflect.ValueOf(mgc).Elem().FieldByName("magic").Elem()

	cookie = magic.FieldByName("cookie").Elem()
	if ok := cookie.IsValid(); !ok {
		t.Errorf("value given %v, want %v", ok, true)
	}

	mgc.Close()

	// Should be NULL (at C level) as magic_close() will free underlying Magic database.
	cookie = magic.FieldByName("cookie").Elem()
	if ok := cookie.IsValid(); ok {
		t.Errorf("value given %v, want %v", ok, false)
	}

	// Should be a no-op ...
	mgc.Close()
}

func TestMagic_String(t *testing.T) {
	mgc, _ := New()
	defer mgc.Close()

	magic := reflect.ValueOf(mgc).Elem().FieldByName("magic").Elem()
	path := magic.FieldByName("path")
	cookie := magic.FieldByName("cookie").Elem().UnsafeAddr()

	// Get whatever the underlying default path is ...
	paths := make([]string, path.Len())
	for i := 0; i < path.Len(); i++ {
		paths[i] = path.Index(i).String()
	}

	v := fmt.Sprintf("Magic{flags:%d path:%s cookie:0x%x}", 0, paths, cookie)
	if ok := compareStrings(mgc.String(), v); !ok {
		t.Errorf("value given %q, want %q", mgc.String(), v)
	}
}

func TestMagic_Path(t *testing.T) {
	var mgc *Magic
	var rv []string

	mgc, _ = New()
	mgc.Close()

	_, err := mgc.Path()

	v := "magic: Magic library is not open"
	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	mgc, _ = New()

	rv, _ = mgc.Path()
	if len(rv) == 0 {
		t.Fatalf("value given \"%T\", should not be empty", rv)
	}

	mgc.Close()

	// XXX(krzysztof): Setting "MAGIC" here breaks tests later as it will
	// be persistent between different tests, sadly needed to be disabled
	// for the time being.
	//
	//	p, err := os.Getwd()
	//	if err != nil {
	//		t.Fatal("unable to get current and/or working directory")
	//	}
	//
	//	p = path.Clean(path.Join(p, "fixtures"))
	//	if err = os.Setenv("MAGIC", p); err != nil {
	//		t.Fatalf("unable to set \"MAGIC\" environment variable to %q", p)
	//	}
	//
	//	v, _ = mgc.Path()
	//	if ok := compareStrings(v[0], p); !ok {
	//		t.Errorf("value given %q, want %q", v[0], p)
	//	}

	// TODO(kwilczynski): Test Magic.Load() affecting Magic.Path() as well. But
	// that requires working os.Clearenv() which is yet to be implemented as
	// per http://golang.org/src/pkg/syscall/env_unix.go?s=1772:1787#L101

	mgc, _ = open()
	defer func() {
		mgc.close()
	}()

	rv, _ = mgc.Path()
	if len(rv) == 0 {
		t.Fatalf("value given \"%T\", should not be empty", rv)
	}
}

func TestMagic_Path_Environment(t *testing.T) {
	if ok, v := oldGoVersion(); ok {
		skip(t, fmt.Sprintf("this version of the Go Language is too old: %s", v))
		return // Should not me reachable on modern Go version.
	}
}

func TestMagic_Flags(t *testing.T) {
	var mgc *Magic

	mgc, _ = New()
	mgc.Close()

	_, err := mgc.Flags()

	v := "magic: Magic library is not open"
	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	mgc, _ = New()
	defer mgc.Close()

	mgc.SetFlags(MIME)

	flags := MIME_TYPE | MIME_ENCODING
	if v, _ := mgc.Flags(); v != flags {
		t.Errorf("value given 0x%06x, want 0x%06x", v, flags)
	}
}

func TestMagic_SetFlags(t *testing.T) {
	var mgc *Magic

	var err error
	var actual, errno int
	var v string

	mgc, _ = New()
	mgc.Close()

	err = mgc.SetFlags(0)

	v = "magic: Magic library is not open"
	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	mgc, _ = New()
	defer mgc.Close()

	var flagsTests = []struct {
		broken   bool
		errno    int
		expected int
		given    int
	}{
		// Test lower boundary limit.
		{true, 22, 0x000000, -0xfffffff},
		// Genuine flags ...
		{false, 0, 0x000000, 0x000000}, // Flag: NONE
		{false, 0, 0x000010, 0x000010}, // Flag: MIME_TYPE
		{false, 0, 0x000400, 0x000400}, // Flag: MIME_ENCODING
		{false, 0, 0x000410, 0x000410}, // Flag: MIME_TYPE, MIME_ENCODING
		// Test upper boundary limit.
		{true, 22, 0x000410, 0xffffffff},
	}

	for _, tt := range flagsTests {
		err = mgc.SetFlags(tt.given)
		actual, _ = mgc.Flags()
		if err != nil && tt.broken {
			errno = err.(*Error).Errno
			if actual != tt.expected || errno != tt.errno {
				t.Errorf("value given {0x%06x %d}, want {0x%06x %d}", actual, errno, tt.expected, tt.errno)
				continue
			}
		}
		if actual != tt.expected {
			t.Errorf("value given 0x%06x, want 0x%06x", actual, tt.expected)
		}
	}

	err = mgc.SetFlags(0xffffffff)

	v = "magic: unknown or invalid flag specified"
	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}
}

func TestMagic_FlagsSlice(t *testing.T) {
	var mgc *Magic

	mgc, _ = New()
	mgc.Close()

	_, err := mgc.FlagsSlice()

	v := "magic: Magic library is not open"
	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	mgc, _ = New()
	defer mgc.Close()

	var actual []int

	var flagsSliceTests = []struct {
		given    int
		expected []int
	}{
		{0x000000, []int{0x000000}},           // Flag: NONE
		{0x000001, []int{0x000001}},           // Flag: DEBUG
		{0x000201, []int{0x000001, 0x000200}}, // Flag: DEBUG, ERROR
		{0x000022, []int{0x000002, 0x000020}}, // Flag: SYMLINK, CONTINUE
		{0x000410, []int{0x000010, 0x000400}}, // Flag: MIME_TTYPE, MIME_ENCODING
	}

	for _, tt := range flagsSliceTests {
		mgc.SetFlags(tt.given)

		actual, _ = mgc.FlagsSlice()
		if ok := reflect.DeepEqual(actual, tt.expected); !ok {
			t.Errorf("value given %v, want %v", actual, tt.expected)
		}
	}
}

func TestMagic_Load(t *testing.T) {
	var mgc *Magic

	var n int
	var rv bool
	var err error
	var p []string
	var v string

	mgc, _ = New()
	mgc.Close()

	_, err = mgc.Load("does/not/exist")

	v = "magic: Magic library is not open"
	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	mgc, _ = New()
	n, _ = Version()

	if n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))
	brokenMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-broken.magic"))

	rv, err = mgc.Load("does/not/exist")

	v = "magic: could not find any valid magic files!"
	if n < 0 {
		// A few releases of libmagic were having issues.
		v = "magic: could not find any magic files!"
	}

	if err == nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, false, v)
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), false, v)
	}

	// XXX(krzysztof): Currently, certain versions of libmagic API will *never*
	// clear an error once there is one, therefore a whole new session has to be
	// created in order to clear it. Unless upstream fixes this bad design choice,
	// there is nothing to do about it, sadly.
	mgc.Close()

	mgc, _ = New()

	rv, err = mgc.Load(genuineMagicFile)
	if err != nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, true, "")
	}

	// Current path should change accordingly ...
	p, _ = mgc.Path()

	if ok := compareStrings(p[0], genuineMagicFile); !ok {
		t.Errorf("value given %q, want %q", p[0], genuineMagicFile)
	}

	rv, err = mgc.Load(brokenMagicFile)

	v = "magic: line 1: No current entry for continuation"
	if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: No current entry for continuation"
	}

	if err == nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, false, v)
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), false, v)
	}

	// Since there was an error, path should remain the same.
	p, _ = mgc.Path()
	if ok := compareStrings(p[0], genuineMagicFile); !ok {
		t.Errorf("value given %q, want %q", p[0], genuineMagicFile)
	}

	mgc.Close()
}

func TestMagic_Load_Environment(t *testing.T) {
	if ok, v := oldGoVersion(); ok {
		skip(t, fmt.Sprintf("this version of the Go Language is too old: %s", v))
		return // Should not me reachable on modern Go version.
	}
}

func TestMagic_Compile(t *testing.T) {
	var mgc *Magic

	var n int
	var rv bool
	var err error
	var genuine, broken, v string

	mgc, _ = New()
	mgc.Close()

	_, err = mgc.Compile("does/not/exist")

	v = "magic: Magic library is not open"
	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	clean := func() {
		files, _ := filepath.Glob("*.mgc")
		for _, f := range files {
			os.Remove(f)
		}
	}

	mgc, _ = New()
	n, _ = Version()

	if n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))
	brokenMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-broken.magic"))

	rv, err = mgc.Compile("does/not/exist")

	v = "magic: could not find any valid magic files!"
	if n < 0 {
		// A few releases of libmagic were having issues.
		v = "magic: could not find any magic files!"
	}

	if err == nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, false, v)
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), false, v)
	}

	// See comment in TestMagic_Load() ...
	mgc.Close()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("unable to get current and/or working directory")
	}

	mgc, _ = New()
	defer mgc.Close()

	os.Chdir(path.Join(wd, path.Join(fixturesDirectory, formatDirectory)))
	defer func() {
		clean()
		os.Chdir(wd)
	}()

	clean()

	_, genuine = path.Split(genuineMagicFile)
	_, broken = path.Split(brokenMagicFile)

	// Re-define as we are no longer in top-level directory ...
	genuine = path.Clean(path.Join(".", genuine))
	broken = path.Clean(path.Join(".", broken))

	rv, err = mgc.Compile(genuine)
	if err != nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, true, "")
	}

	compiledMagicFile := fmt.Sprintf("%s.mgc", genuine)

	stat, err := os.Stat(compiledMagicFile)
	if stat == nil && err != nil {
		x := os.IsNotExist(err)
		t.Errorf("value given {%v %q}, want {%v %q}", x, err.Error(), false, "")
	}

	// Assuming that success would yield a non-zero size compiled Magic file ...
	if stat != nil && err == nil {
		x := os.IsNotExist(err)
		if s := stat.Size(); s < 5 {
			t.Errorf("value given {%v %d}, want {%v > %d}", x, s, false, 5)
		}

		buffer := make([]byte, 5)

		// Header (8 bytes) of the compiled Magic file should be: 1c 04 1e f1 08 00 00 00
		// on any little-endian architecture. Where the 5th byte always denotes which version
		// of the Magic database is it.
		expected := []byte{0x1c, 0x04, 0x1e, 0xf1}

		f, err := os.Open(compiledMagicFile)
		if err != nil {
			t.Fatalf("unable to open file `%s'", compiledMagicFile)
		}
		f.Read(buffer)
		f.Close()

		last := buffer[len(buffer)-1:][0] // Get version only ...
		buffer = buffer[:len(buffer)-1]

		ok := bytes.Equal(buffer, expected)
		if !ok || last == 0 {
			t.Errorf("value given {0x%x 0x%02x}, want {0x%x > 0x%02x}", buffer, last, expected, 0)
		}
	}

	rv, err = mgc.Compile(broken)

	v = "magic: line 1: No current entry for continuation"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: No current entry for continuation"
	}

	if err == nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, false, v)
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), false, v)
	}
}

func TestMagic_Check(t *testing.T) {
	var mgc *Magic

	var n int
	var rv bool
	var err error
	var v string

	mgc, _ = New()
	mgc.Close()

	_, err = mgc.Check("does/not/exist")

	v = "magic: Magic library is not open"
	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	mgc, _ = New()
	n, _ = Version()

	if n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))
	brokenMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-broken.magic"))

	rv, err = mgc.Check("does/not/exist")

	v = "magic: could not find any valid magic files!"
	if n < 0 {
		// A few releases of libmagic were having issues.
		v = "magic: could not find any magic files!"
	}

	if err == nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, false, v)
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), false, v)
	}

	// See comment in TestMagic_Load() ...
	mgc.Close()

	mgc, _ = New()
	defer mgc.Close()

	rv, err = mgc.Check(genuineMagicFile)
	if err != nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, true, "")
	}

	rv, err = mgc.Check(brokenMagicFile)

	v = "magic: line 1: No current entry for continuation"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: No current entry for continuation"
	}

	if err == nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, false, v)
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), false, v)
	}
}

func TestMagic_File(t *testing.T) {
	var mgc *Magic

	var ok bool
	var err error
	var v, rv string

	mgc, _ = New()
	mgc.Close()

	_, err = mgc.File("does/not/exist")

	v = "magic: Magic library is not open"
	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	mgc, _ = New()
	defer mgc.Close()

	if n, _ := Version(); n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))
	fakeMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-fake.magic"))

	mgc.SetFlags(NONE)
	mgc.Load(genuineMagicFile)

	rv, _ = mgc.File(sampleImageFile)

	v = "PNG image data, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	mgc.SetFlags(MIME)

	rv, _ = mgc.File(sampleImageFile)

	v = "image/png; charset=binary"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	mgc.SetFlags(NONE)
	mgc.Load(fakeMagicFile)

	rv, _ = mgc.File(sampleImageFile)

	v = "Go Gopher image, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	mgc.SetFlags(MIME)

	rv, _ = mgc.File(sampleImageFile)

	v = "image/x-go-gopher; charset=binary"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	mgc.SetFlags(NONE)

	// We expect no error to be risen at this time.
	rv, err = mgc.File("does/not/exist")
	if err != nil {
		t.Errorf("value given %q, want %q", rv, v)
	}

	// Return value contains the error message as per the IEEE 1003.1 standard.
	v = "cannot open `does/not/exist' (No such file or directory)"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	// Now we expect an error to be risen.
	mgc.SetFlags(ERROR)

	_, err = mgc.File("does/not/exist")

	v = "magic: cannot stat `does/not/exist' (No such file or directory)"
	if ok = compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}
}

func TestMagic_Buffer(t *testing.T) {
	var mgc *Magic

	var f *os.File

	var ok bool
	var err error
	var v, rv string

	mgc, _ = New()
	mgc.Close()

	_, err = mgc.Buffer([]byte{})

	v = "magic: Magic library is not open"
	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	mgc, _ = New()
	defer mgc.Close()

	if n, _ := Version(); n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))
	fakeMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-fake.magic"))

	buffer := &bytes.Buffer{}

	image := func() {
		f, err = os.Open(sampleImageFile)
		if err != nil {
			t.Fatalf("unable to open file `%s'", sampleImageFile)
		}
		io.Copy(buffer, f)
		f.Close()
	}

	image()

	mgc.SetFlags(NONE)
	mgc.Load(genuineMagicFile)

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "PNG image data, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	mgc.SetFlags(MIME)

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "image/png; charset=binary"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	mgc.SetFlags(NONE)
	mgc.Load(fakeMagicFile)

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "Go Gopher image, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	mgc.SetFlags(MIME)

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "image/x-go-gopher; charset=binary"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()
	buffer.WriteString("Hello, 世界")

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "text/plain; charset=utf-8"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	mgc.SetFlags(NONE)

	buffer.Reset()
	buffer.WriteString("#!/bin/bash\n\n")

	rv, err = mgc.Buffer(buffer.Bytes())

	// This is correct since custom Magic database was loaded,
	// libmagic does not have enough know-how to correctly
	// identify Bash scripts.
	v = "ASCII text"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	// Load two custom Magic databases now, one of which has
	// correct magic to detect Bash shell scripts.
	mgc.Load(genuineMagicFile, shellMagicFile)

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "Bourne-Again shell script, ASCII text executable"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()

	// Re-load Gopher PNG image ...
	image()

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "PNG image data, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()
	buffer.WriteString("#!/bin/sh\n\n")

	rv, err = mgc.Buffer(buffer.Bytes())

	// Quite redundant, but fun ...
	v = "POSIX shell script, ASCII text executable"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()
	buffer.Write([]byte{0x0})

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "very short file (no magic)"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()

	defer func() {
		r := recover()
		if r == nil {
			t.Error("did not panic")
			return
		}
		v = "runtime error: index out of range [0] with length 0"
		if ok := compareStrings(r.(error).Error(), v); !ok {
			t.Errorf("value given %q, want %q", r.(error).Error(), v)
			return
		}
	}()

	// Will panic ...
	mgc.Buffer(buffer.Bytes())
}

func TestMagic_Descriptor(t *testing.T) {
	var mgc *Magic

	var f *os.File

	var n int
	var ok bool
	var err error
	var v, rv string

	mgc, _ = New()
	mgc.Close()

	_, err = mgc.Descriptor(0)

	v = "magic: Magic library is not open"
	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	mgc, _ = New()
	defer mgc.Close()

	n, _ = Version()

	if n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))

	fakeMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-fake.magic"))

	// Sadly, the function `const char* magic_descriptor(struct magic_set*, int)',
	// which is a part of libmagic will *kindly* close file referenced by given
	// file descriptor for us, and so we have to re-open each time. This only
	// concerns certain versions of libmagic, but its better to stay on the
	// safe side ...
	file := func() {
		f, err = os.Open(sampleImageFile)
		if err != nil {
			t.Fatalf("unable to open file `%s'", sampleImageFile)
		}
	}

	file()

	mgc.SetFlags(NONE)
	mgc.Load(genuineMagicFile)

	rv, err = mgc.Descriptor(f.Fd())

	v = "PNG image data, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	f.Close()
	file()

	mgc.SetFlags(MIME)

	rv, err = mgc.Descriptor(f.Fd())

	v = "image/png; charset=binary"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	f.Close()
	file()

	mgc.SetFlags(NONE)
	mgc.Load(fakeMagicFile)

	rv, err = mgc.Descriptor(f.Fd())

	v = "Go Gopher image, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	f.Close()
	file()
	mgc.SetFlags(MIME)

	rv, err = mgc.Descriptor(f.Fd())

	v = "image/x-go-gopher; charset=binary"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	f.Close()

	_, err = mgc.Descriptor(f.Fd())

	v = "magic: bad file descriptor"
	if ok = compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	// Reading from standard input (0) will yield no data in this case.
	rv, err = mgc.Descriptor(0)

	v = "application/x-empty; charset=binary"
	if n, _ := Version(); n < 515 {
		// Older version of libmagic reports same error differently.
		v = "application/x-empty"
	}

	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}
}

func TestMagic_Separator(t *testing.T) {
	mgc, _ := New()
	defer mgc.Close()

	var flags int
	var rv string
	var actual []string

	n, _ := Version()

	v := []string{"Bourne-Again shell script text executable", "a /bin/bash script, ASCII text executable", "data"}
	if n < 524 || n > 536 {
		// Some older version of libmagic reported this correctly, and latest
		// version of libmagic has been fixed so it does not try to report text
		// files as "data" any more, albeit it returns an extra separator entry
		// which causes an extra entry to be added.
		v = []string{"Bourne-Again shell script text executable", "a /bin/bash script, ASCII text executable"}
	}
	if n > 533 && n < 535 {
		v = []string{"Bourne-Again shell script text executable", "a /bin/bash script, ASCII text executable", ""}
	}

	var separatorTests = []struct {
		flags    int
		expected []string
	}{
		// Flag: MAGIC_NONE
		{0x000000, []string{"Bourne-Again shell script, ASCII text executable"}},
		// Flag: MAGIC_CONTINUE
		{0x000020, v},
		// Flag: MIME_ENCODING
		{0x000400, []string{"us-ascii"}},
		// Flag: MIME_TYPE, MIME_ENCODING
		{0x000410, []string{"text/x-shellscript; charset=us-ascii"}},
	}

	buffer := []byte("#!/bin/bash\n\n")
	mgc.Load(shellMagicFile)

	for _, tt := range separatorTests {
		mgc.SetFlags(tt.flags)

		rv, _ = mgc.Buffer(buffer)
		flags, _ = mgc.Flags()

		actual = strings.Split(rv, Separator)
		if ok := reflect.DeepEqual(actual, tt.expected); !ok || flags != tt.flags {
			t.Errorf("value given {0x%x %v}, want {0x%x %v}", flags, actual, tt.flags, tt.expected)
		}
	}
}

func Test_open(t *testing.T) {
	mgc, _ := open()
	defer func() {
		mgc.close()
	}()

	func(v interface{}) {
		if _, ok := v.(*Magic); !ok {
			t.Fatalf("not a Magic type: %s", reflect.TypeOf(v).String())
		}
	}(mgc)

	magic := reflect.ValueOf(mgc).Elem().FieldByName("magic").Elem()
	path := magic.FieldByName("path")
	cookie := magic.FieldByName("cookie").Elem().UnsafeAddr()

	if path.Kind() != reflect.Slice || path.Len() > 0 {
		t.Errorf("value given {%v ?}, want {%v %d}", path.Kind(), reflect.Slice, 0)
	}

	if reflect.ValueOf(cookie).Kind() != reflect.Uintptr || cookie == 0 {
		t.Errorf("value given {%v 0x%x}, want {%v > %d}", reflect.ValueOf(cookie).Kind(), cookie, reflect.Uintptr, 0)
	}
}

func Test_close(t *testing.T) {
	mgc, _ := open()
	mgc.close()

	value := reflect.ValueOf(mgc).Elem().FieldByName("magic")
	path := value.Elem().FieldByName("path")
	cookie := value.Elem().FieldByName("cookie").Elem()

	if path.Kind() != reflect.Slice || path.Len() > 0 {
		t.Errorf("value given {%v ?}, want {%v %d}", path.Kind(), reflect.Slice, 0)
	}

	// Should be NULL (at C level) as magic_close() will free underlying Magic database.
	if ok := cookie.IsValid(); ok {
		t.Errorf("value given %v, want %v", ok, false)
	}
}

func Test_destroy(t *testing.T) {
	mgc, _ := open()
	mgc.destroy()

	magic := reflect.ValueOf(mgc).Elem().FieldByName("magic").Elem()
	path := magic.FieldByName("path")
	cookie := magic.FieldByName("cookie").Elem()

	if path.Kind() != reflect.Slice || path.Len() > 0 {
		t.Errorf("value given {%v ?}, want {%v %d}", path.Kind(), reflect.Slice, 0)
	}

	// Should be NULL (at C level) as magic_close() will free underlying Magic database.
	if ok := cookie.IsValid(); ok {
		t.Errorf("value given %v, want %v", ok, false)
	}
}

func TestOpen(t *testing.T) {
	var mgc *Magic

	var n int
	var ok bool
	var err error
	var rv, v string

	n, _ = Version()

	err = Open(nil)

	v = "magic: not a function or nil pointer"
	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	err = Open(func(m *Magic) error {
		// There should be an error originating from magic.New()
		// which is wrapped inside magic.Open().
		return nil
	}, "does/not/exist")

	v = "magic: could not find any valid magic files!"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: could not find any magic files!"
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	if n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))

	err = Open(func(m *Magic) error {
		m.Load(genuineMagicFile)
		a, b := m.File(sampleImageFile)
		rv = a   // Pass outside the closure for verification.
		return b // Or return nil here ...
	})

	if err != nil {
		if ok := compareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), false, v)
		}
	}

	v = "PNG image data, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	err = Open(func(m *Magic) error {
		// A canary value to test error propagation ...
		panic("123abc456")
	})

	v = "magic: 123abc456"
	if ok = compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	_ = Open(func(m *Magic) error {
		mgc = m // Pass outside the closure ...
		return nil
	})

	magic := reflect.ValueOf(mgc).Elem().FieldByName("magic").Elem()
	cookie := magic.FieldByName("cookie").Elem()

	// Should be NULL (at C level) as magic_close() will free underlying Magic database.
	if ok := cookie.IsValid(); ok {
		t.Errorf("value given %v, want %v", ok, false)
	}
}

func TestCompile(t *testing.T) {
	var rv bool
	var err error
	var genuine, broken, v string

	clean := func() {
		files, _ := filepath.Glob("*.mgc")
		for _, f := range files {
			os.Remove(f)
		}
	}

	n, _ := Version()

	_, err = Compile("does/not/exist")

	v = "magic: could not find any valid magic files!"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: could not find any magic files!"
	}

	if err == nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, false, v)
	}

	if ok := compareStrings(err.Error(), v); !ok && !rv {
		t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), false, v)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("unable to get current and/or working directory")
	}

	os.Chdir(path.Join(wd, path.Join(fixturesDirectory, formatDirectory)))
	defer func() {
		clean()
		os.Chdir(wd)
	}()

	clean()

	if n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))
	brokenMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-broken.magic"))

	_, genuine = path.Split(genuineMagicFile)
	_, broken = path.Split(brokenMagicFile)

	// Re-define as we are no longer in top-level directory ...
	genuine = path.Clean(path.Join(".", genuine))
	broken = path.Clean(path.Join(".", broken))

	rv, err = Compile(genuine)
	if err != nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, true, "")
	}

	rv, err = Compile(broken)

	v = "magic: line 1: No current entry for continuation"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: No current entry for continuation"
	}

	if err == nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, false, v)
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), false, v)
	}
}

func TestCheck(t *testing.T) {
	var rv bool
	var err error
	var v string

	n, _ := Version()

	_, err = Check("does/not/exist")

	v = "magic: could not find any valid magic files!"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: could not find any magic files!"
	}

	if err == nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, false, v)
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), false, v)
	}

	if n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))
	brokenMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-broken.magic"))

	rv, err = Check(genuineMagicFile)
	if err != nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, true, "")
	}

	rv, err = Check(brokenMagicFile)

	v = "magic: line 1: No current entry for continuation"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: No current entry for continuation"
	}

	if err == nil {
		t.Errorf("value given {%v \"%v\"}, want {%v %q}", rv, err, false, v)
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), false, v)
	}
}

func TestVersion(t *testing.T) {
	v, _ := version()
	if reflect.ValueOf(v).Kind() != reflect.Int || v <= 0 {
		t.Errorf("value given {%v %d}, want {%v > %d}", reflect.ValueOf(v).Kind(), v, reflect.Int, 0)
	}
}

func TestVersionString(t *testing.T) {
	rv, _ := version()

	s, _ := versionString()
	if reflect.ValueOf(s).Kind() != reflect.String || s == "" {
		t.Errorf("value given {%v %d}, want {%v > %d}", reflect.ValueOf(s).Kind(), len(s), reflect.String, 0)
	}

	v := fmt.Sprintf("%d.%02d", rv/100, rv%100)
	if ok := compareStrings(s, v); !ok {
		t.Errorf("value given %q, want %q", s, v)
	}
}

func TestVersionSlice(t *testing.T) {
	rv, _ := version()

	s, _ := versionSlice()
	if reflect.ValueOf(s).Kind() != reflect.Slice || len(s) == 0 {
		t.Errorf("value given {%v %d}, want {%v > %d}", reflect.ValueOf(s).Kind(), len(s), reflect.Slice, 0)
	}

	v := []int{rv / 100, rv % 100}
	if ok := reflect.DeepEqual(s, v); !ok {
		t.Errorf("value given %v, want %v", s, v)
	}
}

func TestFileMime(t *testing.T) {
	var ok bool
	var err error
	var v, rv string

	n, _ := Version()

	_, err = FileMime(sampleImageFile, "does/not/exist")

	v = "magic: could not find any valid magic files!"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: could not find any magic files!"
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	if n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))
	brokenMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-broken.magic"))
	fakeMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-fake.magic"))

	rv, err = FileMime("does/not/exist", genuineMagicFile)
	if rv == "" && err != nil {
		v = "magic: cannot open `does/not/exist' (No such file or directory)"
		if ok := compareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), "", v)
		}
	}

	rv, _ = FileMime(sampleImageFile, genuineMagicFile)

	v = "image/png; charset=binary"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	rv, _ = FileMime(sampleImageFile, fakeMagicFile)
	v = "image/x-go-gopher; charset=binary"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	rv, err = FileMime(sampleImageFile, brokenMagicFile)
	if rv == "" && err != nil {
		n, _ := Version()

		v = "magic: line 1: No current entry for continuation"
		if n < 518 && n >= 514 {
			// A few releases of libmagic were having issues.
			v = "magic: no magic files loaded"
		} else if n < 0 {
			// Older version of libmagic reports same error differently.
			v = "magic: No current entry for continuation"
		}

		if ok := compareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), false, v)
		}
	}
}

func TestFileType(t *testing.T) {
	var ok bool
	var err error
	var v, rv string

	n, _ := Version()

	_, err = FileType(sampleImageFile, "does/not/exist")

	v = "magic: could not find any valid magic files!"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: could not find any magic files!"
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	if n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))
	brokenMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-broken.magic"))
	fakeMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-fake.magic"))

	rv, err = FileType("does/not/exist", genuineMagicFile)
	if rv == "" && err != nil {
		v = "magic: cannot open `does/not/exist' (No such file or directory)"
		if ok := compareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), "", v)
		}
	}

	rv, _ = FileType(sampleImageFile, genuineMagicFile)
	v = "image/png"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	rv, _ = FileType(sampleImageFile, fakeMagicFile)
	v = "image/x-go-gopher"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	rv, _ = FileType(sampleImageFile, brokenMagicFile)
	if rv == "" && err != nil {
		n, _ := Version()

		v = "magic: line 1: No current entry for continuation"
		if n < 518 && n >= 514 {
			// A few releases of libmagic were having issues.
			v = "magic: no magic files loaded"
		} else if n < 0 {
			// Older version of libmagic reports same error differently.
			v = "magic: No current entry for continuation"
		}

		if ok := compareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), false, v)
		}
	}
}

func TestFileEncoding(t *testing.T) {
	var ok bool
	var err error
	var v, rv string

	n, _ := Version()

	_, err = FileEncoding(sampleImageFile, "does/not/exist")

	v = "magic: could not find any valid magic files!"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: could not find any magic files!"
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	if n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))
	brokenMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-broken.magic"))
	fakeMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-fake.magic"))

	rv, err = FileEncoding("does/not/exist", genuineMagicFile)
	if rv == "" && err != nil {
		v = "magic: cannot open `does/not/exist' (No such file or directory)"
		if ok := compareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), "", v)
		}
	}

	v = "binary" // Binary data will always have this encoding.

	rv, _ = FileEncoding(sampleImageFile, genuineMagicFile)
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	rv, _ = FileEncoding(sampleImageFile, fakeMagicFile)
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	rv, _ = FileEncoding(sampleImageFile, brokenMagicFile)
	if rv == "" && err != nil {
		n, _ := Version()

		v = "magic: line 1: No current entry for continuation"
		if n < 518 && n >= 514 {
			// A few releases of libmagic were having issues.
			v = "magic: no magic files loaded"
		} else if n < 0 {
			// Older version of libmagic reports same error differently.
			v = "magic: No current entry for continuation"
		}

		if ok := compareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v %q}, want {%v %q}", rv, err.Error(), false, v)
		}
	}
}

func TestBufferMime(t *testing.T) {
	var ok bool
	var err error
	var v, rv string

	n, _ := Version()

	_, err = BufferMime([]byte{}, "does/not/exist")

	v = "magic: could not find any valid magic files!"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: could not find any magic files!"
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	if n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))
	fakeMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-fake.magic"))

	buffer := &bytes.Buffer{}

	f, err := os.Open(sampleImageFile)
	if err != nil {
		t.Fatalf("unable to open file `%s'", sampleImageFile)
	}
	io.Copy(buffer, f)
	f.Close()

	rv, _ = BufferMime(buffer.Bytes(), genuineMagicFile)

	v = "image/png; charset=binary"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	rv, _ = BufferMime(buffer.Bytes(), fakeMagicFile)

	v = "image/x-go-gopher; charset=binary"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()
	buffer.WriteString("Hello, 世界")

	rv, _ = BufferMime(buffer.Bytes())

	v = "text/plain; charset=utf-8"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()
	buffer.WriteString("#!/bin/bash\n\n")

	rv, _ = BufferMime(buffer.Bytes())

	v = "text/x-shellscript; charset=us-ascii"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()
	buffer.Write([]byte{0x0})

	rv, _ = BufferMime(buffer.Bytes())

	v = "application/octet-stream; charset=binary"
	if n, _ := Version(); n < 515 {
		// A few releases of libmagic were having issues.
		v = "application/octet-stream"
	}

	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()

	defer func() {
		r := recover()
		if r == nil {
			t.Error("did not panic")
			return
		}

		v = "runtime error: index out of range [0] with length 0"
		if ok := compareStrings(r.(error).Error(), v); !ok {
			t.Errorf("value given %q, want %q", r.(error).Error(), v)
			return
		}
	}()

	// Will panic ...
	BufferMime(buffer.Bytes())
}

func TestBufferType(t *testing.T) {
	var ok bool
	var err error
	var v, rv string

	n, _ := Version()

	_, err = BufferType([]byte{}, "does/not/exist")

	v = "magic: could not find any valid magic files!"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: could not find any magic files!"
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	if n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))
	fakeMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-fake.magic"))

	buffer := &bytes.Buffer{}

	f, err := os.Open(sampleImageFile)
	if err != nil {
		t.Fatalf("unable to open file `%s'", sampleImageFile)
	}
	io.Copy(buffer, f)
	f.Close()

	rv, _ = BufferType(buffer.Bytes(), genuineMagicFile)

	v = "image/png"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	rv, _ = BufferType(buffer.Bytes(), fakeMagicFile)

	v = "image/x-go-gopher"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()
	buffer.WriteString("Hello, 世界")

	rv, _ = BufferType(buffer.Bytes())

	v = "text/plain"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()
	buffer.WriteString("#!/bin/bash\n\n")

	rv, _ = BufferType(buffer.Bytes())

	v = "text/x-shellscript"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()
	buffer.Write([]byte{0x0})

	rv, _ = BufferType(buffer.Bytes())

	v = "application/octet-stream"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()

	defer func() {
		r := recover()
		if r == nil {
			t.Error("did not panic")
			return
		}
		v = "runtime error: index out of range [0] with length 0"
		if ok := compareStrings(r.(error).Error(), v); !ok {
			t.Errorf("value given %q, want %q", r.(error).Error(), v)
			return
		}
	}()

	// Will panic ...
	BufferType(buffer.Bytes())
}

func TestBufferEncoding(t *testing.T) {
	var ok bool
	var err error
	var v, rv string

	n, _ := Version()

	_, err = BufferEncoding([]byte{}, "does/not/exist")

	v = "magic: could not find any valid magic files!"
	if n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	} else if n < 0 {
		// Older version of libmagic reports same error differently.
		v = "magic: could not find any magic files!"
	}

	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given %q, want %q", err.Error(), v)
	}

	if n >= 519 {
		formatDirectory = "new-format"
	}

	genuineMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png.magic"))
	fakeMagicFile := path.Clean(path.Join(fixturesDirectory, formatDirectory, "png-fake.magic"))

	buffer := &bytes.Buffer{}

	f, err := os.Open(sampleImageFile)
	if err != nil {
		t.Fatalf("unable to open file `%s'", sampleImageFile)
	}
	io.Copy(buffer, f)
	f.Close()

	rv, _ = BufferEncoding(buffer.Bytes(), genuineMagicFile)

	v = "binary"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	rv, _ = BufferEncoding(buffer.Bytes(), fakeMagicFile)

	v = "binary"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()
	buffer.WriteString("Hello, 世界")

	rv, _ = BufferEncoding(buffer.Bytes())

	v = "utf-8"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()
	buffer.WriteString("#!/bin/bash\n\n")

	rv, _ = BufferEncoding(buffer.Bytes())

	v = "us-ascii"
	if ok = compareStrings(rv, v); !ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()
	buffer.Write([]byte{0x0})

	rv, _ = BufferEncoding(buffer.Bytes())

	v = "" // Should be empty ...
	if ok = compareStrings(rv, v); ok {
		t.Errorf("value given %q, want %q", rv, v)
	}

	buffer.Reset()

	defer func() {
		r := recover()
		if r == nil {
			t.Error("did not panic")
			return
		}
		v = "runtime error: index out of range [0] with length 0"
		if ok := compareStrings(r.(error).Error(), v); !ok {
			t.Errorf("value given %q, want %q", r.(error).Error(), v)
			return
		}
	}()

	// Will panic ...
	BufferEncoding(buffer.Bytes())
}
