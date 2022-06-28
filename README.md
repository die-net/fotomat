Fotomat [![Build Status](https://github.com/die-net/fotomat/actions/workflows/go-test.yml/badge.svg)](https://github.com/die-net/fotomat/actions/workflows/go-test.yml) [![Coverage Status](https://coveralls.io/repos/github/die-net/fotomat/badge.svg?branch=main)](https://coveralls.io/github/die-net/fotomat?branch=main) [![Go Report Card](https://goreportcard.com/badge/github.com/die-net/fotomat)](https://goreportcard.com/report/github.com/die-net/fotomat)
=======

Fotomat is an extremely fast image resizing proxy, enabling on-the-fly resizing and cropping of JPEG, PNG, GIF, and WebP images. Written in [Go](https://golang.org/doc/) and using the fast and flexible [VIPS](http://www.vips.ecs.soton.ac.uk/index.php?title=Libvips) image library, it aims to deliver beautiful images in the shortest time and at the smallest file size possible.

Documentation
-------------

See [features](https://github.com/die-net/fotomat/blob/main/doc/features.md), [building instructions](https://github.com/die-net/fotomat/blob/main/doc/building.md), [command-line
flags](https://github.com/die-net/fotomat/blob/main/doc/flags.md), and
[benchmarks](https://github.com/die-net/fotomat/blob/main/doc/benchmarks.md).

There's also API detail for Fotomat's [thumbnail](https://godoc.org/github.com/die-net/fotomat/thumbnail), [format](https://godoc.org/github.com/die-net/fotomat/format), and [vips wrapper](https://godoc.org/github.com/die-net/fotomat/vips) libraries.


License
-------

Copyright 2013, 2014, 2015, 2016, 2017, 2018 Aaron Hopkins and contributors

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at: http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
