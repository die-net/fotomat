Building:
========

Install [Go](http://golang.org/doc/install), git, and
[VIPS 8.2+](http://www.vips.ecs.soton.ac.uk/index.php?title=Stable).

On OSX, this is as simple as:

    brew install go git homebrew/science/vips

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
