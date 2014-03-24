Imager
======

Golang-based Image Thumbnailing and Cropping Library

This is a simple wrapper around the [ImageMagick][] [bindings for go][].
[ImageMagick]: http://www.imagemagick.org/
[bindings for go]: https://github.com/gographics/imagick

It aims to generate high-quality, compact images using available performance
optimizations.  Features include:

- JPEG pre-scaling: When used with libjpeg-turbo, can provide a 4x speedup
on scaling large images to small thumbnails.

- Progressive JPEGs: Up to 20% smaller JPEGs with the same image quality.

- [Lanczos][] downsampling: Better-looking thumbnails with fewer artifacts.
[Lanczos]: http://en.wikipedia.org/wiki/Lanczos_resampling

- [Sharpening][]: Remove some of the blurriness caused by downsampling.
[Sharpening]: http://en.wikipedia.org/wiki/Unsharp_masking

- [Color management aware][]: ICC Color profiles are applied, and colors are
converted to match the web-standard sRGB before the profile is removed to
save space.  Images won't have perfect fidelity on color-managed
workstations, but will be much closer than just stripping the color profiles
would be.
[Color management aware]: http://en.wikipedia.org/wiki/ICC_profile

- Metadata stripping: Remove potentially large metadata from each image;
particularly useful for images saved by Photoshop.

- Limited input formats: Only accepts common web image formats, preventing
potential attackers from being able to feed bad data to ImageMagick's
rarely-used and potentially buggy image parsers.
