package magic_test

import (
	"fmt"
	"strings"

	"github.com/kwilczynski/go-magic"
)

// This example shows how to use the Separator to split results
// when the "CONTINUE" flag is set and more than one match was
// returned by the Magic library.
func Example_separator() {
	buffer := []byte("#!/bin/bash\n\n")

	// Open and load the default Magic database.
	m, err := magic.New()
	if err != nil {
		panic(fmt.Sprintf("An has error occurred: %s\n", err))
	}

	m.SetFlags(magic.CONTINUE)
	result, err := m.Buffer(buffer)
	if err != nil {
		panic(fmt.Sprintf("Unable to determine buffer data type: %s\n", err))
	}

	fmt.Println("Matches for data in the buffer are:")
	for _, s := range strings.Split(result, magic.Separator) {
		fmt.Printf("\t%s\n", s)
	}
	m.Close()
	// Output:
	// Matches for data in the buffer are:
	//	Bourne-Again shell script text executable
	//	a /bin/bash script, ASCII text executable
}

// This example shows how to use Open together with a closure.
func Example_closure() {
	var s string

	// When using magic.Open you don't have to worry
	// about closing the the Magic database.
	err := magic.Open(func(m *magic.Magic) error {
		m.SetFlags(magic.MIME)
		mime, err := m.File("test/fixtures/gopher.png")
		if err != nil {
			return err // Propagate error outside of the closure.
		}
		s = mime // Pass results outside.
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("An has error occurred: %s\n", err))
	}
	fmt.Printf("File MIME type is: %s\n", s)
	// Output:
	// File MIME type is: image/png; charset=binary
}

// func Example_disable_autoload() {
// }
