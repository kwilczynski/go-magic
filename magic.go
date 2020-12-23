// +build !windows

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

type Option func(*Magic) error

func DisableAutoload(mgc *Magic) error {
	mgc.autoload = false
	return nil
}

func Load(files ...string) Option {
	return func(mgc *Magic) error {
		DisableAutoload(mgc)
		_, err := mgc.Load(files...)
		return err
	}
}

func LoadBuffers(buffers ...[]byte) Option {
	return func(mgc *Magic) error {
		DisableAutoload(mgc)
		_, err := mgc.LoadBuffers(buffers...)
		return err
	}
}

type magic struct {
	sync.Mutex

	flags  int       // Current flags set (bitmask).
	paths  []string  // List of Magic database files currently in-use.
	cookie C.magic_t // Magic database session cookie (a "magic_set" struct on the C side).

	autoload bool // Enables autoloading of Magic database files.
}

// open opens and initializes underlying Magic library and sets the
// finalizer on the object accordingly.
func open() (*Magic, error) {
	// Can only fail allocating memory in this particular case.
	cMagic := C.magic_open_wrapper(C.int(NONE))
	if cMagic == nil {
		errno := syscall.ENOMEM
		return nil, &Error{int(errno), "failed to initialize Magic library"}
	}

	mgc := &Magic{&magic{flags: NONE, cookie: cMagic, autoload: true}}
	runtime.SetFinalizer(mgc.magic, (*magic).close)

	return mgc, nil
}

// close closes underlying Magic library and clears finalizer currently
// set on the object.
func (m *magic) close() {
	if m != nil && m.cookie != nil {
		// This will free resources on the Magic library side.
		C.magic_close_wrapper(m.cookie)
		m.paths = []string{}
		m.cookie = nil
	}
	runtime.SetFinalizer(m, nil)
}

// Magic represents the underlying Magic library.
type Magic struct {
	*magic
}

// New opens and initializes Magic library.
//
// Optionally, a multiple distinct Magic database files can
// be provided to load, otherwise a default database (usually
// available system-wide) will be loaded.
//
// Alternatively, the "MAGIC" environment variable can be used
// to name any desired Magic database files to be loaded, but
// it must be set prior to calling this function for it to take
// effect.
//
// Remember to call Close to release initialized resources
// and close currently opened Magic library, or use Open which
// will ensure that Close is called once the closure finishes.
//
// If there is an error originating from the underlying Magic
// library, it will be of type *Error.
func New(options ...Option) (*Magic, error) {
	mgc, err := open()
	if err != nil {
		return nil, err
	}

	for _, option := range options {
		if err := option(mgc); err != nil {
			mgc.close()
			return nil, err
		}
	}

	if mgc.autoload {
		if _, err := mgc.Load(); err != nil {
			return nil, err
		}
	}

	return mgc, nil
}

// Close releases all initialized resources and closes
// currently open Magic library.
func (mgc *Magic) Close() {
	mgc.Lock()
	defer mgc.Unlock()

	mgc.close()
}

// IsClosed returns true if the underlying Magic library has
// been closed, or false otherwise.
func (mgc *Magic) IsClosed() bool {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc != nil && mgc.cookie != nil {
		return false
	}

	return true
}

// String returns a string representation of the Magic type.
func (mgc *Magic) String() string {
	mgc.Lock()
	defer mgc.Unlock()

	open := false
	if mgc != nil && mgc.cookie != nil {
		open = true
	}

	s := fmt.Sprintf("Magic{flags:%d paths:%v cookie:%p open:%t}", mgc.flags, mgc.paths, mgc.cookie, open)

	return s
}

// Paths returns a slice containing fully-qualified path for each
// of Magic database files that was loaded and is currently in use.
//
// Optionally, if the "MAGIC" environment variable is present,
// then each path from it will be taken into the account and the
// value that this function returns will be updated accordingly.
//
// If there is an error, it will be of type *Error.
func (mgc *Magic) Paths() ([]string, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return []string{}, mgc.error()
	}

	// Respect the "MAGIC" environment variable, if present.
	if len(mgc.paths) > 0 && os.Getenv("MAGIC") == "" {
		return mgc.paths, nil
	}

	paths := C.GoString(C.magic_getpath_wrapper())
	mgc.paths = strings.Split(paths, ":")

	return mgc.paths, nil
}

