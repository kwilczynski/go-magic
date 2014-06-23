// +build !windows

/*
 * magic.go
 *
 * Copyright 2013-2014 Krzysztof Wilczynski
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

package magic

/*
#cgo LDFLAGS: -lmagic
#cgo !darwin LDFLAGS: -Wl,--as-needed
#cgo CFLAGS: -std=gnu99

#include "functions.h"
*/
import "C"

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

// Separator is a field separator that can be used to split
// results when the CONTINUE flag is set causing all valid
// matches found by the Magic library to be returned.
const Separator string = "\x5c\x30\x31\x32\x2d\x20"

type magic struct {
	sync.Mutex
	flags  int
	path   []string
	cookie C.magic_t
}

func (m *magic) close() {
	if m != nil && m.cookie != nil {
		// This will free resources on the Magic library side.
		C.magic_close(m.cookie)
		m.path = []string{}
		m.cookie = nil
	}
	runtime.SetFinalizer(m, nil)
}

// Magic represents the underlying Magic library.
type Magic struct {
	*magic
}

// New opens and initializes Magic library.  Optionally,
// a multiple distinct Magic database files can be provided
// to load, otherwise a default database (usually available
// system-wide) will be loaded.  Alternatively, the "MAGIC"
// environment variable can be used to name any desired Magic
// database files to be loaded, but it must be set prior to
// calling this function for it to take effect.  Remember
// to call Close in order to release initialized resources
// and close currently open Magic library.  Alternatively,
// use Open which will ensure that Close is called once
// the closure finishes.
func New(files ...string) (*Magic, error) {
	mgc, err := open()
	if err != nil {
		return nil, err
	}

	if _, err := mgc.Load(files...); err != nil {
		return nil, err
	}
	return mgc, nil
}

// Close releases all initialized resources and closes
// currently open Magic library.
func (mgc *Magic) Close() {
	mgc.Lock()
	defer mgc.Unlock()
	mgc.magic.close()
}

// String returns a string representation of the Magic type.
func (mgc *Magic) String() string {
	return fmt.Sprintf("Magic{flags:%d path:%s cookie:%p}",
		mgc.flags, mgc.path, mgc.cookie)
}

// Path returns a slice containing fully-qualified path for each
// of Magic database files that was loaded and is currently in use.
// If the "MAGIC" environment variable is present, then each path
// from it will be taken into the account and the value that this
// function returns will be updated accordingly.
func (mgc *Magic) Path() ([]string, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return []string{}, mgc.error()
	}

	// Respect the "MAGIC" environment variable, if present.
	if len(mgc.path) > 0 && os.Getenv("MAGIC") == "" {
		return mgc.path, nil
	}
	rv := C.GoString(C.magic_getpath_wrapper())
	mgc.path = strings.Split(rv, ":")
	return mgc.path, nil
}

// Flags returns a value (bitmask) representing current flags set.
func (mgc *Magic) Flags() (int, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return -1, mgc.error()
	}
	return mgc.flags, nil
}

// FlagsSlice returns a slice containing each distinct flag that
// is currently set and included as a part of the current value
// (bitmask) of flags.  Results are sorted in an ascending order.
func (mgc *Magic) FlagsSlice() ([]int, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return []int{}, mgc.error()
	}

	if mgc.flags == 0 {
		return []int{0}, nil
	}

	flags := make([]int, 0)

	n := 0
	for i := mgc.flags; i > 0; i = i - n {
		n = int(math.Log2(float64(i)))
		n = int(math.Pow(2, float64(n)))
		flags = append(flags, n)
	}
	sort.Ints(flags)
	return flags, nil
}

