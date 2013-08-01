/*
 * functions.c
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

#include "functions.h"

struct save {
    int fd;
    fpos_t position;
};

typedef struct save save_t;

static int
suppress_error_output(void *p)
{
    save_t *s = p;

    fflush(stderr);
    fgetpos(stderr, &s->position);

    s->fd = dup(fileno(stderr));
    
    if (freopen("/dev/null", "w", stderr) == NULL) {
        return errno;
    }
    return 0;
}

static void
restore_error_output(void *p)
{
    save_t *s = p;

    fflush(stderr);
    dup2(s->fd, fileno(stderr));
    close(s->fd);
    clearerr(stderr);
    fsetpos(stderr, &s->position);
    setvbuf(stderr, NULL, _IONBF, 0);
}

inline const char*
magic_getpath_wrapper(void)
{
    return magic_getpath(NULL, 0);
}

inline int
magic_setflags_wrapper(struct magic_set *ms, int flags) {
    if (flags < MAGIC_NONE || flags > MAGIC_NO_CHECK_BUILTIN) {
        errno = EINVAL;
        return -EINVAL;
    }
    return magic_setflags(ms, flags);
}

inline int
magic_load_wrapper(struct magic_set *ms, const char *magicfile)
{
    int rv;
    save_t s;
    
    SUPPRESS_ERROR_OUTPUT(&s, rv);
    rv = magic_load(ms, magicfile);
    RESTORE_ERROR_OUTPUT(&s);

    return rv;
}

inline int
magic_compile_wrapper(struct magic_set *ms, const char *magicfile)
{
    int rv;
    save_t s;
    
    SUPPRESS_ERROR_OUTPUT(&s, rv);
    rv = magic_compile(ms, magicfile);
    RESTORE_ERROR_OUTPUT(&s);

    return rv;
}

inline int
magic_check_wrapper(struct magic_set *ms, const char *magicfile)
{
    int rv;
    save_t s;
    
    SUPPRESS_ERROR_OUTPUT(&s, rv);
    rv = magic_check(ms, magicfile);
    RESTORE_ERROR_OUTPUT(&s);

    return rv;
}

inline int
magic_version_wrapper(void)
{
#if defined(MAGIC_VERSION) && MAGIC_VERSION >= 513
    return magic_version();
#else
# if defined(HAVE_WARNING)
#  warning "function `int magic_version(void)' not implemented"
# else
#  pragma message("function `int magic_version(void)' not implemented")
# endif
    errno = ENOSYS;
    return -ENOSYS;
#endif
}
