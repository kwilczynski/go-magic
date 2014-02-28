/*
 * errors_test.go
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

import (
	"reflect"
	"testing"
)

func TestMagicError(t *testing.T) {
	mgc, err := New()
	if err != nil {
		t.Fatalf("unable to create new Magic type: %s", err.Error())
	}
	defer mgc.Close()

	err = mgc.error()
	func(v interface{}) {
		if _, ok := v.(*MagicError); !ok {
			t.Fatalf("not a MagicError type: %s", reflect.TypeOf(v).String())
		}
	}(err)
}

func TestMagicError_Error(t *testing.T) {
	var v string

	mgc, err := New()
	if err != nil {
		t.Fatalf("unable to create new Magic type: %s", err.Error())
	}
	defer mgc.Close()

	v = "magic: no magic files loaded"
	if rv, _ := Version(); rv < 0 {
		// Older version of libmagic behaves differently.
		v = "magic: unknown error"
	}

	err = mgc.error()
	if ok := CompareStrings(err.Error(), v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", err.Error(), v)
	}
}

func TestMagicError_Errno(t *testing.T) {
	var v int

	mgc, err := New()
	if err != nil {
		t.Fatalf("unable to create new Magic type: %s", err.Error())
	}
	defer mgc.Close()

	v = 0
	if rv, _ := Version(); rv < 0 {
		// Older version of libmagic behaves differently.
		v = -1
	}

	e := mgc.error()
	if e.Errno != v {
		t.Errorf("value given %d, want %d", e.Errno, v)
	}
}

func TestMagicError_Message(t *testing.T) {
	var v string

	mgc, err := New()
	if err != nil {
		t.Fatalf("unable to create new Magic type: %s", err.Error())
	}
	defer mgc.Close()

	v = "no magic files loaded"
	if rv, _ := Version(); rv < 0 {
		// Older version of libmagic behaves differently.
		v = "unknown error"
	}

	e := mgc.error()
	if ok := CompareStrings(e.Message, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", e.Message, v)
	}
}
