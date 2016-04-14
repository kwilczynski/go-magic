/*
 * errors_test.go
 *
 * Copyright 2013-2016 Krzysztof Wilczynski
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

import (
	"fmt"
	"reflect"
	"testing"
)

func TestError(t *testing.T) {
	mgc, err := New()
	if err != nil {
		t.Fatalf("unable to create new Magic type: %s", err.Error())
	}
	defer mgc.Close()

	err = mgc.error()
	func(v interface{}) {
		if _, ok := v.(*Error); !ok {
			t.Fatalf("not a Error type: %s", reflect.TypeOf(v).String())
		}
	}(err)
}

func TestError_Error(t *testing.T) {
	var v string

	mgc, err := New()
	if err != nil {
		t.Fatalf("unable to create new Magic type: %s", err.Error())
	}
	defer mgc.Close()

	v = "magic: unknown error"
	if n, _ := Version(); n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	}

	err = mgc.error()
	if ok := CompareStrings(err.Error(), v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", err.Error(), v)
	}

	v = "the quick brown fox jumps over the lazy dog"

	err = &Error{0, v}
	if ok := CompareStrings(err.Error(), fmt.Sprintf("magic: %s", v)); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", err.Error(), v)
	}
}

func TestError_Errno(t *testing.T) {
	var v int

	mgc, err := New()
	if err != nil {
		t.Fatalf("unable to create new Magic type: %s", err.Error())
	}
	defer mgc.Close()

	v = -1
	if n, _ := Version(); n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = 0
	}

	err = mgc.error()
	if err.(*Error).Errno != v {
		t.Errorf("value given %d, want %d", err.(*Error).Errno, v)
	}

	v = 42

	err = &Error{v, ""}
	if err.(*Error).Errno != v {
		t.Errorf("value given %d, want %d", err.(*Error).Errno, v)
	}
}

func TestError_Message(t *testing.T) {
	var v string

	mgc, err := New()
	if err != nil {
		t.Fatalf("unable to create new Magic type: %s", err.Error())
	}
	defer mgc.Close()

	v = "unknown error"
	if n, _ := Version(); n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "no magic files loaded"
	}

	err = mgc.error()
	if ok := CompareStrings(err.(*Error).Message, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", err.(*Error).Message, v)
	}

	v = "the quick brown fox jumps over the lazy dog"

	err = &Error{0, v}
	if ok := CompareStrings(err.(*Error).Message, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", err.(*Error).Message, v)
	}
}
