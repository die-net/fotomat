# Fotomat as a Docker image meant to be used directly on Docker-based
# production systems.
#
# Automatically built by Docker Hub and available as dienet/fotomat:latest.
# To rebuild locally: docker build -t dienet/fotomat:latest .
#
# To run serving local images from /path/to/images:
#   docker run -v /path/to/images:/images dienet/fotomat:latest -listen=:3520 -local_image_directory=/images
#
# To run as an HTTP image proxy, trusting the host header:
#   docker run dienet/fotomat:latest -listen=:3520

FROM debian:stretch as builder

# Set up an /export/ directory with a tmp/ and etc/passwd
RUN mkdir -p -m 1777 /export/tmp
RUN useradd -r fotomat
RUN install -D /etc/passwd /export/etc/passwd

# Apt-get our dependencies, download, build, and install VIPS, and download and install Go.
ADD preinstall.sh /app/src/github.com/die-net/fotomat/
RUN DEBIAN_FRONTEND=noninteractive CFLAGS="-O2 -ftree-vectorize -msse2 -ffast-math" \
    VIPS_OPTIONS="--prefix=/usr" \
    /app/src/github.com/die-net/fotomat/preinstall.sh

# Add the rest of our code.
ADD . /app/src/github.com/die-net/fotomat/

# Build and install Fotomat
RUN GOPATH=/app /usr/local/go/bin/go get -t github.com/die-net/fotomat/...

# Test fotomat
RUN GOPATH=/app /usr/local/go/bin/go test -v github.com/die-net/fotomat/...

# Install Fotomat and all of its dependencies into /export.
RUN install -sD /app/bin/fotomat /export/app/bin/fotomat
RUN ldd /app/bin/fotomat | awk '($2=="=>"){print $3};(substr($1,1,1)=="/"){print $1}' | xargs -i{} install -D {} /export{}


FROM scratch

ENTRYPOINT ["/app/bin/fotomat"]

CMD ["-listen=:3520"]

EXPOSE 3520

COPY --from=builder /export/ /

USER fotomat

# Make sure the app runs at all.
RUN ["/app/bin/fotomat", "--version"]
