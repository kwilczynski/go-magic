package magic_test

import (
	"fmt"

	"github.com/kwilczynski/go-magic"
)

// This example show the basic usage of the package: Open and initialize
// the Magic library, set appropriate flags, for a given file find its
// MIME identification (as per the flag set), print the results and close
// releasing all initialized resources.
func Example_basic() {
	// Open and load the default Magic database.
	m, err := magic.New()
	if err != nil {
		panic(fmt.Sprintf("An has error occurred: %s\n", err))
	}

	m.SetFlags(magic.MIME)
	mime, err := m.File("test/fixtures/gopher.png")
	if err != nil {
		panic(fmt.Sprintf("Unable to determine file MIME: %s\n", err))
	}
	fmt.Printf("File MIME is: %s\n", mime)

	m.Close()
	// Output:
	// File MIME is: image/png; charset=binary
}

// func Example_must() {
// }

// This example shows how to quickly find MIME type for a file.
func ExampleFileType() {
	// The magic.FileType function will open the Magic database,
	// set flags to "MIME", return the result, and then close
	// the Magic database afterwards.
	mime, err := magic.FileType("test/fixtures/gopher.png")
	if err != nil {
		panic(fmt.Sprintf("An has error occurred: %s\n", err))
	}
	fmt.Printf("File MIME type is: %s\n", mime)
	// Output:
	// File MIME type is: image/png
}

// This example show how to identify encoding type of a buffer.
func ExampleBufferEncoding() {
	buffer := []byte("Hello, 世界")

	mime, err := magic.BufferEncoding(buffer)
	if err != nil {
		panic(fmt.Sprintf("An has error occurred: %s\n", err))
	}
	fmt.Printf("Data in the buffer is encoded as: %s\n", mime)
	// Output:
	// Data in the buffer is encoded as: utf-8
}
