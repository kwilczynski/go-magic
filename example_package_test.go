/*
 * example_test.go
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

package magic_test

import (
	"fmt"
	"strings"

	"github.com/kwilczynski/go-magic"
)

// This example shows how to use the Separator to split results
// when the CONTINUE flag is set and more than one match was
// returned by the Magic library.
func Example_separator() {
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
	// Should output:
	// Matches for data in the buffer are:
	//      Bourne-Again shell script text executable
	//      a /bin/bash script, ASCII text executable
}

// This example shows how to use Open together with a closure.
func Example_closure() {
	var s string

	// When using magic.Open you don't have to worry
	// about closing the underlying Magic database.
	err := magic.Open(func(m *magic.Magic) error {
		m.SetFlags(magic.MIME)
		mime, err := m.File("fixtures/gopher.png")
		if err != nil {
			return err // Propagate error outside of the closure.
		}
		s = mime // Pass results outside ...
		return nil
	})

	if err != nil {
		panic(fmt.Sprintf("An error occurred: %s\n", err))
	}

	fmt.Printf("File MIME type is: %s\n", s)
	// Output:
	// File MIME type is: image/png; charset=binary
}
