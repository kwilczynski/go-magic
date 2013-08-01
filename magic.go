// +build linux,cgo darwin,cgo !windows
// +build 386 amd64

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
	"runtime"
	"strings"
	"sync"
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

func New() *Magic {
	mgc := &Magic{&magic{flags: NONE, cookie: C.magic_open(C.int(NONE))}}
	runtime.SetFinalizer(mgc.magic, (*magic).close)
	return mgc
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

func (mgc *Magic) Path() []string {
	mgc.Lock()
	defer mgc.Unlock()

	if len(mgc.path) != 0 && os.Getenv("MAGIC") == "" {
		return mgc.path
	}
	rv := C.GoString(C.magic_getpath_wrapper())
	mgc.path = strings.Split(rv, ":")
	return mgc.path
}

func (mgc *Magic) Flags() int {
	mgc.Lock()
	defer mgc.Unlock()
	return mgc.flags
}

func (mgc *Magic) SetFlags(flags int) error {
	mgc.Lock()
	defer mgc.Unlock()

	rv := int(C.magic_setflags_wrapper(mgc.cookie, C.int(flags)))
	if rv < 0 {
		return &MagicError{rv, "invalid argument"}
	}
	mgc.flags = flags
	return nil
}

func (mgc *Magic) Load(files ...string) error {
	mgc.Lock()
	defer mgc.Unlock()

	var cfiles *C.char
	defer C.free(unsafe.Pointer(cfiles))

	if len(files) != 0 {
		cfiles = C.CString(strings.Join(files, ":"))
	} else {
		cfiles = C.magic_getpath_wrapper()
	}

	if rv := C.magic_load_wrapper(mgc.cookie, cfiles); rv < 0 {
		return mgc.error()
	}
	mgc.path = strings.Split(C.GoString(cfiles), ":")
	return nil
}

func (mgc *Magic) Compile(files ...string) error {
	mgc.Lock()
	defer mgc.Unlock()

	var cfiles *C.char
	if len(files) != 0 {
		cfiles = C.CString(strings.Join(files, ":"))
		defer C.free(unsafe.Pointer(cfiles))
	}

	if rv := C.magic_compile_wrapper(mgc.cookie, cfiles); rv < 0 {
		return mgc.error()
	}
	return nil
}

func (mgc *Magic) Check(files ...string) error {
	mgc.Lock()
	defer mgc.Unlock()

	var cfiles *C.char
	if len(files) != 0 {
		cfiles = C.CString(strings.Join(files, ":"))
		defer C.free(unsafe.Pointer(cfiles))
	}

	if rv := C.magic_check_wrapper(mgc.cookie, cfiles); rv < 0 {
		return mgc.error()
	}
	return nil
}

func (mgc *Magic) File(filename string) (string, error) {
	mgc.Lock()
	defer mgc.Unlock()

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

	cstring := C.magic_descriptor(mgc.cookie, C.int(fd))
	if cstring == nil {
		return "", mgc.error()
	}
	return C.GoString(cstring), nil
}

func (mgc *Magic) Version() (int, error) {
	mgc.Lock()
	defer mgc.Unlock()

	rv := int(C.magic_version_wrapper())
	if rv < 0 {
		return rv, &MagicError{rv, "function not implemented"}
	}
	return rv, nil
}

func (mgc *Magic) error() *MagicError {
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

func Compile(files ...string) error {
	mgc := New()
	defer mgc.Close()
	if err := mgc.Compile(files...); err != nil {
		return err
	}
	return nil
}

func Check(files ...string) error {
	mgc := New()
	defer mgc.Close()
	if err := mgc.Check(files...); err != nil {
		return err
	}
	return nil
}

func FileMime(filename string, files ...string) (string, error) {
	mgc := New()
	defer mgc.Close()
	if err := mgc.Load(files...); err != nil {
		return "", err
	}
	mgc.SetFlags(MIME)
	return mgc.File(filename)
}

func FileEncoding(filename string, files ...string) (string, error) {
	mgc := New()
	defer mgc.Close()
	if err := mgc.Load(files...); err != nil {
		return "", err
	}
	mgc.SetFlags(MIME_ENCODING)
	return mgc.File(filename)
}

func FileType(filename string, files ...string) (string, error) {
	mgc := New()
	defer mgc.Close()
	if err := mgc.Load(files...); err != nil {
		return "", err
	}
	mgc.SetFlags(MIME_TYPE)
	return mgc.File(filename)
}

func BufferMime(buffer []byte, files ...string) (string, error) {
	mgc := New()
	defer mgc.Close()
	if err := mgc.Load(files...); err != nil {
		return "", err
	}
	mgc.SetFlags(MIME)
	return mgc.Buffer(buffer)
}

func BufferEncoding(buffer []byte, files ...string) (string, error) {
	mgc := New()
	defer mgc.Close()
	if err := mgc.Load(files...); err != nil {
		return "", err
	}
	mgc.SetFlags(MIME_ENCODING)
	return mgc.Buffer(buffer)
}

func BufferType(buffer []byte, files ...string) (string, error) {
	mgc := New()
	defer mgc.Close()
	if err := mgc.Load(files...); err != nil {
		return "", err
	}
	mgc.SetFlags(MIME_TYPE)
	return mgc.Buffer(buffer)
}
