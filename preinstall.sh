#!/bin/bash
set -euo pipefail

# This script tries to install the following VIPS and Go versions and their
# dependencies.  Supported Linux distributions include Debian 8/9, Ubuntu,
# Mint, RHEL/Centos/SL 6/7, Fedora, and Amazon Linux.

# Usage: sudo ./preinstall.sh

VIPS_VERSION=${VIPS_VERSION:-8.5.9}
GO_VERSION=${GO_VERSION:-1.10.2}

export PATH="/usr/local/bin:/usr/bin:/bin:${PATH:-}"
export PKG_CONFIG_PATH="/usr/local/lib/pkgconfig:/usr/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
export CFLAGS="${CFLAGS:--O2 -ftree-vectorize -march=native -ffast-math}"
export CXXFLAGS="${CXXFLAGS:-$CFLAGS}"

function ver {
  printf "%03d%03d%03d%03d" $(echo "$1" | tr '.' ' ')
}

if [ "$(uname -s)" != "Linux" ]; then
  echo "Sorry, this script is only useful on Linux."
  exit 1
fi

# Verify that we're running as root
if [ "$(id -u)" -ne "0" ]; then
  echo "Sorry, I can't install without sudo or root. Try: sudo $0"
  exit 1
fi

# Try to get a compact release string for various Linux distributions.
if [ -f /etc/os-release ]; then
  . /etc/os-release
  release="${ID}-${VERSION_ID:-unknown}"  # VERSION_ID isn't set for sid.
else
  release="$( cat /etc/redhat-release /etc/system-release 2> /dev/null | head -1 || true )"
  if [ -z "$release" ]; then
    release="unknown"
  fi
fi

# Try to figure out how to install our dependencies
case "$release" in
debian-8|ubuntu-1[456].*|mint-17.*)
  # Debian 8, Ubuntu 14-16, Mint 17
  apt-get -q update
  apt-get install -y -q --no-install-recommends ca-certificates git curl tar automake build-essential libglib2.0-dev libjpeg-dev libpng12-dev libwebp-dev libgif-dev liblcms2-dev libtiff5-dev libxml2-dev libexif-dev libexpat1-dev libfftw3-dev libffi-dev
  ;;
debian-9|debian-unknown)
  # Debian 9 or sid
  apt-get -q update
  apt-get install -y -q --no-install-recommends ca-certificates git curl tar automake build-essential libglib2.0-dev libjpeg-dev libpng-dev libwebp-dev libgif-dev liblcms2-dev libtiff5-dev libxml2-dev libexif-dev libexpat1-dev libfftw3-dev libffi-dev
  ;;
centos-7*|rhel-7*)
  # RHEL/CentOS/SL 7
  yum -y install epel-release
  yum -y update
  yum install -y curl tar findutils git automake make gcc gcc-c++ glib2-devel libexif-devel libjpeg-turbo-devel libpng-devel libtiff-devel libwebp-devel giflib-devel lcms2-devel libxml2-devel expat-devel libffi-devel jbigkit-devel
  ;;
fedora-2[1-3])
  # Fedora 21-23
  yum install -y curl tar findutils git automake make gcc gcc-c++ glib2-devel libexif-devel libjpeg-turbo-devel libpng-devel libtiff-devel libwebp-devel giflib-devel lcms2-devel libxml2-devel expat-devel libffi-devel jbigkit-devel fftw3-devel fontconfig-devel libtool-ltdl-devel
  ;;
"Red Hat Enterprise Linux release 6."*|"CentOS release 6."*|"Scientific Linux release 6."*)
  # RHEL/CentOS/SL 6
  yum -y install epel-release
  yum -y update
  yum install -y curl tar findutils git automake make gcc gcc-c++ glib2-devel libexif-devel libjpeg-turbo-devel libpng-devel libtiff-devel libwebp-devel giflib-devel lcms2-devel libxml2-devel expat-devel
  ;;
*)
  echo "Sorry, I don't yet know how to install on $release ($(uname -a))."
  exit 1
  ;;
esac

if ! type pkg-config >/dev/null; then
  echo "Sorry, I don't yet know how to install on a system without pkg-config"
fi

if pkg-config --exists vips && pkg-config --atleast-version=$VIPS_VERSION vips; then
  echo "Found libvips $(pkg-config --modversion vips) installed"
elif [ "$VIPS_VERSION" = "skip" ]; then
  echo "Skipping VIPS installation"
else
  url="https://github.com/jcupitt/libvips/releases/download/v${VIPS_VERSION}/vips-${VIPS_VERSION}.tar.gz"
  echo "Building libvips $VIPS_VERSION from source $url"
  rm -rf vips-$VIPS_VERSION || true
  mkdir vips-$VIPS_VERSION
  curl -sSL "$url" | tar --strip-components=1 -C vips-$VIPS_VERSION -xzf -
  cd vips-$VIPS_VERSION
  ./configure --disable-debug --disable-dependency-tracking --disable-static --without-orc --without-magick \
      --with-OpenEXR --with-jpeg --with-lcms --with-libexif --with-giflib \
      --with-tiff --with-libwebp --with-png ${VIPS_OPTIONS:-}
  make -j $( getconf _NPROCESSORS_ONLN 2> /dev/null || echo 1 )
  make install
  cd ..
  rm -rf vips-$VIPS_VERSION
  ldconfig
  echo "Installed libvips $(pkg-config --modversion vips)"
fi

if type go 2>/dev/null; then
  echo "Found $(go version) installed"
elif [ "$GO_VERSION" = "skip" ]; then
  echo "Skipping Go installation"
else
  arch="$( uname -sm | tr '[A-Z ]' '[a-z-]' | sed 's/i[3-6]86/386/;s/x86_64/amd64/' )"
  if [ "$arch" != "linux-386" -a "$arch" != "linux-amd64" ]; then
    echo "Sorry, I don't know how to install Go for $arch"
    exit 1
  fi

  mkdir -p /usr/local/go && \
  curl -sSL https://storage.googleapis.com/golang/go${GO_VERSION}.${arch}.tar.gz | \
      tar --strip-components=1 -C /usr/local/go -xzf -
  ln -sf ../go/bin/go /usr/local/bin/go
  echo "Installed $(go version)"
fi

echo "Ready to build Fotomat."
