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

FROM debian:buster as builder

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get dist-upgrade -y -q --no-install-recommends

# Apt-get our dependencies, download, build, and install VIPS, and download and install Go.
ADD preinstall.sh /app/src/github.com/die-net/fotomat/
RUN VIPS_OPTIONS="--prefix=/usr" \
    /app/src/github.com/die-net/fotomat/preinstall.sh

# Install busybox.
RUN apt-get install -y -q --no-install-recommends busybox

# Add the rest of our code.
ADD . /app/src/github.com/die-net/fotomat/

# Build and install Fotomat
RUN GOPATH=/app /usr/local/go/bin/go get -ldflags="-s -w" -t github.com/die-net/fotomat/...

# Test fotomat
RUN GOPATH=/app /usr/local/go/bin/go test -v github.com/die-net/fotomat/...

# Set up an /export/ directory with the very basics of a system
RUN mkdir -m 0755 -p /export/etc /export/home /export/bin /export/usr/bin /export/sbin /export/usr/sbin && \
    mkdir -m 0700 -p /export/root /export/proc /export/dev && \
    mkdir -p -m 1777 /export/tmp
RUN useradd -r fotomat
RUN cp -a --parents \
    /etc/nsswitch.conf \
    /etc/passwd \
    /etc/group \
    /etc/shadow \
    /etc/localtime \
    /usr/share/zoneinfo/UTC \
    /etc/ssl/certs/ca-certificates.crt \
    /export/

# Copy busybox, Fotomat, DNS libraries, and all of their dependencies into /export.
RUN for file in \
        /bin/busybox \
        /app/bin/fotomat \
        /lib/x86_64-linux-gnu/libnss_files.so.2 \
        /lib/x86_64-linux-gnu/libnss_dns.so.2 \
        /lib/x86_64-linux-gnu/libnss_compat.so.2; do \
        echo $file; \
        ldd $file; \
    done | \
    awk '($2=="=>"){print $3};(substr($1,1,1)=="/"){print $1}' | \
    sort -u | \
    xargs -I{} install -D {} /export{}


FROM scratch

ENTRYPOINT ["/app/bin/fotomat"]

CMD ["-listen=:3520"]

EXPOSE 3520

COPY --from=builder /export/ /

VOLUME /tmp

# Expand busybox
RUN ["/bin/busybox", "--install"]

USER fotomat

# Make sure the app runs at all.
RUN ["/app/bin/fotomat", "--version"]
