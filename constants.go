/*
 * constants.go
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

package magic

/*
#include "functions.h"
*/
import "C"

const (
	NONE           int = C.MAGIC_NONE
	DEBUG          int = C.MAGIC_DEBUG
	SYMLINK        int = C.MAGIC_SYMLINK
	COMPRESS       int = C.MAGIC_COMPRESS
	DEVICES        int = C.MAGIC_DEVICES
	MIME_TYPE      int = C.MAGIC_MIME_TYPE
	CONTINUE       int = C.MAGIC_CONTINUE
	CHECK          int = C.MAGIC_CHECK
	PRESERVE_ATIME int = C.MAGIC_PRESERVE_ATIME
	RAW            int = C.MAGIC_RAW
	ERROR          int = C.MAGIC_ERROR
	MIME_ENCODING  int = C.MAGIC_MIME_ENCODING
	MIME           int = C.MAGIC_MIME
	APPLE          int = C.MAGIC_APPLE

	NO_CHECK_COMPRESS int = C.MAGIC_NO_CHECK_COMPRESS
	NO_CHECK_TAR      int = C.MAGIC_NO_CHECK_TAR
	NO_CHECK_SOFT     int = C.MAGIC_NO_CHECK_SOFT
	NO_CHECK_APPTYPE  int = C.MAGIC_NO_CHECK_APPTYPE
	NO_CHECK_ELF      int = C.MAGIC_NO_CHECK_ELF
	NO_CHECK_TEXT     int = C.MAGIC_NO_CHECK_TEXT
	NO_CHECK_CDF      int = C.MAGIC_NO_CHECK_CDF
	NO_CHECK_TOKENS   int = C.MAGIC_NO_CHECK_TOKENS
	NO_CHECK_ENCODING int = C.MAGIC_NO_CHECK_ENCODING
	NO_CHECK_BUILTIN  int = C.MAGIC_NO_CHECK_BUILTIN
	NO_CHECK_ASCII    int = C.MAGIC_NO_CHECK_TEXT

	NO_CHECK_FORTRAN int = C.MAGIC_NO_CHECK_FORTRAN
	NO_CHECK_TROFF   int = C.MAGIC_NO_CHECK_TROFF
)
