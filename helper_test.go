/*
 * helper_test.go
 *
 * Copyright 2013-2015 Krzysztof Wilczynski
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
	"bytes"
	"path"
	"reflect"
	"testing"
)

var (
	// Auxiliary files for use in tests ...
	fixturesDirectory = "fixtures"

	// Default directory containing files using old-style Magic format.
	formatDirectory = "old-format"

	// PNG image data, 1634 x 2224, 8-bit/color RGBA, non-interlaced
	sampleImageFile = path.Clean(path.Join(fixturesDirectory, "gopher.png"))

	// Magic file for testing only ...
	shellMagicFile = path.Clean(path.Join(fixturesDirectory, "shell.magic"))
)

func Skip(t *testing.T, message string) {
	// XXX(krzysztof): Attempt to circumvent lack of T.Skip() prior to Go version go1.1 ...
	f := reflect.ValueOf(t).MethodByName("Skip")
	if ok := f.IsValid(); !ok {
		f = reflect.ValueOf(t).MethodByName("Log")
	}

	f.Call([]reflect.Value{reflect.ValueOf(message)})
}

func CompareStrings(this, other string) bool {
	if this == "" || other == "" {
		return false
	}
	return bytes.Equal([]byte(this), []byte(other))
}
