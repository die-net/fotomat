fotomat
=======

Go-based image thumbnailing proxy, using many of the size, speed, and
quality optimizations available in
[VIPS](http://www.vips.ecs.soton.ac.uk/index.php?title=Libvips) via the
[Fotomat imager](https://github.com/die-net/fotomat/tree/master/imager)
library.

Building:
--------

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

Alternatively if you use Docker, there's an up-to-date Docker image:

    docker pull dienet/fotomat:latest

Command-line flags:
------------------

    -listen string
        [IP]:port to listen for incoming connections. (default "127.0.0.1:3520")
    -local_image_directory string
        Enable local image serving from this path ("" = proxy instead).
    -max_buffer_pixels uint
        Maximum number of pixels to allocate for an intermediate image buffer. (default 6500000)
    -max_connections int
        The maximum number of incoming connections allowed. (default 9223372036854775807)
    -max_image_threads int
        Maximum number of threads simultaneously processing images. (default 4)
    -max_output_dimension int
        Maximum width or height of an image response. (default 2048)
    -max_processing_duration duration
        Maximum duration we can be processing an image before assuming we crashed (0 = disable). (default 1m0s)

It defaults to:

* Listening on IPv4 localhost on port 3520. Specify ```-listen=:3520``` to listen for remote connections. IPv6 is supported.

* Proxy mode, where the image is fetched from the host supplied in the Host header via http port 80. If you want to disable proxy mode and serve files from a local directory instead, pass ```-local_image_directory```.

* Only allocating image buffers that are at most 6,500,000 pixels (width * height). It can read larger JPEGs than this because it scale them down by a factor of 8 or 16 when decoding.

* Allowing as many VIPS threads to be running as the machine has CPU cores. Raising this probably won't increase throughput, but lowering it may reduce memory usage.

* Allowing output images to be up to 2048 x 2048. Raising this will allow larger images, and be slower.

* Limiting a single VIPS operation to 1 minute, after which it assumes it has hit a VIPS bug and crashes the process.  Raise this if actual image operations take longer.

License
-------

Copyright 2013, 2014, 2015, 2016 Aaron Hopkins and contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at: http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
