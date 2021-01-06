package magic

import "testing"

func TestConstants(t *testing.T) {
	var constantTests = []struct {
		given    int
		expected int
	}{
		{
			MIME,
			MIME_TYPE | MIME_ENCODING,
		},
		{
			NO_CHECK_ASCII,
			NO_CHECK_TEXT,
		},
		{
			NO_CHECK_FORTRAN,
			0,
		},
		{
			NO_CHECK_TROFF,
			0,
		},
		{
			NO_CHECK_BUILTIN,
			NO_CHECK_COMPRESS | NO_CHECK_TAR | NO_CHECK_APPTYPE | NO_CHECK_ELF | NO_CHECK_TEXT | NO_CHECK_CSV | NO_CHECK_CDF | NO_CHECK_TOKENS | NO_CHECK_ENCODING | NO_CHECK_JSON,
		},
	}

	for _, tt := range constantTests {
		if tt.given != tt.expected {
			t.Errorf("value given 0x%x, want 0x%x", tt.given, tt.expected)
		}
	}
}

// func TestParameters(t *testing.T) {
// }
