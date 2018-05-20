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

	v = "magic: an unknown error has occurred"
	if n, _ := Version(); n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "magic: no magic files loaded"
	}

	err = mgc.error()
	if ok := compareStrings(err.Error(), v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", err.Error(), v)
	}

	v = "the quick brown fox jumps over the lazy dog"

	err = &Error{0, v}
	if ok := compareStrings(err.Error(), fmt.Sprintf("magic: %s", v)); !ok {
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

	v = "an unknown error has occurred"
	if n, _ := Version(); n < 518 && n >= 514 {
		// A few releases of libmagic were having issues.
		v = "no magic files loaded"
	}

	err = mgc.error()
	if ok := compareStrings(err.(*Error).Message, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", err.(*Error).Message, v)
	}

	v = "the quick brown fox jumps over the lazy dog"

	err = &Error{0, v}
	if ok := compareStrings(err.(*Error).Message, v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", err.(*Error).Message, v)
	}
}
