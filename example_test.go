/*
 * example_test.go
 *
 * Copyright 2013-2014 Krzysztof Wilczynski
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
	"fmt"
	"strings"

	"github.com/kwilczynski/go-magic"
)

func ExampleMagic() {
	// Open and load default Magic database ...
	m, err := magic.New()
	if err != nil {
		panic(fmt.Sprintf("An error occurred: %s\n", err))
	}

	m.SetFlags(magic.MIME)
	mime, err := m.File("fixtures/gopher.png")
	if err != nil {
		panic(fmt.Sprintf("Unable to determine file MIME: %s\n", err))
	}
	fmt.Printf("File MIME type is: %s\n", mime)

	m.Close()
	// Output:
	// File MIME type is: image/png; charset=binary
}

func ExampleMagic_Continue() {
	buffer := []byte("#!/bin/bash\n\n")

	// Open and load default Magic database ...
	m, err := magic.New()
	if err != nil {
		panic(fmt.Sprintf("An error occurred: %s\n", err))
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

func ExampleFileType() {
	mime, err := magic.FileType("fixtures/gopher.png")
	if err != nil {
		panic(fmt.Sprintf("An error occurred: %s\n", err))
	}
	fmt.Printf("File type is: %s\n", mime)
	// Output:
	// File type is: image/png
}

func ExampleBufferEncoding() {
	buffer := []byte("Hello, 世界")

	mime, err := magic.BufferEncoding(buffer)
	if err != nil {
		panic(fmt.Sprintf("An error occurred: %s\n", err))
	}
	fmt.Printf("Data in the buffer is enconded as: %s\n", mime)
	// Output:
	// Data in the buffer is enconded as: utf-8
}