// SetFlags sets the flags to the new value (bitmask).  Depending
// on which flags are current set the results and/or behavior of
// the Magic library will change accordingly.
func (mgc *Magic) SetFlags(flags int) error {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return mgc.error()
	}

	rv, err := C.magic_setflags_wrapper(mgc.cookie, C.int(flags))
	if rv < 0 && err != nil {
		errno := err.(syscall.Errno)
		switch errno {
		case syscall.EINVAL:
			return &MagicError{int(errno), "unknown or invalid flag specified"}
		case syscall.ENOSYS:
			return &MagicError{int(errno), "flag is not implemented"}
		default:
			return mgc.error()
		}
	}

	mgc.flags = flags
	return nil
}

// Load
func (mgc *Magic) Load(files ...string) (bool, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return false, mgc.error()
	}

	var cfiles *C.char
	defer C.free(unsafe.Pointer(cfiles))

	if len(files) > 0 {
		cfiles = C.CString(strings.Join(files, ":"))
	} else {
		cfiles = C.magic_getpath_wrapper()
	}

	if rv := C.magic_load_wrapper(mgc.cookie, cfiles, C.int(mgc.flags)); rv < 0 {
		return false, mgc.error()
	}
	mgc.path = strings.Split(C.GoString(cfiles), ":")
	return true, nil
}

// Compile
func (mgc *Magic) Compile(files ...string) (bool, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return false, mgc.error()
	}

	var cfiles *C.char
	if len(files) > 0 {
		cfiles = C.CString(strings.Join(files, ":"))
		defer C.free(unsafe.Pointer(cfiles))
	}

	if rv := C.magic_compile_wrapper(mgc.cookie, cfiles, C.int(mgc.flags)); rv < 0 {
		return false, mgc.error()
	}
	return true, nil
}

// Check
func (mgc *Magic) Check(files ...string) (bool, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return false, mgc.error()
	}

	var cfiles *C.char
	if len(files) > 0 {
		cfiles = C.CString(strings.Join(files, ":"))
		defer C.free(unsafe.Pointer(cfiles))
	}

	if rv := C.magic_check_wrapper(mgc.cookie, cfiles, C.int(mgc.flags)); rv < 0 {
		return false, mgc.error()
	}
	return true, nil
}

// File
func (mgc *Magic) File(filename string) (string, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return "", mgc.error()
	}

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	cstring := C.magic_file_wrapper(mgc.cookie, cfilename, C.int(mgc.flags))
	if cstring == nil {
		rv, _ := Version()

		// Handle the case when the "ERROR" flag is set.
		if mgc.flags&ERROR != 0 {
			return "", mgc.error()
		} else if rv < 515 {
			C.magic_errno(mgc.cookie)
			cstring = C.magic_error(mgc.cookie)
		}
	}

	if cstring == nil {
		return "", &MagicError{-1, "unknown result or nil pointer"}
	}

	s := C.GoString(cstring)
	if s == "" || s == "(null)" {
		return "", &MagicError{-1, "empty or invalid result"}
	}
	return s, nil
}

// Buffer
func (mgc *Magic) Buffer(buffer []byte) (string, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return "", mgc.error()
	}

	p, length := unsafe.Pointer(&buffer[0]), C.size_t(len(buffer))

	cstring := C.magic_buffer_wrapper(mgc.cookie, p, length, C.int(mgc.flags))
	if cstring == nil {
		return "", mgc.error()
	}
	return C.GoString(cstring), nil
}

// Descriptor
func (mgc *Magic) Descriptor(fd uintptr) (string, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return "", mgc.error()
	}

	cstring := C.magic_descriptor_wrapper(mgc.cookie, C.int(fd), C.int(mgc.flags))
	if cstring == nil {
		return "", mgc.error()
	}
	return C.GoString(cstring), nil
}

func (mgc *Magic) error() *MagicError {
	if mgc.cookie == nil {
		errno := syscall.EFAULT
		return &MagicError{int(errno), "Magic library is not open"}
	}

	cstring := C.magic_error(mgc.cookie)
	if cstring != nil {
		s := C.GoString(cstring)
		if s == "" || s == "(null)" {
			return &MagicError{-1, "empty or invalid error message"}
		} else {
			errno := int(C.magic_errno(mgc.cookie))
			return &MagicError{errno, s}
		}
	}
	return &MagicError{-1, "unknown error"}
}

