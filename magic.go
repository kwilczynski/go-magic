/*
 * magic.go
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

package magic

/*
#cgo LDFLAGS: -lmagic
#include "functions.h"
*/
import "C"

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

type magic struct {
	sync.Mutex
	flags  int
	path   []string
	cookie C.magic_t
}

func (m *magic) close() {
	if m != nil && m.cookie != nil {
		C.magic_close(m.cookie)
		m.cookie = nil
	}
	runtime.SetFinalizer(m, nil)
}

type Magic struct {
	*magic
}

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

func (mgc *Magic) Close() {
	mgc.Lock()
	defer mgc.Unlock()
	mgc.magic.close()
}

func (mgc *Magic) String() string {
	return fmt.Sprintf("Magic{flags:%d path:%s cookie:%p}",
		mgc.flags, mgc.path, mgc.cookie)
}

func (mgc *Magic) Path() ([]string, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return []string{}, mgc.error()
	}

	if len(mgc.path) > 0 && os.Getenv("MAGIC") == "" {
		return mgc.path, nil
	}
	rv := C.GoString(C.magic_getpath_wrapper())
	mgc.path = strings.Split(rv, ":")
	return mgc.path, nil
}

func (mgc *Magic) Flags() (int, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return -1, mgc.error()
	}
	return mgc.flags, nil
}

func (mgc *Magic) SetFlags(flags int) error {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return mgc.error()
	}

	rv, err := C.magic_setflags_wrapper(mgc.cookie, C.int(flags))
	if rv < 0 && err != nil {
		errno := err.(syscall.Errno)
		return &MagicError{int(errno), errno.Error()}
	}

	mgc.flags = flags
	return nil
}

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

	if rv := C.magic_load_wrapper(mgc.cookie, cfiles); rv < 0 {
		return false, mgc.error()
	}
	mgc.path = strings.Split(C.GoString(cfiles), ":")
	return true, nil
}

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

	if rv := C.magic_compile_wrapper(mgc.cookie, cfiles); rv < 0 {
		return false, mgc.error()
	}
	return true, nil
}

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

	if rv := C.magic_check_wrapper(mgc.cookie, cfiles); rv < 0 {
		return false, mgc.error()
	}
	return true, nil
}

func (mgc *Magic) File(filename string) (string, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return "", mgc.error()
	}

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	cstring := C.magic_file(mgc.cookie, cfilename)
	if cstring == nil {
		return "", mgc.error()
	}
	return C.GoString(cstring), nil
}

func (mgc *Magic) Buffer(buffer []byte) (string, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return "", mgc.error()
	}

	p, length := unsafe.Pointer(&buffer[0]), C.size_t(len(buffer))

	cstring := C.magic_buffer(mgc.cookie, p, length)
	if cstring == nil {
		return "", mgc.error()
	}
	return C.GoString(cstring), nil
}

func (mgc *Magic) Descriptor(fd uintptr) (string, error) {
	mgc.Lock()
	defer mgc.Unlock()

	if mgc.cookie == nil {
		return "", mgc.error()
	}

	cstring := C.magic_descriptor(mgc.cookie, C.int(fd))
	if cstring == nil {
		return "", mgc.error()
	}
	return C.GoString(cstring), nil
}

func (mgc *Magic) error() *MagicError {
	if mgc.cookie == nil {
		errno := syscall.EINVAL
		return &MagicError{int(syscall.EINVAL), errno.Error()}
	}

	errno := int(C.magic_errno(mgc.cookie))
	cstring := C.magic_error(mgc.cookie)
	if cstring != nil {
		return &MagicError{errno, C.GoString(cstring)}
	}
	return nil
}

func (mgc *Magic) destroy() {
	mgc.Close()
}

func open() (*Magic, error) {
	rv := C.magic_open(C.int(NONE))
	if rv == nil {
		errno := syscall.ENOMEM
		return nil, &MagicError{int(errno), errno.Error()}
	}

	mgc := &Magic{&magic{flags: NONE, cookie: rv}}
	runtime.SetFinalizer(mgc.magic, (*magic).close)
	return mgc, nil
}

func Open(f func(magic *Magic) error, files ...string) (err error) {
	var ok bool
	errno := syscall.EINVAL

	if f == nil || reflect.TypeOf(f).Kind() != reflect.Func {
		return &MagicError{int(errno), errno.Error()}
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
				err = &MagicError{int(errno), fmt.Sprintf("%v", r)}
			}
		}
	}()

	return f(mgc)
}

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

func Version() (int, error) {
	rv, err := C.magic_version_wrapper()
	if rv < 0 && err != nil {
		errno := err.(syscall.Errno)
		return -1, &MagicError{int(errno), errno.Error()}
	}
	return int(rv), nil
}

func FileMime(filename string, files ...string) (string, error) {
	mgc, err := New(files...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	mgc.SetFlags(MIME)
	return mgc.File(filename)
}

func FileEncoding(filename string, files ...string) (string, error) {
	mgc, err := New(files...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	mgc.SetFlags(MIME_ENCODING)
	return mgc.File(filename)
}

func FileType(filename string, files ...string) (string, error) {
	mgc, err := New(files...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	mgc.SetFlags(MIME_TYPE)
	return mgc.File(filename)
}

func BufferMime(buffer []byte, files ...string) (string, error) {
	mgc, err := New(files...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	mgc.SetFlags(MIME)
	return mgc.Buffer(buffer)
}

func BufferEncoding(buffer []byte, files ...string) (string, error) {
	mgc, err := New(files...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	mgc.SetFlags(MIME_ENCODING)
	return mgc.Buffer(buffer)
}

func BufferType(buffer []byte, files ...string) (string, error) {
	mgc, err := New(files...)
	if err != nil {
		return "", err
	}
	defer mgc.Close()

	mgc.SetFlags(MIME_TYPE)
	return mgc.Buffer(buffer)
}