// Parameter -
//
// If there is an error, it will be of type *Error.
func (mgc *Magic) Parameter(parameter int) (int, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return -1, mgc.error()
	}

	var value int
	p := unsafe.Pointer(&value)

	cResult, err := C.magic_getparam_wrapper(mgc.cookie, C.int(parameter), p)
	if cResult < 0 && err != nil {
		errno := err.(syscall.Errno)
		if errno == syscall.EINVAL {
			return -1, &Error{int(errno), "unknown or invalid parameter specified"}
		}
		return -1, mgc.error()
	}

	return value, nil
}

// SetParameter -
//
// If there is an error, it will be of type *Error.
func (mgc *Magic) SetParameter(parameter int, value int) error {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return mgc.error()
	}

	p := unsafe.Pointer(&value)

	cResult, err := C.magic_setparam_wrapper(mgc.cookie, C.int(parameter), p)
	if cResult < 0 && err != nil {
		errno := err.(syscall.Errno)
		switch errno {
		case syscall.EINVAL:
			return &Error{int(errno), "unknown or invalid parameter specified"}
		case syscall.EOVERFLOW:
			return &Error{int(errno), "invalid parameter value specified"}
		default:
			return mgc.error()
		}
	}

	return nil
}

// Flags returns a value (bitmask) representing current flags set.
//
// If there is an error, it will be of type *Error.
func (mgc *Magic) Flags() (int, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return -1, mgc.error()
	}

	var cRv C.int

	if cRv = C.magic_getflags_wrapper(mgc.cookie); cRv < 0 {
		return -1, mgc.error()
	}
	mgc.flags = int(cRv)

	return mgc.flags, nil
}

// SetFlags sets the flags to the new value (bitmask).
//
// Depending on which flags are current set the results and/or
// behavior of the Magic library will change accordingly.
//
// If there is an error, it will be of type *Error.
func (mgc *Magic) SetFlags(flags int) error {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return mgc.error()
	}

	cResult, err := C.magic_setflags_wrapper(mgc.cookie, C.int(flags))
	if cResult < 0 && err != nil {
		errno := err.(syscall.Errno)
		if errno == syscall.EINVAL {
			return &Error{int(errno), "unknown or invalid flag specified"}
		}
		return mgc.error()
	}
	mgc.flags = flags

	return nil
}

// FlagsSlice returns a slice containing each distinct flag that
// is currently set and included as a part of the current value
// (bitmask) of flags.
//
// Results are sorted in an ascending order.
//
// If there is an error, it will be of type *Error.
func (mgc *Magic) FlagsSlice() ([]int, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return []int{}, mgc.error()
	}

	if mgc.flags == 0 {
		return []int{0}, nil
	}

	var n int
	var flags []int

	// Split current value (bitmask) into a list
	// of distinct flags (bits) currently set.
	for i := mgc.flags; i > 0; i -= n {
		n = int(math.Log2(float64(i)))
		n = int(math.Pow(2, float64(n)))
		flags = append(flags, n)
	}
	sort.Ints(flags)

	return flags, nil
}

// Load -
//
// If there is an error, it will be of type *Error.
func (mgc *Magic) Load(files ...string) (bool, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return false, mgc.error()
	}

	var cFiles *C.char
	defer C.free(unsafe.Pointer(cFiles))

	cFlags := C.int(mgc.flags)

	// Assemble the list of custom Magic files into a colon-separated
	// list that is required by the underlying Magic library, otherwise
	// defer to the default list of paths provided by the Magic library.
	if len(files) > 0 {
		cFiles = C.CString(strings.Join(files, ":"))
	} else {
		cFiles = C.magic_getpath_wrapper()
	}

	if cResult := C.magic_load_wrapper(mgc.cookie, cFiles, cFlags); cResult < 0 {
		return false, mgc.error()
	}
	mgc.paths = strings.Split(C.GoString(cFiles), ":")

	return true, nil
}

// LoadBuffers -
//
// If there is an error, it will be of type *Error.
func (mgc *Magic) LoadBuffers(buffers ...[]byte) (bool, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return false, mgc.error()
	}

	cSize := C.size_t(len(buffers))
	cPointers := make([]uintptr, cSize)
	cSizes := make([]C.size_t, cSize)

	cFlags := C.int(mgc.flags)

	for i := range buffers {
		cPointers[i] = uintptr(unsafe.Pointer(&buffers[i][0]))
		cSizes[i] = C.size_t(len(buffers[i]))
	}

	p := (*unsafe.Pointer)(unsafe.Pointer(&cPointers[0]))
	s := (*C.size_t)(unsafe.Pointer(&cSizes[0]))

	if cResult := C.magic_load_buffers_wrapper(mgc.cookie, p, s, cSize, cFlags); cResult < 0 {
		return false, mgc.error()
	}

	return true, nil
}

