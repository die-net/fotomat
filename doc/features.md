Features
========

Fotomat aims to generate high-quality, compact images using available performance optimizations.  Features include:

* JPEG pre-scaling: When used with libjpeg-turbo, can provide a 4x speedup on scaling large images to small thumbnails.

* Generates progressive JPEGs: Up to 20% smaller JPEGs with the same image quality.

* [Lanczos](http://en.wikipedia.org/wiki/Lanczos_resampling)-like resampling: Better-looking thumbnails with fewer artifacts.

* Sharpening: Optionally remove some of the blurriness caused by resampling.

* Photo detection: Converts PNG to much smaller JPEGs if it detects that the PNG is a photo.

* Optional WebP: Serve WebP images to capable browsers (Chrome, Android Browser, and Opera) that are 20% smaller than JPEG.

* Metadata stripping: Remove potentially large metadata from each image; particularly useful for images saved by Photoshop.

* Limited input formats: Only accepts common web image formats (JPG, PNG, GIF, and WebP), preventing potential attackers from being able to feed bad data to rarely-used and potentially buggy image parsers.

* Auto-rotation: Camera sensors generally only store photos as landscape, with a header indicating which way it should be rotated when decoded. The rotation is applied and the orientation header reset.

* [Color management aware](http://en.wikipedia.org/wiki/ICC_profile): ICC Color profiles are applied, and colors are converted to match the web-standard sRGB before the profile is removed to save space.  Images won't have perfect fidelity on color-managed workstations, but will be much closer than just stripping the color profiles would be.
