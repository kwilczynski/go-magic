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
	"reflect"
	"testing"

	. "github.com/kwilczynski/magic"
)

func TestNew(t *testing.T) {
	mgc := New()
	func(v interface{}) {
		if _, ok := v.(*Magic); !ok {
			t.Fatalf("not a Magic type: %s", reflect.TypeOf(mgc).String())
		}
	}(mgc)
}