// Compile -
//
// If there is an error, it will be of type *Error.
func (mgc *Magic) Compile(files ...string) (bool, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return false, mgc.error()
	}

	var cFiles *C.char
	defer C.free(unsafe.Pointer(cFiles))

	cFlags := C.int(mgc.flags)

	if len(files) > 0 {
		cFiles = C.CString(strings.Join(files, ":"))
	}

	if cResult := C.magic_compile_wrapper(mgc.cookie, cFiles, cFlags); cResult < 0 {
		return false, mgc.error()
	}

	return true, nil
}

// Check -
//
// If there is an error, it will be of type *Error.
func (mgc *Magic) Check(files ...string) (bool, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return false, mgc.error()
	}

	var cFiles *C.char
	defer C.free(unsafe.Pointer(cFiles))

	cFlags := C.int(mgc.flags)

	if len(files) > 0 {
		cFiles = C.CString(strings.Join(files, ":"))
	}

	if cRv := C.magic_check_wrapper(mgc.cookie, cFiles, cFlags); cRv < 0 {
		return false, mgc.error()
	}

	return true, nil
}

// File -
//
// If there is an error, it will be of type *Error.
func (mgc *Magic) File(filename string) (string, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return "", mgc.error()
	}

	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	cString := C.magic_file_wrapper(mgc.cookie, cFilename, C.int(mgc.flags))
	if cString == nil {
		version := Version()
		// Handle the case when the "ERROR" flag is set regardless
		// of the current version of the underlying Magic library.
		//
		// Prior to version 5.15 the correct behaviour that concerns
		// the following IEEE 1003.1 standards was broken:
		//
		//   http://pubs.opengroup.org/onlinepubs/007904975/utilities/file.html
		//   http://pubs.opengroup.org/onlinepubs/9699919799/utilities/file.html
		//
		// This is an attempt to mitigate the problem and correct
		// it to achieve the desired behaviour as per the standards.
		if mgc.flags&ERROR != 0 {
			return "", mgc.error()
		} else if version < 515 || mgc.flags&EXTENSION != 0 {
			C.magic_errno_wrapper(mgc.cookie)
			cString = C.magic_error_wrapper(mgc.cookie)
		}
	}

	// This case should not happen, ever.
	if cString == nil {
		return "", &Error{-1, "unknown result or nil pointer"}
	}

	// Depending on the version of the underlying
	// Magic library the magic_file() function can
	// fail and either yield no results or return
	// the "(null)" string instead.  Often this
	// would indicate that an older version of
	// the Magic library is in use.
	s := C.GoString(cString)
	if s == "" || s == "(null)" {
		return "", &Error{-1, "empty or invalid result"}
	}

	return s, nil
}

// Buffer -
//
// If there is an error, it will be of type *Error.
func (mgc *Magic) Buffer(buffer []byte) (string, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return "", mgc.error()
	}

	cFlags := C.int(mgc.flags)

	p := unsafe.Pointer(&buffer[0])
	cSize := C.size_t(len(buffer))

	cString := C.magic_buffer_wrapper(mgc.cookie, p, cSize, cFlags)
	if cString == nil {
		return "", mgc.error()
	}

	return C.GoString(cString), nil
}

// Descriptor -
//
// If there is an error, it will be of type *Error.
func (mgc *Magic) Descriptor(fd uintptr) (string, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return "", mgc.error()
	}

	cFd := C.int(fd)

	cFlags := C.int(mgc.flags)

	cString, err := C.magic_descriptor_wrapper(mgc.cookie, cFd, cFlags)
	if cString == nil && err != nil {
		errno := err.(syscall.Errno)
		if errno == syscall.EBADF {
			return "", &Error{int(errno), "bad file descriptor"}
		}
		return "", mgc.error()
	}

	return C.GoString(cString), nil
}

// error retrieves an error from the underlying Magic library.
func (mgc *Magic) error() *Error {
	if mgc.cookie == nil {
		errno := syscall.EFAULT
		return &Error{int(errno), "Magic library is not open"}
	}

	cString := C.magic_error_wrapper(mgc.cookie)
	if cString != nil {
		// Depending on the version of the underlying
		// Magic library, the error reporting facilities
		// can fail and either yield no results or return
		// the "(null)" string instead.  Often this would
		// indicate that an older version of the Magic
		// library is in use.
		s := C.GoString(cString)
		if s == "" || s == "(null)" {
			return &Error{-1, "empty or invalid error message"}
		}
		errno := int(C.magic_errno_wrapper(mgc.cookie))
		return &Error{errno, s}
	}

	return &Error{-1, "an unknown error has occurred"}
}

