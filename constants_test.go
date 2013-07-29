/*
 * constants_test.go
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
	"testing"

	. "github.com/kwilczynski/magic"
)

type ConstantsTests struct {
	given    int
	expected int
}

func TestConstants(t *testing.T) {
	mgc := New()
	defer mgc.Close()

	NO_CHECK_BUILTIN_override := 0x3fb000
	if rv := mgc.Version(); rv > 0 {
		NO_CHECK_BUILTIN_override = NO_CHECK_COMPRESS |
			NO_CHECK_TAR |
			NO_CHECK_APPTYPE |
			NO_CHECK_ELF |
			NO_CHECK_TEXT |
			NO_CHECK_CDF |
			NO_CHECK_TOKENS |
			NO_CHECK_ENCODING
	}

	var ct = []ConstantsTests{
		{MIME, MIME_TYPE | MIME_ENCODING},
		{NO_CHECK_BUILTIN, NO_CHECK_BUILTIN_override},
		{NO_CHECK_ASCII, NO_CHECK_TEXT},
	}

	for _, tt := range ct {
		if tt.given != tt.expected {
			t.Errorf("value given 0x%x, want 0x%x",
				tt.given, tt.expected)
		}
	}
}
