/*
 * functions.h
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

#ifndef _MAGIC_FUNCTIONS_WRAPPER_H
#define _MAGIC_FUNCTIONS_WRAPPER_H

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <errno.h>
#include <magic.h>

#if defined(__cplusplus)
extern "C" {
#endif

#if defined(__GNUC__) || defined(__clang__)
# define HAVE_WARNING 1
#endif

#if !defined(EINVAL)
# define EINVAL 22
#endif

#if !defined(ENOSYS)
# define ENOSYS 38
#endif

#define SUPPRESS_ERROR_OUTPUT(a, b)       \
    do {                                  \
        (b) = suppress_error_output((a)); \
        if ((b) != 0) {                   \
            return (b);                   \
        }                                 \
    } while(0)

#define RESTORE_ERROR_OUTPUT(a)    \
    do {                           \
        restore_error_output((a)); \
    } while(0)

extern int errno;

extern const char* magic_getpath_wrapper(void);

extern int magic_setflags_wrapper(struct magic_set *ms, int flags);
extern int magic_load_wrapper(struct magic_set *ms, const char *magicfile);
extern int magic_compile_wrapper(struct magic_set *ms, const char *magicfile);
extern int magic_check_wrapper(struct magic_set *ms, const char *magicfile);

extern int magic_version_wrapper(void);

#if defined(__cplusplus)
}
#endif

#endif /* _MAGIC_FUNCTIONS_WRAPPER_H */
