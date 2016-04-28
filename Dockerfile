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

FROM debian:jessie

ADD . /app/src/github.com/die-net/fotomat

ENTRYPOINT ["/app/bin/fotomat"]

CMD ["-listen=:3520"]

EXPOSE 3520

RUN \
    # Apply updates and install our dependencies
    apt-get -q update && \
    apt-get -y -q dist-upgrade && \
    # Apt-get our dependencies, download, build, and install VIPS, and download and install Go.
    DEBIAN_FRONTEND=noninteractive CFLAGS="-O2 -ftree-vectorize -msse2 -ffast-math" VIPS_OPTIONS="--disable-gtk-doc-html --disable-pyvips8 --without-cfitsio --without-fftw --without-gsf --without-matio --without-openslide --without-orc --without-pangoft2 --without-python --without-x --without-zip" \
        /app/src/github.com/die-net/fotomat/preinstall.sh && \

    # Create a few directories
    mkdir -p /app/pkg /app/bin && \

    # Build, install, and test fotomat
    GOPATH=/app /usr/local/go/bin/go get -t github.com/die-net/fotomat/... && \
    GOPATH=/app /usr/local/go/bin/go test -v github.com/die-net/fotomat/... && \
    strip /app/bin/fotomat && \

    # Add a fotomat user for it to run as, and make filesystem read-only to that user.
    useradd -m fotomat -s /bin/bash && \

    # Mark fotomat's dependencies as needed, to avoid autoremoval
    ldd /app/bin/fotomat | awk '($2=="=>"&&substr($3,1,11)!="/usr/local/"){print $3}' | \
        xargs dpkg -S | cut -d: -f1 | sort -u | xargs apt-get install && \

    # And remove almost everything else that we installed
    apt-get remove -y git automake build-essential libglib2.0-dev libjpeg-dev libpng12-dev \
       libwebp-dev libtiff5-dev libexif-dev libgif-dev libfftw3-dev libffi-dev && \
    apt-get autoremove -y && \
    apt-get autoclean && \
    apt-get clean && \
    rm -rf /usr/local/go /app/pkg /var/lib/apt/lists/* /tmp/*

# Start by default as a non-root user.
USER fotomat
