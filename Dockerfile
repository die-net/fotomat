FROM debian:jessie

ADD . /app/src/github.com/die-net/fotomat

ENTRYPOINT ["/app/bin/fotomat"]

CMD ["-listen=:3520"]

EXPOSE 3520

RUN apt-get -q update && \
    apt-get -y -q dist-upgrade && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y -q --no-install-recommends \
       ca-certificates curl git automake build-essential \
       gobject-introspection gtk-doc-tools libglib2.0-dev libjpeg-dev \
       libpng12-dev libwebp-dev libtiff5-dev libexif-dev libxml2-dev swig libmagickwand-dev libpango1.0-dev \
       libmatio-dev libopenslide-dev && \
    mkdir -p /usr/local/go /usr/local/vips /app/pkg /app/bin && \
    curl -sS https://storage.googleapis.com/golang/go1.4.3.linux-amd64.tar.gz | \
        tar --strip-components=1 -C /usr/local/go -xzf - && \
    curl -sS http://www.vips.ecs.soton.ac.uk/supported/8.2/vips-8.2.2.tar.gz | \
        tar --strip-components=1 -C /usr/local/vips -xzf - && \
    cd /usr/local/vips && \
    ./configure --enable-debug=no --without-python --without-orc --without-fftw --without-gsf && \
    make && make install && ldconfig && \
    GOPATH=/app /usr/local/go/bin/go get -t github.com/die-net/fotomat github.com/die-net/fotomat/imager && \
    GOPATH=/app /usr/local/go/bin/go test github.com/die-net/fotomat github.com/die-net/fotomat/imager && \
    apt-get remove -y curl automake build-essential && \
    apt-get autoremove -y && \
    apt-get autoclean && \
    apt-get clean && \
    rm -rf /usr/local/go /usr/local/vips /app/pkg /var/lib/apt/lists/*
