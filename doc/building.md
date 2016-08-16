Building:
========

Install [Go](http://golang.org/doc/install), git, and
[VIPS 8.3+](http://www.vips.ecs.soton.ac.uk/index.php?title=Stable).

If you haven't used Go before, first create a source tree for your Go code:

    mkdir -p $HOME/gocode/src
    export GOPATH=$HOME/gocode

On OSX, you can install all of the dependencies with [Homebrew](http://brew.sh/):

    brew install go git homebrew/science/vips

On most Linux flavors, you can install Go and VIPS with the ```preinstall.sh``` script:

    git clone https://github.com/die-net/fotomat.git $GOPATH/src/github.com/die-net/fotomat/
    cd $GOPATH/src/github.com/die-net/fotomat/
    sudo ./preinstall.sh

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
