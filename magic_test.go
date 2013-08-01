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
	"os"
	"path"
	"reflect"
	"testing"

	. "github.com/kwilczynski/magic"
)

func CompareStrings(this, other string) bool {
	if this == "" || other == "" {
		return false
	}
	return bytes.Equal([]byte(this), []byte(other))
}

func TestNew(t *testing.T) {
	mgc := New()
	defer mgc.Close()

	func(v interface{}) {
		if _, ok := v.(*Magic); !ok {
			t.Fatalf("not a Magic type: %s", reflect.TypeOf(mgc).String())
		}
	}(mgc)
}

func TestMagic_String(t *testing.T) {
	mgc := New()
	defer mgc.Close()

	magic := reflect.ValueOf(mgc).Elem().FieldByName("magic").Elem()
	cookie := magic.FieldByName("cookie").Elem().Index(0).UnsafeAddr()

	v := fmt.Sprintf("Magic{flags:%d path:%s cookie:0x%x}", 0, []string{}, cookie)
	if ok := CompareStrings(mgc.String(), v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", mgc.String(), v)
	}
}

func TestMagic_Path(t *testing.T) {
	mgc := New()
	defer mgc.Close()

	v := mgc.Path()
	if len(v) == 0 {
		t.Fatalf("value given \"%T\", should not be empty", v)
	}

	p, err := os.Getwd()
	if err != nil {
		t.Fatal("unable to get current and/or working directory")
	}

	p = path.Clean(path.Join(p, "fixtures"))
	if err = os.Setenv("MAGIC", p); err != nil {
		t.Fatalf("unable to set \"MAGIC\" environment variable to \"%s\"", p)
	}

	if ok := CompareStrings(mgc.Path()[0], p); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", mgc.Path()[0], p)
	}

	// TODO(kwilczynski): Test Magic.Load() affecting Magic.Path() as well. But
	// that requires working os.Clearenv() which is yet to be implemented as
	// per http://golang.org/src/pkg/syscall/env_unix.go?s=1772:1787#L101
}

func TestMagic_Flags(t *testing.T) {
	mgc := New()
	defer mgc.Close()

	mgc.SetFlags(MIME)

	v := MIME_TYPE | MIME_ENCODING
	if mgc.Flags() != v {
		t.Errorf("value given 0x%06x, want 0x%06x", mgc.Flags(), v)
	}
}

func TestMagic_SetFlags(t *testing.T) {
	mgc := New()
	defer mgc.Close()

	var flagsTests = []struct {
		flag      int
		expected  int
		incorrect bool
		errno     int
	}{
		{-1, 0x000000, true, -22},
		{NONE, 0x000000, false, 0},
		{MIME_TYPE, 0x000010, false, 0},
		{MIME_ENCODING, 0x000400, false, 0},
		{0xffffff, 0x000400, true, -22},
		{MIME, 0x000010 | 0x000400, false, 0},
		{MIME, MIME_TYPE | MIME_ENCODING, false, 0},
		{NO_CHECK_ASCII, NO_CHECK_TEXT, false, 0},
	}

	var err error
	var given, errno int
	for _, tt := range flagsTests {
		err = mgc.SetFlags(tt.flag)
		given = mgc.Flags()
		if err != nil && tt.incorrect {
			errno = err.(*MagicError).Errno
			if given != tt.expected || errno != tt.errno {
				t.Errorf("value given {0x%06x %d}, want {0x%06x %d}",
					given, errno, tt.expected, tt.errno)
				continue
			}
		}
		if given != tt.expected {
			t.Errorf("value given 0x%06x, want 0x%06x", given, tt.expected)
		}
	}

}

func TestMagic_Load(t *testing.T) {
}

func TestMagic_Compile(t *testing.T) {
}

func TestMagic_Check(t *testing.T) {
}

func TestMagic_File(t *testing.T) {
}

func TestMagic_Buffer(t *testing.T) {
}

func TestMagic_Descriptor(t *testing.T) {
}

func TestMagic_Version(t *testing.T) {
}
