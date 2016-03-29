#!/bin/bash
set -euo pipefail

# This script tries to install the following VIPS and Go versions and their
# dependencies.  Supported Linux distributions include Debian 8/9, Ubuntu,
# Mint, RHEL/Centos/SL 6/7, Fedora, and Amazon Linux.

# Usage: sudo ./preinstall.sh

vips_version=8.2.3
go_version=1.5.3

export PATH="/usr/local/bin:/usr/bin:/bin:${PATH:-}"
export PKG_CONFIG_PATH="/usr/local/lib/pkgconfig:/usr/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
export CFLAGS="${CFLAGS:--O2 -ftree-vectorize -march=native -ffast-math}"
export CXXFLAGS="${CXXFLAGS:-$CFLAGS}"

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
debian-[89]|debian-unknown|ubuntu-1[456].*|mint-17.*)
  # Debian 8-9 or sid, Ubuntu 14-16, Mint 17
  apt-get -q update
  apt-get install -y -q --no-install-recommends ca-certificates git curl tar automake build-essential libglib2.0-dev libjpeg-dev libpng12-dev libwebp-dev libtiff5-dev libexif-dev libmagickwand-dev libfftw3-dev libffi-dev
  ;;
centos-7*|rhel-7*)
  # RHEL/CentOS/SL 7
  yum -y install epel-release
  yum -y update
  yum install -y curl tar findutils git automake make gcc gcc-c++ glib2-devel ImageMagick-devel libexif-devel libjpeg-turbo-devel libpng-devel libtiff-devel libwebp-devel libxml2-devel libffi-devel jbigkit-devel 
  ;;
fedora-2[1-3])
  # Fedora 21-23
  yum install -y curl tar findutils git automake make gcc gcc-c++ glib2-devel ImageMagick-devel libexif-devel libjpeg-turbo-devel libpng-devel libtiff-devel libwebp-devel libxml2-devel libffi-devel jbigkit-devel fftw3-devel fontconfig-devel libtool-ltdl-devel
  ;;
"Red Hat Enterprise Linux release 6."*|"CentOS release 6."*|"Scientific Linux release 6."*)
  # RHEL/CentOS/SL 6
  yum -y install epel-release
  yum -y update
  yum install -y curl tar findutils git automake make gcc gcc-c++ glib2-devel ImageMagick-devel libexif-devel libjpeg-turbo-devel libpng-devel libtiff-devel libwebp-devel libxml2-devel
  ;;
*)
  echo "Sorry, I don't yet know how to install on $release ($(uname -a))."
  exit 1
  ;;
esac

if ! type pkg-config >/dev/null; then
  echo "Sorry, I don't yet know how to install on a system without pkg-config"
fi

if pkg-config --exists vips && pkg-config --atleast-version=$vips_version vips; then
  echo "Found libvips $(pkg-config --modversion vips) installed"
else
  echo "Compiling libvips $vips_version from source"
  rm -rf vips-$vips_version || true
  mkdir vips-$vips_version
  curl -sS http://www.vips.ecs.soton.ac.uk/supported/${vips_version%.*}/vips-${vips_version}.tar.gz | \
    tar --strip-components=1 -C vips-$vips_version -xzf -
  cd vips-$vips_version
  ./configure --disable-debug --disable-dependency-tracking --disable-static --without-orc \
      --with-OpenEXR --with-jpeg --with-lcms --with-libexif --with-magick \
      --with-tiff --with-libwebp --with-png ${VIPS_OPTIONS:-}
  make
  make install
  cd ..
  rm -rf vips-$vips_version
  ldconfig
  echo "Installed libvips $(pkg-config --modversion vips)"
fi

if type go 2>/dev/null; then
  echo "Found $(go version) installed"
else
  arch="$( uname -sm | tr '[A-Z ]' '[a-z-]' | sed 's/i[3-6]86/386/;s/x86_64/amd64/' )"
  if [ "$arch" != "linux-386" -a "$arch" != "linux-amd64" ]; then
    echo "Sorry, I don't know how to install Go for $arch"
    exit 1
  fi

  mkdir -p /usr/local/go && \
  curl -sS https://storage.googleapis.com/golang/go${go_version}.${arch}.tar.gz | \
      tar --strip-components=1 -C /usr/local/go -xzf -
  ln -s ../go/bin/go /usr/local/bin
  echo "Installed $(go version)"
fi

echo "Ready to build Fotomat."
