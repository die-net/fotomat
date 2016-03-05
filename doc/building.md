Building:
========

Install [Go](http://golang.org/doc/install), git, and
[VIPS 8.2+](http://www.vips.ecs.soton.ac.uk/index.php?title=Stable).

On OSX, this is as simple as:

    brew install go git homebrew/science/vips

On Debian jessie, you can do:

    # Get dependencies
    apt-get install -y -q --no-install-recommends \
       ca-certificates curl \
       git automake build-essential libglib2.0-dev libjpeg-dev libpng12-dev \
       libwebp-dev libtiff5-dev libexif-dev libmagickwand-dev

    # Create a few directories
    mkdir -p /usr/local/go /usr/local/vips

    # Fetch and install Go
    curl -sS https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz | \
        tar --strip-components=1 -C /usr/local/go -xzf -

    # Fetch and build VIPS (enabling GCC's auto-vectorization)
    curl -sS http://www.vips.ecs.soton.ac.uk/supported/8.2/vips-8.2.2.tar.gz | \
        tar --strip-components=1 -C /usr/local/vips -xzf -
    cd /usr/local/vips
    CFLAGS="-O2 -ftree-vectorize -msse4.2 -ffast-math" CXXFLAGS="-O2 -ftree-vectorize -msse4.2 -ffast-math" \
        ./configure --disable-debug --disable-dependency-tracking --disable-gtk-doc-html --disable-pyvips8 --disable-static \
        --with-OpenEXR --with-jpeg --with-lcms --with-libexif --with-magick --with-tiff --with-libwebp --with-png \
        --without-cfitsio --without-fftw --without-gsf --without-matio --without-openslide --without-orc \
        --without-pangoft2 --without-python --without-x --without-zip
    make
    make install
    ldconfig

If you haven't used Go before, you'll need to create a source tree for your Go code:

    mkdir -p $HOME/gocode/src
    export GOPATH=$HOME/gocode

Then for all OSes:

    go get -u github.com/die-net/fotomat
    
And you'll end up with the executable:```$GOPATH/bin/fotomat```

Docker:
------

Alternatively if you use Docker, there's a
[Dockerfile](https://github.com/die-net/fotomat/blob/master/Dockerfile)
which is used to build an up-to-date
[Docker image](https://hub.docker.com/r/dienet/fotomat/). Fetch it with:

    docker pull dienet/fotomat:latest
