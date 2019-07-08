#!/bin/bash
set -euo pipefail

# This script tries to install the following VIPS and Go versions and their
# dependencies.  Supported Linux distributions include Debian 8/9, Ubuntu,
# Mint, RHEL/Centos/SL 6/7, Fedora, and Amazon Linux.

# Usage: sudo ./preinstall.sh

VIPS_VERSION=${VIPS_VERSION:-8.7.4}
GO_VERSION=${GO_VERSION:-1.12.6}

export PATH="/usr/local/bin:/usr/bin:/bin:${PATH:-}"
export PKG_CONFIG_PATH="/usr/local/lib/pkgconfig:/usr/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
export CFLAGS="${CFLAGS:--O2 -ftree-vectorize -msse2 -ffast-math -fPIC}"
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
  apt-get install -y -q --no-install-recommends automake build-essential ca-certificates curl git libexif-dev libexpat1-dev libffi-dev libfftw3-dev libgif-dev libglib2.0-dev libjpeg-dev liblcms2-dev libpng12-dev libpoppler-glib-dev librsvg2-dev libselinux1-dev libtiff5-dev libwebp-dev libxml2-dev tar
  ;;
debian-9|debian-10|debian-unknown|ubuntu-1[789].*|mint-1[89].*)
  # Debian 9, 10, or sid, Ubuntu 17-19, Mint 18-19
  apt-get -q update
  apt-get install -y -q --no-install-recommends automake build-essential ca-certificates curl git libexif-dev libexpat1-dev libffi-dev libfftw3-dev libgif-dev libglib2.0-dev libjpeg-dev liblcms2-dev libmount-dev libpng-dev libpoppler-glib-dev librsvg2-dev libselinux1-dev libtiff5-dev libwebp-dev libxml2-dev libzstd-dev tar
  ;;
amzn-*|centos-7*|ol-7*|rhel-7*|scientific-7*)
  # RHEL/CentOS/SL 7/Amazon Linux 2/Oracle Linux 7
  yum -y update
  yum install -y automake bzip2-devel curl expat-devel findutils gcc gcc-c++ giflib-devel git glib2-devel jbigkit-devel lcms2-devel libexif-devel libffi-devel libjpeg-turbo-devel libmount-devel libpng-devel librsvg2-devel libselinux-devel libtiff-devel libwebp-devel libxml2-devel make poppler-glib-devel tar
  ;;
fedora-2[6-9])
  # Fedora 26-29
  yum -y update
  yum install -y automake curl expat-devel fftw3-devel findutils fontconfig-devel gcc gcc-c++ giflib-devel git glib2-devel jasper-libs jbigkit-devel lcms2-devel libexif-devel libffi-devel libjpeg-turbo-devel libmount-devel libpng-devel librsvg2-devel libselinux-devel libtiff-devel libtool-ltdl-devel libwebp-devel libxml2-devel make poppler-glib-devel tar
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
  url="https://github.com/libvips/libvips/releases/download/v${VIPS_VERSION}/vips-${VIPS_VERSION}.tar.gz"
  echo "Building libvips $VIPS_VERSION from source $url"
  rm -rf vips-$VIPS_VERSION || true
  mkdir vips-$VIPS_VERSION
  curl -sSL "$url" | tar --strip-components=1 -C vips-$VIPS_VERSION -xzf -
  cd vips-$VIPS_VERSION
  ./configure \
      --disable-debug --disable-dependency-tracking --disable-gtk-doc-html \
      --disable-pyvips8 --disable-static --without-analyze --without-cfitsio \
      --without-fftw --without-gsf --without-magick --without-matio \
      --without-openslide --without-orc --without-pangoft2 --without-ppm \
      --without-radiance --without-x \
      --with-OpenEXR --with-jpeg --with-lcms --with-libexif --with-giflib \
      --with-libwebp --with-png --with-poppler --with-rsvg --with-tiff \
      ${VIPS_OPTIONS:-}
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

  url="https://storage.googleapis.com/golang/go${GO_VERSION}.${arch}.tar.gz"
  echo "Installing Go from ${url}"
  mkdir -p /usr/local/go && \
  curl -sSL $url | tar --strip-components=1 -C /usr/local/go -xzf -
  ln -sf ../go/bin/go /usr/local/bin/go
  echo "Installed $(go version)"
fi

echo "Ready to build Fotomat."
