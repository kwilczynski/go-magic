#if !defined(_COMMON_H)
#define _COMMON_H 1

#if !defined(_GNU_SOURCE)
# define _GNU_SOURCE 1
#endif

#if !defined(_BSD_SOURCE)
# define _BSD_SOURCE 1
#endif

#if defined(__cplusplus)
extern "C" {
#endif

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <errno.h>
#include <assert.h>
#include <locale.h>
#include <xlocale.h>
#include <sys/stat.h>
#include <magic.h>

#if !defined(UNUSED)
# define UNUSED(x) (void)(x)
#endif

#if defined(F_DUPFD_CLOEXEC)
# define HAVE_F_DUPFD_CLOEXEC 1
#endif

#if defined(O_CLOEXEC)
# define HAVE_O_CLOEXEC 1
#endif

#if defined(POSIX_CLOSE_RESTART)
# define HAVE_POSIX_CLOSE_RESTART 1
#endif

#if defined(MAGIC_VERSION) && MAGIC_VERSION >= 513
# define HAVE_MAGIC_VERSION 1
#endif

#if !defined(HAVE_MAGIC_VERSION) || MAGIC_VERSION < 518
# define HAVE_BROKEN_MAGIC 1
#endif

#if defined(__cplusplus)
}
#endif

#endif /* _COMMON_H */

/* vim: set ts=8 sw=4 sts=4 noet : */
