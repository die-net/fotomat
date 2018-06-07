# Build Fotomat RPM for CentOS 7 using Docker.
#
# Run: dist/build rpm-centos7
#
# And you'll end up with a fotomat*.rpm in the current directory.

FROM centos:7

# Apt-get our dependencies, download, build, and install VIPS, and download and install Go.
ADD preinstall.sh /app/src/github.com/die-net/fotomat/

RUN CFLAGS="-O2 -ftree-vectorize -msse2 -ffast-math -fPIC" LDFLAGS="-lstdc++" VIPS_OPTIONS="--disable-shared --enable-static" \
    /app/src/github.com/die-net/fotomat/preinstall.sh

# Add a tool for building RPMs.
RUN yum -y install rpm-build

# Add the rest of our code.
ADD . /app/src/github.com/die-net/fotomat/

# Build and install Fotomat
RUN PKG_CONFIG_PATH=/usr/local/lib/pkgconfig GOPATH=/app CGO_LDFLAGS_ALLOW="-Wl,--export-dynamic" \
    /usr/local/go/bin/go get -tags vips_static -t github.com/die-net/fotomat/...

# Test fotomat
RUN PKG_CONFIG_PATH=/usr/local/lib/pkgconfig GOPATH=/app CGO_LDFLAGS_ALLOW="-Wl,--export-dynamic" \
    /usr/local/go/bin/go test -tags vips_static -v github.com/die-net/fotomat/...

# Update specfile version and use it to build binary RPM.
RUN perl -ne '/FotomatVersion.*\b(\d+\.\d+\.\d+)/ and print "$1\n"' /app/src/github.com/die-net/fotomat/cmd/fotomat/version.go | \
    xargs -i{} perl -p -i~ -e 's/(^Version:\s+)\d+\.\d+\.\d+/${1}{}/' /app/src/github.com/die-net/fotomat/dist/rpm/fotomat.spec
RUN rpmbuild -bb /app/src/github.com/die-net/fotomat/dist/rpm/fotomat.spec
