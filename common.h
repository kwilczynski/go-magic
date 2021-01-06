#if !defined(_COMMON_H)
#define _COMMON_H 1

#if !defined(_GNU_SOURCE)
# define _GNU_SOURCE 1
#endif

#if !defined(_BSD_SOURCE)
# define _BSD_SOURCE 1
#endif

#if !defined(_DEFAULT_SOURCE)
# define _DEFAULT_SOURCE 1
#endif

#if defined(__cplusplus)
extern "C" {
#endif

#include <stdio.h>
#include <stdlib.h>
#include <limits.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <errno.h>
#include <assert.h>
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

#if !defined(MAGIC_NO_CHECK_CSV)
# define MAGIC_NO_CHECK_CSV 0
#endif

#if !defined(MAGIC_NO_CHECK_JSON)
# define MAGIC_NO_CHECK_JSON 0
#endif

#if defined(MAGIC_VERSION) && MAGIC_VERSION >= 532
# define HAVE_MAGIC_GETFLAGS 1
#endif

#if defined(__cplusplus)
}
#endif

#endif /* _COMMON_H */
