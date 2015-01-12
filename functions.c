/*
 * functions.c
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

#if defined(__cplusplus)
extern "C" {
#endif

#include "functions.h"

int suppress_error_output(void *data);
int restore_error_output(void *data);

int override_current_locale(void *data);
int restore_current_locale(void *data);

int
suppress_error_output(void *data)
{
    int local_errno;
    mode_t mode = S_IRWXU | S_IRWXG | S_IRWXO;

    save_t *s = data;
    assert(s != NULL && "Must be a valid pointer to `save_t' type");

    s->data.file.old_fd = -1;
    s->data.file.new_fd = -1;
    s->status = -1;

    fflush(stderr);
    fgetpos(stderr, &s->data.file.position);

    s->data.file.old_fd = dup(fileno(stderr));
    if (s->data.file.old_fd < 0) {
	local_errno = errno;
	goto out;
    }

    s->data.file.new_fd = open("/dev/null", O_WRONLY | O_APPEND, mode);
    if (s->data.file.new_fd < 0) {
	local_errno = errno;

	if (dup2(s->data.file.old_fd, fileno(stderr)) < 0) {
	    local_errno = errno;
	    goto out;
	}

	close(s->data.file.old_fd);
	goto out;
    }

    if (dup2(s->data.file.new_fd, fileno(stderr)) < 0) {
	local_errno = errno;
	goto out;
    }

    close(s->data.file.new_fd);
    return 0;

out:
    s->status = local_errno;
    errno = s->status;

    return -1;
}

int
restore_error_output(void *data)
{
    int local_errno;

    save_t *s = data;
    assert(s != NULL && "Must be a valid pointer to `save_t' type");

    if (s->data.file.old_fd < 0 && s->status != 0) {
	return -1;
    }

    fflush(stderr);

    if (dup2(s->data.file.old_fd, fileno(stderr)) < 0) {
	local_errno = errno;
	goto out;
    }

    close(s->data.file.old_fd);
    clearerr(stderr);
    fsetpos(stderr, &s->data.file.position);

    if (setvbuf(stderr, NULL, _IONBF, 0) != 0) {
	local_errno = errno;
	goto out;
    }

    return 0;

out:
    s->status = local_errno;
    errno = s->status;

    return -1;
}

int
override_current_locale(void *data)
{
    save_t *s = data;
    assert(s != NULL && "Must be a valid pointer to `save_t' type");

    s->status = -1;
    s->data.locale.old_locale = NULL;
    s->data.locale.new_locale = NULL;

    s->data.locale.new_locale = newlocale(LC_ALL_MASK, "C", NULL);
    if (s->data.locale.new_locale == (locale_t)0) {
	goto out;
    }

    assert(s->data.locale.new_locale != NULL && \
	    "Must be a valid pointer to `locale_t' type");

    s->data.locale.old_locale = uselocale(s->data.locale.new_locale);
    if (s->data.locale.old_locale == (locale_t)0) {
	goto out;
    }

    assert(s->data.locale.old_locale != NULL && \
	    "Must be a valid pointer to `locale_t' type");

    s->status = 0;

out:
    return s->status;
}

int
restore_current_locale(void *data)
{
    save_t *s = data;
    assert(s != NULL && "Must be a valid pointer to `save_t' type");

    if (!s->data.locale.new_locale && !s->data.locale.old_locale && s->status != 0) {
	return -1;
    }

    if (uselocale(s->data.locale.old_locale) == (locale_t)0) {
	goto out;
    }

    assert(s->data.locale.new_locale != NULL && \
	    "Must be a valid pointer to `locale_t' type");

    freelocale(s->data.locale.new_locale);

    return 0;

out:
    s->status = -1;
    s->data.locale.old_locale = NULL;
    s->data.locale.new_locale = NULL;

    return -1;
}

inline const char*
magic_getpath_wrapper(void)
{
    return magic_getpath(NULL, 0);
}

inline int
magic_setflags_wrapper(struct magic_set *ms, int flags)
{
    if (flags < MAGIC_NONE || flags > MAGIC_NO_CHECK_BUILTIN) {
	errno = EINVAL;
	return -EINVAL;
    }

    return magic_setflags(ms, flags);
}

inline int
magic_load_wrapper(struct magic_set *ms, const char *magicfile, int flags)
{
    int rv;

    MAGIC_FUNCTION(magic_load, rv, flags, ms, magicfile);

    return rv;
}

inline int
magic_compile_wrapper(struct magic_set *ms, const char *magicfile, int flags)
{
    int rv;

    MAGIC_FUNCTION(magic_compile, rv, flags, ms, magicfile);

    return rv;
}

inline int
magic_check_wrapper(struct magic_set *ms, const char *magicfile, int flags)
{
    int rv;

    MAGIC_FUNCTION(magic_check, rv, flags, ms, magicfile);

    return rv;
}

inline const char*
magic_file_wrapper(struct magic_set *ms, const char* filename, int flags)
{
    const char *cstring;

    MAGIC_FUNCTION(magic_file, cstring, flags, ms, filename);

    return cstring;
}

inline const char*
magic_buffer_wrapper(struct magic_set *ms, const void *buffer, size_t size, int flags)
{
    const char *cstring;

    MAGIC_FUNCTION(magic_buffer, cstring, flags, ms, buffer, size);

    return cstring;
}

inline const char*
magic_descriptor_wrapper(struct magic_set *ms, int fd, int flags)
{
    const char *cstring;

    MAGIC_FUNCTION(magic_descriptor, cstring, flags, ms, fd);

    return cstring;
}

inline int
magic_version_wrapper(void)
{
#if defined(HAVE_MAGIC_VERSION)
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

#if defined(__cplusplus)
}
#endif

/* vim: set ts=8 sw=4 sts=4 noet : */
