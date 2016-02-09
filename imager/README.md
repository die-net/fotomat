# Fotomat Imager

Golang-based Image Thumbnailing and Cropping Library

This is a simple wrapper around the [ImageMagick](http://www.imagemagick.org/) [bindings for go](https://github.com/gographics/imagick) that offers two operations:

* Thumbnail: Scale longest side of original image to fit within W x H pixels.  Examples:

  * 800 x 600 image scaled to fit in 1024 x 1024. Result: 800 x 600
  * 800 x 600 image scaled to fit in 512 x 512. Result: 512 x 384
  * 800 x 600 image scaled to fit in 512 x 256. Result: 341 x 256

* Crop: Scale original image so that the shorter side fits completely within W x H, and cut the excess off of the longer side so that the result is exactly the requested size.

  If the image is too small to do this, generate the largest image that is the same aspect ratio as W x H, and assume the browser will scale it up to exactly W x H.  Examples:

  * 800 x 600 image cropped to 1024 x 1024. Result: 600 x 600
  * 800 x 600 image cropped to 512 x 512. Result: 512 x 512
  * 800 x 600 image cropped to 512 x 256. Result: 512 x 256

It aims to generate high-quality, compact images using available performance optimizations.  Features include:

* JPEG pre-scaling: When used with libjpeg-turbo, can provide a 4x speedup on scaling large images to small thumbnails.

* Progressive JPEGs: Up to 20% smaller JPEGs with the same image quality.

* [Lanczos](http://en.wikipedia.org/wiki/Lanczos_resampling) downsampling: Better-looking thumbnails with fewer artifacts.

* [Sharpening](http://en.wikipedia.org/wiki/Unsharp_masking): Remove some of the blurriness caused by downsampling.

* Auto-rotation: Camera sensors generally only store photos as landscape, with a header indicating which way it should be rotated when decoded. The rotation is applied and the orientation header reset.

* [Color management aware](http://en.wikipedia.org/wiki/ICC_profile): ICC Color profiles are applied, and colors are converted to match the web-standard sRGB before the profile is removed to save space.  Images won't have perfect fidelity on color-managed workstations, but will be much closer than just stripping the color profiles would be.

* Metadata stripping: Remove potentially large metadata from each image; particularly useful for images saved by Photoshop.

* Limited input formats: Only accepts common web image formats (JPG, PNG, and GIF), preventing potential attackers from being able to feed bad data to ImageMagick's rarely-used and potentially buggy image parsers.
