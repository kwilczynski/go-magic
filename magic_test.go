/*
 * magic_test.go
 *
 * Copyright 2013 Krzysztof Wilczynski
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package magic_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"

	. "github.com/kwilczynski/go-magic"
)

const (
	fixturesDirectory = "fixtures"

	// PNG image data, 1634 x 2224, 8-bit/color RGBA, non-interlaced
	sampleImageFile = "gopher.png"

	// Magic files for testing only ...
	genuineMagicFile = "png.magic"
	brokenMagicFile  = "png-broken.magic"
	fakeMagicFile    = "png-fake.magic"
	shellMagicFile   = "shell.magic"
)

var (
	image   string
	genuine string
	broken  string
	fake    string
	shell   string
)

func CompareStrings(this, other string) bool {
	if this == "" || other == "" {
		return false
	}
	return bytes.Equal([]byte(this), []byte(other))
}

func init() {
	image = path.Clean(path.Join(fixturesDirectory, sampleImageFile))
	genuine = path.Clean(path.Join(fixturesDirectory, genuineMagicFile))
	broken = path.Clean(path.Join(fixturesDirectory, brokenMagicFile))
	fake = path.Clean(path.Join(fixturesDirectory, fakeMagicFile))
	shell = path.Clean(path.Join(fixturesDirectory, shellMagicFile))
}

func TestNew(t *testing.T) {
	mgc, err := New()
	if err != nil {
		t.Fatalf("unable to create new Magic type: %s", err.Error())
	}
	defer mgc.Close()

	func(v interface{}) {
		if _, ok := v.(*Magic); !ok {
			t.Fatalf("not a Magic type: %s", reflect.TypeOf(mgc).String())
		}
	}(mgc)
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
	cookie := magic.FieldByName("cookie").Elem().Index(0).UnsafeAddr()

	// Get whatever the underlying default path is ...
	paths := make([]string, path.Len())
	for i := 0; i < path.Len(); i++ {
		paths[i] = path.Index(i).String()
	}

	v := fmt.Sprintf("Magic{flags:%d path:%s cookie:0x%x}", 0, paths, cookie)
	if ok := CompareStrings(mgc.String(), v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", mgc.String(), v)
	}
}

func TestMagic_Path(t *testing.T) {
	mgc, _ := New()
	defer mgc.Close()

	v, _ := mgc.Path()
	if len(v) == 0 {
		t.Fatalf("value given \"%T\", should not be empty", v)
	}

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
	//		t.Fatalf("unable to set \"MAGIC\" environment variable to \"%s\"", p)
	//	}
	//
	//	v, _ = mgc.Path()
	//	if ok := CompareStrings(v[0], p); !ok {
	//		t.Errorf("value given \"%s\", want \"%s\"", v[0], p)
	//	}

	// TODO(kwilczynski): Test Magic.Load() affecting Magic.Path() as well. But
	// that requires working os.Clearenv() which is yet to be implemented as
	// per http://golang.org/src/pkg/syscall/env_unix.go?s=1772:1787#L101
}

func TestMagic_Flags(t *testing.T) {
	mgc, _ := New()
	defer mgc.Close()

	mgc.SetFlags(MIME)

	flags := MIME_TYPE | MIME_ENCODING
	if v, _ := mgc.Flags(); v != flags {
		t.Errorf("value given 0x%06x, want 0x%06x", v, flags)
	}
}

func TestMagic_SetFlags(t *testing.T) {
	mgc, _ := New()
	defer mgc.Close()

	var err error
	var actual, errno int

	var flagsTests = []struct {
		broken   bool
		errno    int
		expected int
		given    int
	}{
		// Test lower boundary limit.
		{true, 22, 0x000000, -0xffffff},
		// Genuine flags ...
		{false, 0, 0x000000, 0x000000}, // Flag: NONE
		{false, 0, 0x000010, 0x000010}, // Flag: MIME_TYPE
		{false, 0, 0x000400, 0x000400}, // Flag: MIME_ENCODING
		{false, 0, 0x000410, 0x000410}, // Flag: MIME_TYPE | MIME_ENCODING
		// Test upper boundary limit.
		{true, 22, 0x000410, 0xffffff},
	}

	for _, tt := range flagsTests {
		err = mgc.SetFlags(tt.given)
		actual, _ = mgc.Flags()
		if err != nil && tt.broken {
			errno = err.(*MagicError).Errno
			if actual != tt.expected || errno != tt.errno {
				t.Errorf("value given {0x%06x %d}, want {0x%06x %d}",
					actual, errno, tt.expected, tt.errno)
				continue
			}
		}
		if actual != tt.expected {
			t.Errorf("value given 0x%06x, want 0x%06x", actual, tt.expected)
		}
	}

}

func TestMagic_Load(t *testing.T) {
	var mgc *Magic

	var rv bool
	var err error
	var p []string

	mgc, _ = New()

	rv, err = mgc.Load("does/not/exist")
	if rv && err != nil {
		v := "magic: could not find any magic files!"
		if ok := CompareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
				rv, err.Error(), false, v)
		}
	}

	// XXX(krzysztof): Currently, libmagic API will *never* clear an error once
	// there is one, therefore a whole new session has to be created in order to
	// clear it. Unless upstream fixes this bad design choice, there is nothing
	// to do about it, sadly.
	mgc.Close()

	mgc, _ = New()

	rv, err = mgc.Load(genuine)
	if !rv && err != nil {
		t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
			rv, err.Error(), true, "")
	}

	// Current path should change accordingly ...
	p, _ = mgc.Path()

	if ok := CompareStrings(p[0], genuine); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", p[0], genuine)
	}

	rv, err = mgc.Load(broken)
	if !rv && err != nil {
		v := "magic: No current entry for continuation"
		if ok := CompareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
				rv, err.Error(), false, v)
		}
	}

	// Since there was an error, path should remain the same.
	p, _ = mgc.Path()
	if ok := CompareStrings(p[0], genuine); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", p[0], genuine)
	}

	mgc.Close()
}

func TestMagic_Compile(t *testing.T) {
	var mgc *Magic

	var rv bool
	var err error

	clean := func() {
		files, _ := filepath.Glob("*.mgc")
		for _, f := range files {
			os.Remove(f)
		}
	}

	mgc, _ = New()

	rv, err = mgc.Compile("does/not/exist")
	if !rv && err != nil {
		v := "magic: could not find any magic files!"
		if ok := CompareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
				rv, err.Error(), false, v)
		}
	}

	// See comment in TestMagic_Load() ...
	mgc.Close()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("unable to get current and/or working directory")
	}

	mgc, _ = New()
	defer mgc.Close()

	os.Chdir(path.Join(wd, fixturesDirectory))

	// Re-define as we are no longer in top-level directory ...
	genuine := path.Clean(path.Join(".", genuineMagicFile))
	broken := path.Clean(path.Join(".", brokenMagicFile))

	compiled := fmt.Sprintf("%s.mgc", genuine)

	defer func() {
		clean()
		os.Chdir(wd)
	}()

	clean()

	rv, err = mgc.Compile(genuine)
	if !rv && err != nil {
		t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
			rv, err.Error(), true, "")
	}

	stat, err := os.Stat(compiled)
	if stat == nil && err != nil {
		v := os.IsNotExist(err)
		t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
			v, err.Error(), false, "")
	}

	// Assuming that success would yield a non-zero size compiled Magic file ...
	if stat != nil && err == nil {
		v := os.IsNotExist(err)
		if s := stat.Size(); s < 5 {
			t.Errorf("value given {%v %d}, want {%v > 5}",
				v, s, false)
		}

		buffer := make([]byte, 5)

		// Header (8 bytes) of the compiled Magic file should be: 1c 04 1e f1 08 00 00 00
		// on any little-endian architecture. Where the 5th byte always denotes which version
		// of the Magic database is it.
		expected := []byte{0x1c, 0x04, 0x1e, 0xf1}

		f, err := os.Open(compiled)
		if err != nil {
			t.Fatalf("")
		}
		f.Read(buffer)
		f.Close()

		last := buffer[len(buffer)-1:][0] // Get version only ...
		buffer = buffer[:len(buffer)-1]

		ok := bytes.Equal(buffer, expected)
		if !ok || last <= 0 {
			t.Errorf("value given {0x%x 0x%02x}, want {0x%x > 0x00}",
				buffer, last, expected)
		}
	}

	rv, err = mgc.Compile(broken)
	if !rv && err != nil {
		v := "magic: No current entry for continuation"
		if ok := CompareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
				rv, err.Error(), false, v)
		}
	}
}

func TestMagic_Check(t *testing.T) {
	var mgc *Magic

	var rv bool
	var err error

	mgc, _ = New()

	rv, err = mgc.Check("does/not/exist")
	if !rv && err != nil {
		v := "magic: could not find any magic files!"
		if ok := CompareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
				rv, err.Error(), false, v)
		}
	}

	// See comment in TestMagic_Load() ...
	mgc.Close()

	mgc, _ = New()
	defer mgc.Close()

	rv, err = mgc.Check(genuine)
	if !rv && err != nil {
		t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
			rv, err.Error(), true, "")
	}

	rv, err = mgc.Check(broken)
	if !rv && err != nil {
		v := "magic: No current entry for continuation"
		if ok := CompareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
				rv, err.Error(), false, v)
		}
	}
}

func TestMagic_File(t *testing.T) {
	mgc, _ := New()
	defer mgc.Close()

	var ok bool
	var err error
	var v, rv string

	mgc.SetFlags(NONE)
	mgc.Load(genuine)

	rv, err = mgc.File(image)

	v = "PNG image data, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	mgc.SetFlags(MIME)

	rv, err = mgc.File(image)

	v = "image/png; charset=binary"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	mgc.SetFlags(NONE)
	mgc.Load(fake)

	rv, err = mgc.File(image)

	v = "Go Gopher image, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	mgc.SetFlags(MIME)

	rv, err = mgc.File(image)

	v = "image/x-go-gopher; charset=binary"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	rv, err = mgc.File("does/not/exist")

	v = "magic: cannot open `does/not/exist' (No such file or directory)"
	if ok = CompareStrings(err.Error(), v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}
}

func TestMagic_Buffer(t *testing.T) {
	mgc, _ := New()
	defer mgc.Close()

	var f *os.File

	var ok bool
	var err error
	var v, rv string

	buffer := &bytes.Buffer{}

	image := func() {
		f, err = os.Open(image)
		if err != nil {
			t.Fatalf("")
		}
		io.Copy(buffer, f)
		f.Close()
	}

	image()

	mgc.SetFlags(NONE)
	mgc.Load(genuine)

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "PNG image data, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	mgc.SetFlags(MIME)

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "image/png; charset=binary"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	mgc.SetFlags(NONE)
	mgc.Load(fake)

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "Go Gopher image, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	mgc.SetFlags(MIME)

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "image/x-go-gopher; charset=binary"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	buffer.Reset()
	buffer.WriteString("Hello, 世界")

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "text/plain; charset=utf-8"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	mgc.SetFlags(NONE)

	buffer.Reset()
	buffer.WriteString("#!/bin/bash\n")

	rv, err = mgc.Buffer(buffer.Bytes())

	// This is correct since custom Magic database was loaded,
	// libmagic does not have enough know-how to correctly
	// identify Bash scripts.
	v = "ASCII text"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	// Load two custom Magic databases now, one of which has
	// correct magic to detect Bash shell scripts.
	mgc.Load(genuine, shell)

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "Bourne-Again shell script, ASCII text executable"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	buffer.Reset()

	// Re-load Gopher PNG image ...
	image()

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "PNG image data, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	buffer.Reset()
	buffer.WriteString("#!/bin/sh\n")

	rv, err = mgc.Buffer(buffer.Bytes())

	// Quite redundant, but fun ...
	v = "POSIX shell script, ASCII text executable"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	buffer.Reset()
	buffer.Write([]byte{0x0})

	rv, err = mgc.Buffer(buffer.Bytes())

	v = "very short file (no magic)"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	buffer.Reset()

	defer func() {
		r := recover()
		if r == nil {
			t.Error("did not panic")
			return
		}
		v = "runtime error: index out of range"
		if ok := CompareStrings(r.(error).Error(), v); !ok {
			t.Errorf("value given \"%s\", want \"%s\"",
				r.(error).Error(), v)
			return
		}
	}()

	// Will panic ...
	mgc.Buffer(buffer.Bytes())
}

func TestMagic_Descriptor(t *testing.T) {
	mgc, _ := New()
	defer mgc.Close()

	var f *os.File

	var ok bool
	var err error
	var v, rv string

	// Sadly, the function `const char* magic_descriptor(struct magic_set*, int)',
	// which is a part of libmagic will *kindly* close file referenced by given
	// file-descriptor for us, and so we have to re-open each time ...
	file := func() {
		f, err = os.Open(image)
		if err != nil {
			t.Fatalf("")
		}
	}

	file()

	mgc.SetFlags(NONE)
	mgc.Load(genuine)

	rv, err = mgc.Descriptor(f.Fd())

	v = "PNG image data, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	f.Close()
	file()

	mgc.SetFlags(MIME)

	rv, err = mgc.Descriptor(f.Fd())

	v = "image/png; charset=binary"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	f.Close()
	file()

	mgc.SetFlags(NONE)
	mgc.Load(fake)

	rv, err = mgc.Descriptor(f.Fd())

	v = "Go Gopher image, 1634 x 2224, 8-bit/color RGBA, non-interlaced"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	f.Close()
	file()
	mgc.SetFlags(MIME)

	rv, err = mgc.Descriptor(f.Fd())

	v = "image/x-go-gopher; charset=binary"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	f.Close()

	rv, err = mgc.Descriptor(f.Fd())

	v = "magic: cannot read `(null)' (Bad file descriptor)"
	if ok = CompareStrings(err.Error(), v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

	// Reading from standard input (0) will yield no data in this case.
	rv, err = mgc.Descriptor(0)

	v = "application/x-empty"
	if ok = CompareStrings(rv, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", rv, v)
	}

}

func TestOpen(t *testing.T) {
}

func TestCompile(t *testing.T) {
	var rv bool
	var err error

	clean := func() {
		files, _ := filepath.Glob("*.mgc")
		for _, f := range files {
			os.Remove(f)
		}
	}

	rv, err = Compile("does/not/exist")
	v := "magic: could not find any magic files!"
	if ok := CompareStrings(err.Error(), v); !ok && !rv {
		t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
			rv, err.Error(), false, v)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("unable to get current and/or working directory")
	}

	os.Chdir(path.Join(wd, fixturesDirectory))

	// Re-define as we are no longer in top-level directory ...
	genuine := path.Clean(path.Join(".", genuineMagicFile))
	broken := path.Clean(path.Join(".", brokenMagicFile))

	defer func() {
		clean()
		os.Chdir(wd)
	}()

	clean()

	rv, err = Compile(genuine)
	if !rv && err != nil {
		t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
			rv, err.Error(), true, "")
	}

	rv, err = Compile(broken)
	if !rv && err != nil {
		v := "magic: No current entry for continuation"
		if ok := CompareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
				rv, err.Error(), false, v)
		}
	}
}

func TestCheck(t *testing.T) {
	var rv bool
	var err error

	rv, err = Check("does/not/exist")
	if !rv && err != nil {
		v := "magic: could not find any magic files!"
		if ok := CompareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
				rv, err.Error(), false, v)
		}
	}

	rv, err = Check(genuine)
	if !rv && err != nil {
		t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
			rv, err.Error(), true, "")
	}

	rv, err = Check(broken)
	if !rv && err != nil {
		v := "magic: No current entry for continuation"
		if ok := CompareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
				rv, err.Error(), false, v)
		}
	}
}

func TestVersion(t *testing.T) {
	// XXX(krzysztof): Attempt to circumvent lack of T.Skip() prior to Go version go1.1 ...
	f := reflect.ValueOf(t).MethodByName("Skip")
	if ok := f.IsValid(); !ok {
		f = reflect.ValueOf(t).MethodByName("Log")
	}

	v, err := Version()
	if err != nil && err.(*MagicError).Errno == int(syscall.ENOSYS) {
		f.Call([]reflect.Value{
			reflect.ValueOf("function `int magic_version(void)' is not implemented"),
		})
		return // Should not me reachable on modern Go version.
	}

	if reflect.ValueOf(v).Kind() != reflect.Int || v <= 0 {
		t.Errorf("value given {%v %d}, want {%v > 0}",
			reflect.ValueOf(v).Kind(), v, reflect.Int)
	}
}

func TestFileMime(t *testing.T) {
}

func TestFileEncoding(t *testing.T) {
}

func TestFileType(t *testing.T) {
}

func TestBufferMime(t *testing.T) {
}

func TestBufferEncoding(t *testing.T) {
}

func TestBufferType(t *testing.T) {
}