func (mgc *Magic) destroy() {
	mgc.Close()
}

func open() (*Magic, error) {
	rv := C.magic_open(C.int(NONE))
	if rv == nil {
		errno := syscall.EPERM
		return nil, &MagicError{int(errno), "failed to initialize Magic library"}
	}

	mgc := &Magic{&magic{flags: NONE, cookie: rv}}
	runtime.SetFinalizer(mgc.magic, (*magic).close)
	return mgc, nil
}

// Open
func Open(f func(magic *Magic) error, files ...string) (err error) {
	var ok bool

	if f == nil || reflect.TypeOf(f).Kind() != reflect.Func {
		return &MagicError{-1, "not a function or nil pointer"}
	}

	mgc, err := New(files...)
	if err != nil {
		return err
	}
	defer mgc.Close()

	defer func() {
		if r := recover(); r != nil {
			err, ok = r.(error)
			if !ok {
				err = &MagicError{-1, fmt.Sprintf("%v", r)}
			}
		}
	}()

	return f(mgc)
}

// Compile
func Compile(files ...string) (bool, error) {
	mgc, err := open()
	if err != nil {
		return false, err
	}
	defer mgc.Close()

	rv, err := mgc.Compile(files...)
	if err != nil {
		return rv, err
	}
	return rv, nil
}

// Check
func Check(files ...string) (bool, error) {
	mgc, err := open()
	if err != nil {
		return false, err
	}
	defer mgc.Close()

	rv, err := mgc.Check(files...)
	if err != nil {
		return rv, err
	}
	return rv, nil
}

// Version returns the underlying Magic library version as in integer
// value in the format "XYY", where X is the major version and Y is
// the minor version number.
func Version() (int, error) {
	rv, err := C.magic_version_wrapper()
	if rv < 0 && err != nil {
		errno := syscall.ENOSYS
		return -1, &MagicError{int(errno), "function is not implemented"}
	}
	return int(rv), nil
}

// VersionString returns the underlying Magic library version as string
// in the format "X.YY".
func VersionString() (string, error) {
	rv, err := Version()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.%02d", rv/100, rv%100), nil
}

// VersionSlice returns a slice containing values of both the major
// and minor version numbers separated from one another.
func VersionSlice() ([]int, error) {
	rv, err := Version()
	if err != nil {
		return []int{}, err
	}
	return []int{rv / 100, rv % 100}, nil
}

// FileMime
func FileMime(filename string, files ...string) (string, error) {
	mgc, err := New(files...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	mgc.SetFlags(MIME)
	return mgc.File(filename)
}

// FileType
func FileType(filename string, files ...string) (string, error) {
	mgc, err := New(files...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	mgc.SetFlags(MIME_TYPE)
	return mgc.File(filename)
}

// FileEncoding
func FileEncoding(filename string, files ...string) (string, error) {
	mgc, err := New(files...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	mgc.SetFlags(MIME_ENCODING)
	return mgc.File(filename)
}

// BufferMime returns MIME identification (both the MIME type
// and MIME encoding), rather than a textual description,
// for the content of the buffer.
func BufferMime(buffer []byte, files ...string) (string, error) {
	mgc, err := New(files...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	mgc.SetFlags(MIME)
	return mgc.Buffer(buffer)
}

// BufferType returns MIME type only, rather than a textual
// description, for the content of the buffer.
func BufferType(buffer []byte, files ...string) (string, error) {
	mgc, err := New(files...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	mgc.SetFlags(MIME_TYPE)
	return mgc.Buffer(buffer)
}

// BufferEncoding returns MIME encoding only, rather than a textual
// description, for the content of the buffer.
func BufferEncoding(buffer []byte, files ...string) (string, error) {
	mgc, err := New(files...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	mgc.SetFlags(MIME_ENCODING)
	return mgc.Buffer(buffer)
}
