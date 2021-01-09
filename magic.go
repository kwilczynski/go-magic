package magic

/*
#cgo LDFLAGS: -lmagic
#cgo !darwin LDFLAGS: -Wl,--as-needed -Wl,--no-undefined
#cgo CFLAGS: -std=c99 -fPIC

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
const Separator string = "\n- "

// Option represents an option that can be set when creating a new object.
type Option func(*Magic) error

// DoNotStopOnErrors
func DoNotStopOnErrors(mgc *Magic) error {
	mgc.errors = false
	return nil
}

// DisableAutoload disables autoloading of the Magic database files when
// creating a new object.
//
// This option can be used to prevent the Magic database files from being
// loaded from the default location on the filesystem so that the Magic
// database can be loaded later manually from a different location using
// the Load function, or from a buffer in memory using the LoadBuffers
// function.
func DisableAutoload(mgc *Magic) error {
	mgc.autoload = false
	return nil
}

// WithFiles
func WithFiles(files ...string) Option {
	return func(mgc *Magic) error {
		return mgc.Load(files...)
	}
}

// WithBuffers
func WithBuffers(buffers ...[]byte) Option {
	return func(mgc *Magic) error {
		return mgc.LoadBuffers(buffers...)
	}
}

type magic struct {
	sync.RWMutex
	// Current flags set (bitmask).
	flags int
	// List of the Magic database files currently in-use.
	paths []string
	// The Magic database session cookie.
	cookie C.magic_t
	// Enable autoloading of the Magic database files.
	autoload bool
	// Enable reporting of I/O-related errors as first class errors.
	errors bool
	// The Magic database has been loaded successfully.
	loaded bool
}

// open opens and initializes the Magic library and sets the finalizer
// on the object.
func open() (*Magic, error) {
	// Can only fail allocating memory in this particular case.
	cMagic := C.magic_open_wrapper(C.int(NONE))
	if cMagic == nil {
		return nil, &Error{int(syscall.ENOMEM), "failed to initialize Magic library"}
	}
	mgc := &Magic{&magic{flags: NONE, cookie: cMagic, autoload: true, errors: true}}
	runtime.SetFinalizer(mgc.magic, (*magic).close)
	return mgc, nil
}

// close closes the Magic library and clears finalizer set on the object.
func (m *magic) close() {
	if m != nil && m.cookie != nil {
		// This will free resources on the Magic library side.
		C.magic_close_wrapper(m.cookie)
		m.paths = []string{}
		m.cookie = nil
	}
	runtime.SetFinalizer(m, nil)
}

// error retrieves an error from the Magic library.
func (m *magic) error() error {
	if cString := C.magic_error_wrapper(m.cookie); cString != nil {
		// Depending on the version of the Magic library,
		// the error reporting facilities can fail and
		// either yield no results or return the "(null)"
		// string instead. Often this would indicate that
		// an older version of the Magic library is in use.
		s := C.GoString(cString)
		if s == "" || s == "(null)" {
			return &Error{-1, "empty or invalid error message"}
		}
		return &Error{int(C.magic_errno_wrapper(m.cookie)), s}
	}
	return &Error{-1, "an unknown error has occurred"}
}

// Magic represents the Magic library.
type Magic struct {
	*magic
}

// New opens and initializes the Magic library.
//
// Optionally, a multiple distinct the Magic database files can
// be provided to load, otherwise a default database (usually
// available system-wide) will be loaded.
//
// Alternatively, the "MAGIC" environment variable can be used
// to name any desired the Magic database files to be loaded, but
// it must be set prior to calling this function for it to take
// effect.
//
// Remember to call Close to release initialized resources
// and close currently opened the Magic library, or use Open
// which will ensure that Close is called once the closure
// finishes.
func New(options ...Option) (*Magic, error) {
	mgc, err := open()
	if err != nil {
		return nil, err
	}

	if s := os.Getenv("MAGIC_DO_NOT_AUTOLOAD"); s != "" {
		mgc.autoload = false
	}
	if s := os.Getenv("MAGIC_DO_NOT_STOP_ON_ERROR"); s != "" {
		mgc.errors = false
	}

	for _, option := range options {
		if err := option(mgc); err != nil {
			mgc.close()
			return nil, err
		}
	}

	if mgc.autoload && !mgc.loaded {
		if err := mgc.Load(); err != nil {
			return nil, err
		}
	}
	return mgc, nil
}

/// Must
func Must(magic *Magic, err error) *Magic {
	if err != nil {
		panic(err)
	}
	return magic
}

// Close releases all initialized resources and closes
// currently open the Magic library.
func (mgc *Magic) Close() {
	mgc.Lock()
	defer mgc.Unlock()
	mgc.close()
}

// IsOpen returns true if the Magic library is currently
// open, or false otherwise.
func (mgc *Magic) IsOpen() bool {
	mgc.RLock()
	defer mgc.RUnlock()
	return verifyOpen(mgc) == nil
}

// IsClosed returns true if the Magic library has
// been closed, or false otherwise.
func (mgc *Magic) IsClosed() bool {
	return !mgc.IsOpen()
}

// HasLoaded returns true if the Magic library has
// been loaded successfully, or false otherwise.
func (mgc *Magic) HasLoaded() bool {
	mgc.RLock()
	defer mgc.RUnlock()
	return verifyLoaded(mgc) == nil
}

// String returns a string representation of the Magic type.
func (mgc *Magic) String() string {
	mgc.RLock()
	defer mgc.RUnlock()
	s := fmt.Sprintf("Magic{flags:%d paths:%v open:%t loaded:%t}", mgc.flags, mgc.paths, mgc.IsOpen(), mgc.HasLoaded())
	return s
}

// Paths returns a slice containing fully-qualified path for each
// of the Magic database files that was loaded and is currently
// in use.
//
// Optionally, if the "MAGIC" environment variable is present,
// then each path from it will be taken into the account and the
// value that this function returns will be updated accordingly.
func (mgc *Magic) Paths() ([]string, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if err := verifyOpen(mgc); err != nil {
		return []string{}, err
	}

	// Respect the "MAGIC" environment variable, if present.
	if len(mgc.paths) > 0 && os.Getenv("MAGIC") == "" {
		return mgc.paths, nil
	}
	paths := C.GoString(C.magic_getpath_wrapper())
	return strings.Split(paths, ":"), nil
}

// Parameter
func (mgc *Magic) Parameter(parameter int) (int, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if err := verifyOpen(mgc); err != nil {
		return -1, err
	}

	var value int
	p := unsafe.Pointer(&value)

	cResult, err := C.magic_getparam_wrapper(mgc.cookie, C.int(parameter), p)
	if cResult < 0 && err != nil {
		if errno := err.(syscall.Errno); errno == syscall.EINVAL {
			return -1, &Error{int(errno), "unknown or invalid parameter specified"}
		}
		return -1, mgc.error()
	}
	return value, nil
}

// SetParameter
func (mgc *Magic) SetParameter(parameter int, value int) error {
	mgc.Lock()
	defer mgc.Unlock()

	if err := verifyOpen(mgc); err != nil {
		return err
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
func (mgc *Magic) Flags() (int, error) {
	mgc.RLock()
	defer mgc.RUnlock()

	if err := verifyOpen(mgc); err != nil {
		return -1, err
	}

	cRv, err := C.magic_getflags_wrapper(mgc.cookie)
	if cRv < 0 && err != nil {
		if err.(syscall.Errno) == syscall.ENOSYS {
			return mgc.flags, nil
		}
		return -1, mgc.error()
	}
	return int(cRv), nil
}

// SetFlags sets the flags to the new value (bitmask).
//
// Depending on which flags are current set the results and/or
// behavior of the Magic library will change accordingly.
func (mgc *Magic) SetFlags(flags int) error {
	mgc.Lock()
	defer mgc.Unlock()

	if err := verifyOpen(mgc); err != nil {
		return err
	}

	cResult, err := C.magic_setflags_wrapper(mgc.cookie, C.int(flags))
	if cResult < 0 && err != nil {
		if errno := err.(syscall.Errno); errno == syscall.EINVAL {
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
func (mgc *Magic) FlagsSlice() ([]int, error) {
	mgc.RLock()
	defer mgc.RUnlock()

	if err := verifyOpen(mgc); err != nil {
		return []int{}, err
	}
	if mgc.flags == 0 {
		return []int{0}, nil
	}

	var (
		n     int
		flags []int
	)

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

// Load
func (mgc *Magic) Load(files ...string) error {
	mgc.Lock()
	defer mgc.Unlock()

	if err := verifyOpen(mgc); err != nil {
		return err
	}
	// Clear paths. To be set again when the Magic
	// database files are successfully loaded.
	mgc.paths = []string{}

	var cFiles *C.char
	defer C.free(unsafe.Pointer(cFiles))

	// Assemble the list of custom database Magic files into a
	// colon-separated list that is required by the Magic library,
	// otherwise defer to the default list of paths provided by
	// the Magic library.
	if len(files) > 0 {
		cFiles = C.CString(strings.Join(files, ":"))
	} else {
		cFiles = C.magic_getpath_wrapper()
	}

	if cRv := C.magic_load_wrapper(mgc.cookie, cFiles, C.int(mgc.flags)); cRv < 0 {
		mgc.loaded = false
		return mgc.error()
	}
	mgc.loaded = true
	mgc.paths = strings.Split(C.GoString(cFiles), ":")
	return nil
}

// LoadBuffers
func (mgc *Magic) LoadBuffers(buffers ...[]byte) error {
	mgc.Lock()
	defer mgc.Unlock()

	if err := verifyOpen(mgc); err != nil {
		return err
	}

	var (
		empty []byte
		p     *unsafe.Pointer
		s     *C.size_t
	)
	// Clear paths. To be set again when the Magic
	// database files are successfully loaded.
	mgc.paths = []string{}

	cSize := C.size_t(len(buffers))
	cPointers := make([]uintptr, cSize)
	cSizes := make([]C.size_t, cSize)

	for i := range buffers {
		// An attempt to load the Magic database from a number of
		// buffers in memory where a single buffer is empty would
		// result in a failure.
		cPointers[i] = uintptr(unsafe.Pointer(&empty))
		if s := len(buffers[i]); s > 0 {
			cPointers[i] = uintptr(unsafe.Pointer(&buffers[i][0]))
			cSizes[i] = C.size_t(s)
		}
	}

	if cSize > 0 {
		p = (*unsafe.Pointer)(unsafe.Pointer(&cPointers[0]))
		s = (*C.size_t)(unsafe.Pointer(&cSizes[0]))
	}

	if cRv := C.magic_load_buffers_wrapper(mgc.cookie, p, s, cSize, C.int(mgc.flags)); cRv < 0 {
		mgc.loaded = false
		// Loading a compiled Magic database from a buffer in memory can
		// often cause failure, sadly there isn't a proper error messages
		// in some of the cases, thus the assumtion is that it failed
		// at it couldn't be loaded, whatever the reason.
		if cString := C.magic_error_wrapper(mgc.cookie); cString != nil {
			return &Error{-1, C.GoString(cString)}
		}
		return &Error{-1, "unable to load Magic database"}
	}
	mgc.loaded = true
	return nil
}

// Compile
func (mgc *Magic) Compile(file string) error {
	mgc.RLock()
	defer mgc.RUnlock()

	if err := verifyOpen(mgc); err != nil {
		return err
	}

	cFile := C.CString(file)
	defer C.free(unsafe.Pointer(cFile))

	if cRv := C.magic_compile_wrapper(mgc.cookie, cFile, C.int(mgc.flags)); cRv < 0 {
		return mgc.error()
	}
	return nil
}

// Check
func (mgc *Magic) Check(file string) (bool, error) {
	mgc.RLock()
	defer mgc.RUnlock()

	if err := verifyOpen(mgc); err != nil {
		return false, err
	}

	cFile := C.CString(file)
	defer C.free(unsafe.Pointer(cFile))

	if cRv := C.magic_check_wrapper(mgc.cookie, cFile, C.int(mgc.flags)); cRv < 0 {
		return false, mgc.error()
	}
	return true, nil
}

// File
func (mgc *Magic) File(file string) (string, error) {
	mgc.RLock()
	defer mgc.RUnlock()

	if err := verifyOpen(mgc); err != nil {
		return "", err
	}
	if err := verifyLoaded(mgc); err != nil {
		return "", err
	}

	cFile := C.CString(file)
	defer C.free(unsafe.Pointer(cFile))

	var cString *C.char

	flagsSaveAndRestore(mgc, func() {
		cString = C.magic_file_wrapper(mgc.cookie, cFile, C.int(mgc.flags))
	})
	if cString == nil {
		// Handle the case when the "ERROR" flag is set regardless
		// of the current version of the Magic library.
		//
		// Prior to version 5.15 the correct behavior that concerns
		// the following IEEE 1003.1 standards was broken:
		//
		//   http://pubs.opengroup.org/onlinepubs/007904975/utilities/file.html
		//   http://pubs.opengroup.org/onlinepubs/9699919799/utilities/file.html
		//
		// This is an attempt to mitigate the problem and correct
		// it to achieve the desired behavior as per the standards.
		if mgc.errors || mgc.flags&ERROR != 0 {
			return "", mgc.error()
		}
		cString = C.magic_error_wrapper(mgc.cookie)
	}
	return errorOrString(mgc, cString)
}

// Buffer
func (mgc *Magic) Buffer(buffer []byte) (string, error) {
	mgc.RLock()
	defer mgc.RUnlock()

	if err := verifyOpen(mgc); err != nil {
		return "", err
	}
	if err := verifyLoaded(mgc); err != nil {
		return "", err
	}

	var (
		cString *C.char
		p       unsafe.Pointer
	)

	cSize := C.size_t(len(buffer))
	if cSize > 0 {
		p = unsafe.Pointer(&buffer[0])
	}

	flagsSaveAndRestore(mgc, func() {
		cString = C.magic_buffer_wrapper(mgc.cookie, p, cSize, C.int(mgc.flags))
	})
	return errorOrString(mgc, cString)
}

// Descriptor
func (mgc *Magic) Descriptor(fd uintptr) (string, error) {
	mgc.RLock()
	defer mgc.RUnlock()

	if err := verifyOpen(mgc); err != nil {
		return "", err
	}
	if err := verifyLoaded(mgc); err != nil {
		return "", err
	}

	var (
		err     error
		cString *C.char
	)

	flagsSaveAndRestore(mgc, func() {
		cString, err = C.magic_descriptor_wrapper(mgc.cookie, C.int(fd), C.int(mgc.flags))
	})
	if err != nil {
		if errno := err.(syscall.Errno); errno == syscall.EBADF {
			return "", &Error{int(errno), "bad file descriptor"}
		}
	}
	return errorOrString(mgc, cString)
}

// Open
func Open(f func(*Magic) error, options ...Option) (err error) {
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

// Compile
func Compile(file string) error {
	mgc, err := open()
	if err != nil {
		return err
	}
	defer mgc.close()
	return mgc.Compile(file)
}

// Check
func Check(file string) (bool, error) {
	mgc, err := open()
	if err != nil {
		return false, err
	}
	defer mgc.close()
	return mgc.Check(file)
}

// Version returns the Magic library version as an integer
// value in the format "XYY", where X is the major version
// and Y is the minor version number.
func Version() int {
	return int(C.magic_version_wrapper())
}

// VersionString returns the Magic library version as
// a string in the format "X.YY".
func VersionString() string {
	v := Version()
	return fmt.Sprintf("%d.%02d", v/100, v%100)
}

// VersionSlice returns a slice containing values of both the
// major and minor version numbers separated from one another.
func VersionSlice() []int {
	v := Version()
	return []int{v / 100, v % 100}
}

// FileMime returns MIME identification (both the MIME type
// and MIME encoding), rather than a textual description,
// for the named file.
func FileMime(file string, options ...Option) (string, error) {
	mgc, err := New(options...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	if err := mgc.SetFlags(MIME); err != nil {
		return "", err
	}
	return mgc.File(file)
}

// FileType returns MIME type only, rather than a textual
// description, for the named file.
func FileType(file string, options ...Option) (string, error) {
	mgc, err := New(options...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	if err := mgc.SetFlags(MIME_TYPE); err != nil {
		return "", err
	}
	return mgc.File(file)
}

// FileEncoding returns MIME encoding only, rather than a textual
// description, for the content of the buffer.
func FileEncoding(file string, options ...Option) (string, error) {
	mgc, err := New(options...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	if err := mgc.SetFlags(MIME_ENCODING); err != nil {
		return "", err
	}
	return mgc.File(file)
}

// BufferMime returns MIME identification (both the MIME type
// and MIME encoding), rather than a textual description,
// for the content of the buffer.
func BufferMime(buffer []byte, options ...Option) (string, error) {
	mgc, err := New(options...)
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
func BufferType(buffer []byte, options ...Option) (string, error) {
	mgc, err := New(options...)
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
func BufferEncoding(buffer []byte, options ...Option) (string, error) {
	mgc, err := New(options...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	if err := mgc.SetFlags(MIME_ENCODING); err != nil {
		return "", err
	}
	return mgc.Buffer(buffer)
}

func verifyOpen(mgc *Magic) error {
	if mgc != nil && mgc.cookie != nil {
		return nil
	}
	return &Error{int(syscall.EFAULT), "Magic library is not open"}
}

func verifyLoaded(mgc *Magic) error {
	// Magic database can only ever be loaded
	// if the Magic library is currently open.
	if err := verifyOpen(mgc); err == nil && mgc.loaded {
		return nil
	}
	return &Error{-1, "Magic database not loaded"}
}

func flagsSaveAndRestore(mgc *Magic, f func()) {
	var flags int

	flags, mgc.flags = mgc.flags, mgc.flags|RAW
	// Make sure to set the "ERROR" flag so that any
	// I/O-related errors will become first class
	// errors reported back by the Magic library.
	if mgc.errors {
		mgc.flags |= ERROR
	}

	ok := mgc.flags&CONTINUE != 0 || mgc.flags&ERROR != 0
	if ok {
		C.magic_setflags_wrapper(mgc.cookie, C.int(mgc.flags))
	}
	defer func() {
		if ok && flags > 0 {
			C.magic_setflags_wrapper(mgc.cookie, C.int(mgc.flags))
		}
	}()
	mgc.flags = flags
	f()
}

func errorOrString(mgc *Magic, cString *C.char) (string, error) {
	if cString == nil {
		return "", &Error{-1, "unknown result or nil pointer"}
	}
	s := C.GoString(cString)
	if s != "" {
		return s, nil
	}
	if s == "???" || s == "(null)" {
		// The Magic flag that support primarily files e.g.,
		// MAGIC_EXTENSION, etc., would not return a meaningful
		// value for directories and special files, and such.
		// Thus, it's better to return an empty string to
		// indicate lack of results, rather than a confusing
		// string consisting of three questions marks.
		if mgc.flags&EXTENSION != 0 {
			return "", nil
		}
		// Depending on the version of the Magic library
		// the magic_file() function can fail and either
		// yield no results or return the "(null)" string
		// instead. Often this would indicate that an
		// older version of the Magic library is in use.
		return "", &Error{-1, "empty or invalid result"}
	}
	return "", mgc.error()
}
