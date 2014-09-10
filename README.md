fotomat
=======

Golang-based image thumbnailing and cropping proxy, using many of the size,
speed, and quality optimizations available in ImageMagick via the [Fotomat
imager](https://github.com/die-net/fotomat/tree/master/imager) library.

Building:
--------

Install [Go](http://golang.org/doc/install), git, and ImageMagick 6.7+ with
Color Management support.

On OSX, this is:

        brew update && brew install imagemagick --with-little-cms

On RHEL 6.4, I rebuilt a [Fedora Rawhide source
RPM](http://mirror.pnl.gov/fedora/linux/development/rawhide/source/SRPMS/i/),
and commented out a couple of dependencies that aren't available on RHEL 6.

On Debian or Ubuntu, this should just be:

	apt-get update && apt-get install imagemagick imagemagick-devel

Then for all OSes:

	git clone https://github.com/die-net/fotomat.git
	cd fotomat
	export GOPATH=$(pwd)
	go get
	go build

And you'll end up with an "fotomat" binary in the current directory.

Command-line flags:
------------------

	-listen="127.0.0.1:3520": [IP]:port to listen for incoming connections.
	-max_buffer_dimension=2048: Maximum width or height of an image buffer to allocate.
	-max_connections=4096: The maximum number of incoming connections allowed.
	-max_processing_duration=1m0s: Maximum duration we can be processing an image before assuming we crashed (0 = disable).
	-max_threads=4: Maximum number of OS threads to create.

It defaults to dual-stack IPv4/IPv6.  If you want IPv4-only, specify an IPv4
listen address, like -listen="0.0.0.0:8080".

It will try to raise "ulimit -n" to the max_connections that you specify. 
It defaults to raising the limit as much as it can; if you want it higher
than this, you'll likely need to set the ulimit higher as root.

The workers count defaults to the number of CPUs you have in /proc/cpuinfo.
