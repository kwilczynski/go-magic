#!/bin/bash

set -e
set -o pipefail

export PATH="/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin:${PATH}"

cat <<'EOF' | tee /etc/apt/apt.conf.d/99apt &>/dev/null
DPkg::options { "--force-confdef"; "--force-confnew"; }
EOF

export DEBIAN_FRONTEND='noninteractive'
export DEBCONF_NONINTERACTIVE_SEEN='true'

apt-get update &> /dev/null

apt-get install --assume-yes \
    wget            \
    git             \
    mercurial       \
    make            \
    bison           \
    flex            \
    build-essential \
    autoconf        \
    automake        \
    autotools-dev   \
    libltdl-dev     \
    libtool         \
    libtool-doc &> /dev/null

apt-get remove --purge --assume-yes \
    libmagic-dev &> /dev/null

for action in 'autoremove' 'autoclean' 'clean'; do
    apt-get --assume-yes "$action" &>/dev/null
done

ARCHIVE_NAME="file-${VERSION}.tar.gz"

rm -Rf \
    "file-${VERSION}" \
    "file-${VERSION}.tar.gz"

MIRRORS=(
    'https://src.fedoraproject.org/lookaside/pkgs/rpms/file'
    'ftp://ftp.astron.com/pub/file'
)

for mirror in "${MIRRORS[@]}"; do
    wget -O "$ARCHIVE_NAME" "${mirror}/${ARCHIVE_NAME}" && break
done

echo "$SHA1 *${ARCHIVE_NAME}" | sha1sum -c

tar -zxf "$ARCHIVE_NAME"

pushd "file-${VERSION}" &> /dev/null

if [[ $VERSION != '5.09' ]]; then
    rm -f src/magic.h
fi

if [[ $VERSION == '5.14' ]]; then
    PATCH_URL='https://gist.githubusercontent.com/kwilczynski/6583179/raw/8c5fbac07472a0a4b68d4028fc9031647ba50876/file-5.14.diff'
    PATCH_LEVEL=1
fi

if [[ $VERSION == "5.18" ]]; then
    PATCH_URL='https://gist.githubusercontent.com/kwilczynski/9925996/raw/483a877749c925a8dcc068de8e3ce5b92f5ea826/338-341-342.diff'
    PATCH_LEVEL=1
fi

if [[ -n $PATCH_URL ]]; then
    curl -sL "$PATCH_URL" > patch.diff

    for option in '--dry-run -s -i' '-i'; do
        if ! patch -l -t -p${PATCH_LEVEL} $option patch.diff; then
            break
        fi
    done
fi

for action in clean distclean; do
    make "$action" || true
done

libtoolize
aclocal
autoheader
automake --add-missing
autoconf

if ! ./configure --prefix=/usr --enable-fsect-man5; then
  $CC --version
  cat config.log
  declare -p
  exit $?
fi

make -j 4
make install

ldconfig

popd &> /dev/null

rm -Rf \
    "file-${VERSION}" \
    "file-${VERSION}.tar.gz"
