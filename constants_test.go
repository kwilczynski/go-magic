package magic

import (
	"testing"
)

func TestConstants(t *testing.T) {
	// Any recent version of libmagic have 0x37b000 by default.
	flags := NO_CHECK_COMPRESS | NO_CHECK_TAR |
		NO_CHECK_APPTYPE | NO_CHECK_ELF | NO_CHECK_TEXT |
		NO_CHECK_CDF | NO_CHECK_TOKENS | NO_CHECK_ENCODING

	rv, _ := Version()
	// Older versions of libmagic have 0x3fb000 here historically ...
	if rv < 0 && NO_CHECK_BUILTIN != 0x37b000 {
		flags ^= 0x080000 // 0x37b000 ^ 0x080000 is 0x3fb000
	}
	// Starting from version 5.34, the value libmagic has is 0x77b000 by default.
	if rv > 533 {
		flags ^= 0x0400000 // 0x37b000 ^ 0x040000 is 0x77b000
	}
	// Latest version of libmagic have 0x7fb000 by default.
	if rv > 537 {
		flags ^= 0x0080000 // 0x77b000 ^ 0x0080000 is 0x7fb000
	}

	// Check if underlaying constants coming from libmagic are sane.
	var constantTests = []struct {
		given    int
		expected int
	}{
		{MIME, MIME_TYPE | MIME_ENCODING},
		{NO_CHECK_BUILTIN, flags},
		{NO_CHECK_ASCII, NO_CHECK_TEXT},
	}

	for _, tt := range constantTests {
		if tt.given != tt.expected {
			t.Errorf("value given 0x%x, want 0x%x",
				tt.given, tt.expected)
		}
	}
}

func TestParameters(t *testing.T) {
}
