package magic

import (
	"testing"
)

func TestConstants(t *testing.T) {
	var constantTests = []struct {
		given    int
		expected []int
	}{
		{
			MIME,
			[]int{
				MIME_TYPE,
				MIME_ENCODING,
			},
		},
		{
			NO_CHECK_ASCII,
			[]int{
				NO_CHECK_TEXT,
			},
		},
		{
			NO_CHECK_BUILTIN,
			[]int{
				NO_CHECK_COMPRESS,
				NO_CHECK_TAR,
				NO_CHECK_APPTYPE,
				NO_CHECK_ELF,
				NO_CHECK_TEXT,
				NO_CHECK_CSV,
				NO_CHECK_CDF,
				NO_CHECK_TOKENS,
				NO_CHECK_ENCODING,
				NO_CHECK_JSON,
			},
		},
	}

	for _, tt := range constantTests {
		expected := 0
		for _, flag := range tt.expected {
			if flag > -1 {
				expected |= flag
			}
		}

		if tt.given != expected {
			t.Errorf("value given 0x%x, want 0x%x", tt.given, expected)
		}
	}
}

func TestParameters(t *testing.T) {
}
