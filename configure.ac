AC_INIT(dedup,0.1)
AM_INIT_AUTOMAKE
AC_PREFIX_DEFAULT(/usr/local)

PACKAGE=dedup

AC_CHECK_PROG(HAVE_GO, go, true, false, /bin /usr/local/bin /usr/bin)
if test "x$HAVE_GO" = "xfalse"; then
    AC_MSG_ERROR([Requested program go not found])
fi

AC_PATH_PROG([go],[go])
AC_PATH_PROG([GO],[go])

AC_CHECK_PROG(HAVE_CP, cp, true, false, /bin /usr/local/bin /usr/bin)
if test "x$HAVE_CP" = "xfalse"; then
    AC_MSG_ERROR([Requested program cp not found])
fi
AC_PATH_PROG([CP],[cp])

AC_PROG_INSTALL
AC_CANONICAL_HOST

test "x$prefix" == "xNONE" && prefix=$ac_default_prefix
PREFIX=$prefix

AC_OUTPUT([
Makefile
])
