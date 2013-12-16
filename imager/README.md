Imager
======

Golang-based Image Thumbnailing and Cropping Library

Copyright &copy; 2013 Aaron Hopkins tools@die.net

This is a simple wrapper around the [ImageMagick][] [bindings for go][].
[ImageMagick]: http://www.imagemagick.org/
[bindings for go]: https://github.com/gographics/imagick

It aims to generate high-quality, compact images using available performance
optimizations.  Features include:

- JPEG pre-scaling: When used with libjpeg-turbo, can provide a 4x speedup
on scaling large images to small thumbnails.

- Progressive JPEGs: 20% smaller JPEGs with the same image quality.

- [Lanczos][] downsampling: Better-looking thumbnails with fewer artifacts.
[Lanczos]: http://en.wikipedia.org/wiki/Lanczos_resampling

- [Sharpening][]: Remove some of the blurriness caused by downsampling.
[Sharpening]: http://en.wikipedia.org/wiki/Unsharp_masking

- Metadata stripping: Remove potentially large metadata from each image;
particularly useful for images saved by Photoshop.

- Limited input formats: Only accepts common web image formats, preventing
potential attackers from being able to feed bad data to ImageMagick's
rarely-used and potentially buggy image parsers.
