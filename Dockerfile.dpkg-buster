# Build Fotomat RPM for Debian buster using Docker.
#
# Run: dist/build dpkg-buster
#
# And you'll end up with a fotomat*.dpkg in the current directory.

FROM debian:buster

# Apt-get our dependencies, download, build, and install VIPS, and download and install Go.
ADD preinstall.sh /app/src/github.com/die-net/fotomat/
RUN DEBIAN_FRONTEND=noninteractive CFLAGS="-O2 -ftree-vectorize -msse2 -ffast-math -fPIC" \
    VIPS_OPTIONS="--disable-shared --enable-static" \
    /app/src/github.com/die-net/fotomat/preinstall.sh

# Add dpkg build tool.
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y -q --no-install-recommends fakeroot

# Create a few directories
# mkdir -p /app/pkg /app/bin

# Add the rest of our code.
ADD . /app/src/github.com/die-net/fotomat/

# Build and install Fotomat
RUN GOPATH=/app CGO_LDFLAGS_ALLOW="-Wl,--export-dynamic" /usr/local/go/bin/go get -tags vips_static -t github.com/die-net/fotomat/...

# Test fotomat
RUN GOPATH=/app CGO_LDFLAGS_ALLOW="-Wl,--export-dynamic" /usr/local/go/bin/go test -tags vips_static -v github.com/die-net/fotomat/...

# Build the dpkg.
RUN /app/src/github.com/die-net/fotomat/dist/build-dpkg /app/bin/fotomat
