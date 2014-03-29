/*
 * constants.go
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

package magic

/*
#include "functions.h"
*/
import "C"

const (
	// No special handling and/or flags specified. Default behaviour.
	NONE int = C.MAGIC_NONE

	// Print debugging messages to standard error output.
	DEBUG int = C.MAGIC_DEBUG

	// If the file queried is a symbolic link, follow it.
	SYMLINK int = C.MAGIC_SYMLINK

	// If the file is compressed, unpack it and look at the contents.
	COMPRESS int = C.MAGIC_COMPRESS

	// If the file is a block or character special device, then open
	// the device and try to look at the contents.
	DEVICES int = C.MAGIC_DEVICES

	// Return a MIME type string, instead of a textual description.
	MIME_TYPE int = C.MAGIC_MIME_TYPE

	//  Return all matches, not just the first.
	CONTINUE int = C.MAGIC_CONTINUE

	// Check the Magic database for consistency and print warnings to
	// standard error output.
	CHECK int = C.MAGIC_CHECK

	// Attempt to preserve access time (atime, utime or utimes) of the
	// file queried on systems that support such system calls.
	PRESERVE_ATIME int = C.MAGIC_PRESERVE_ATIME

	// Do not translate unprintable characters to an octal representation.
	RAW int = C.MAGIC_RAW

	// Treat operating system errors while trying to open files and follow
	// symbolic links as first class errors, instead of storing them in the
	// Magic library error buffer for retrieval later.
	ERROR int = C.MAGIC_ERROR

	// Return a MIME encoding, instead of a textual description.
	MIME_ENCODING int = C.MAGIC_MIME_ENCODING

	// A shorthand for using MIME_TYPE and MIME_ENCODING flags together.
	MIME int = C.MAGIC_MIME

	// Return the Apple creator and type.
	APPLE int = C.MAGIC_APPLE

	// Do not look for, or inside compressed files.
	NO_CHECK_COMPRESS int = C.MAGIC_NO_CHECK_COMPRESS

	// Do not look for, or inside tar archive files.
	NO_CHECK_TAR int = C.MAGIC_NO_CHECK_TAR

	// Do not consult Magic files.
	NO_CHECK_SOFT int = C.MAGIC_NO_CHECK_SOFT

	// Check for EMX application type (only supported on EMX).
	NO_CHECK_APPTYPE int = C.MAGIC_NO_CHECK_APPTYPE

	// Do not check for ELF files (do not examine ELF file details).
	NO_CHECK_ELF int = C.MAGIC_NO_CHECK_ELF

	// Do not check for various types of text files.
	NO_CHECK_TEXT int = C.MAGIC_NO_CHECK_TEXT

	// Do not check for CDF files.
	NO_CHECK_CDF int = C.MAGIC_NO_CHECK_CDF

	// Do not look for known tokens inside ASCII files.
	NO_CHECK_TOKENS int = C.MAGIC_NO_CHECK_TOKENS

	// Return a MIME encoding, instead of a textual description.
	NO_CHECK_ENCODING int = C.MAGIC_NO_CHECK_ENCODING

	// Do not use built-in tests; only consult the Magic files.
	NO_CHECK_BUILTIN int = C.MAGIC_NO_CHECK_BUILTIN

	// Do not check for various types of text files, same as NO_CHECK_TEXT.
	NO_CHECK_ASCII int = C.MAGIC_NO_CHECK_TEXT

	// Do not look for Fortran sequences inside ASCII files.
	NO_CHECK_FORTRAN int = C.MAGIC_NO_CHECK_FORTRAN

	// Do not look for troff sequences inside ASCII files.
	NO_CHECK_TROFF int = C.MAGIC_NO_CHECK_TROFF
)