// Open -
//
// If there is an error, it will be of type *Error.
func Open(f func(magic *Magic) error, options ...Option) (err error) {
	var ok bool

	if f == nil || reflect.TypeOf(f).Kind() != reflect.Func {
		return &Error{-1, "not a function or nil pointer"}
	}

	mgc, err := New(options...)
	if err != nil {
		return err
	}
	defer mgc.Close()

	// Make sure to return a proper error should there
	// be any failure originating from within the closure.
	defer func() {
		if r := recover(); r != nil {
			err, ok = r.(error)
			if !ok {
				err = &Error{-1, fmt.Sprintf("%v", r)}
			}
		}
	}()

	return f(mgc)
}

// Compile -
//
// If there is an error, it will be of type *Error.
func Compile(files ...string) (bool, error) {
	mgc, err := open()
	if err != nil {
		return false, err
	}
	defer mgc.Close()

	result, err := mgc.Compile(files...)
	if err != nil {
		return result, err
	}

	return result, nil
}

// Check -
//
// If there is an error, it will be of type *Error.
func Check(files ...string) (bool, error) {
	mgc, err := open()
	if err != nil {
		return false, err
	}
	defer mgc.Close()

	result, err := mgc.Check(files...)
	if err != nil {
		return result, err
	}

	return result, nil
}

// Version returns the underlying Magic library version as an integer
// value in the format "XYY", where X is the major version and Y is
// the minor version number.
func Version() int {
	cRv := C.magic_version_wrapper()
	return int(cRv)
}

// VersionString returns the underlying Magic library version
// as string in the format "X.YY".
//
// If there is an error, it will be of type *Error.
func VersionString() string {
	version := Version()

	s := fmt.Sprintf("%d.%02d", version/100, version%100)

	return s
}

// VersionSlice returns a slice containing values of both the
// major and minor version numbers separated from one another.
//
// If there is an error, it will be of type *Error.
func VersionSlice() []int {
	version := Version()

	s := []int{version / 100, version % 100}

	return s
}

// FileMime returns MIME identification (both the MIME type
// and MIME encoding), rather than a textual description,
// for the named file.
//
// If there is an error, it will be of type *Error.
func FileMime(name string, files ...string) (string, error) {
	mgc, err := New(Load(files...))
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	if err := mgc.SetFlags(MIME); err != nil {
		return "", err
	}

	return mgc.File(name)
}

// FileType returns MIME type only, rather than a textual
// description, for the named file.
//
// If there is an error, it will be of type *Error.
func FileType(name string, files ...string) (string, error) {
	mgc, err := New(Load(files...))
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	if err := mgc.SetFlags(MIME_TYPE); err != nil {
		return "", err
	}

	return mgc.File(name)
}

// FileEncoding returns MIME encoding only, rather than a textual
// description, for the content of the buffer.
//
// If there is an error, it will be of type *Error.
func FileEncoding(name string, files ...string) (string, error) {
	mgc, err := New(Load(files...))
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	if err := mgc.SetFlags(MIME_ENCODING); err != nil {
		return "", err
	}

	return mgc.File(name)
}

// BufferMime returns MIME identification (both the MIME type
// and MIME encoding), rather than a textual description,
// for the content of the buffer.
//
// If there is an error, it will be of type *Error.
func BufferMime(buffer []byte, files ...string) (string, error) {
	mgc, err := New(Load(files...))
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	if err := mgc.SetFlags(MIME); err != nil {
		return "", err
	}

	return mgc.Buffer(buffer)
}

// BufferType returns MIME type only, rather than a textual
// description, for the content of the buffer.
//
// If there is an error, it will be of type *Error.
func BufferType(buffer []byte, files ...string) (string, error) {
	mgc, err := New(Load(files...))
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	if err := mgc.SetFlags(MIME_TYPE); err != nil {
		return "", err
	}

	return mgc.Buffer(buffer)
}

// BufferEncoding returns MIME encoding only, rather than a textual
// description, for the content of the buffer.
//
// If there is an error, it will be of type *Error.
func BufferEncoding(buffer []byte, files ...string) (string, error) {
	mgc, err := New(Load(files...))
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	if err := mgc.SetFlags(MIME_ENCODING); err != nil {
		return "", err
	}

	return mgc.Buffer(buffer)
}
