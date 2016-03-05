Usage
=====

Command-line flags:
------------------

When using the fotomat server, options affecting how the server behaves and resources it will eat:

```
-listen string
    [IP]:port to listen for incoming connections.
    (default "127.0.0.1:3520")
-fetch_timeout duration
    How long to wait to receive original image from source (0=disable).
    (default 30s)
-local_image_directory string
    Enable local image serving from this path ("" = proxy instead).
-max_connections int
    The maximum number of incoming connections allowed.
    (defaults to maximum allowed by OS)
-max_image_threads int
    Maximum number of threads simultaneously processing images.
    (defaults to one thread per available CPU)
-max_processing_duration duration
    Maximum duration we can be processing an image before assuming we crashed
    (0=disable). (default 1m0s)
```

And controlling the generated images:

```
-always_interpolate (default false)
    Always use slower high-quality interpolator for final 2x shrink.
-lossless
    Allow saving as PNG even without transparency. (default true)
-lossless_webp
    When saving in WebP, allow lossless encoding.
-lossy_if_photo
    Save as lossy if image is detected as a photo. (default true)
-max_buffer_pixels int
    Maximum number of pixels to allocate for an intermediate image buffer.
    (default 6500000)
-max_output_dimension int
    Maximum width or height of an image response. (default 2048)
-sharpen
    Sharpen after resize. (default false)
```

Notes:

* Listening on IPv4 localhost on port 3520. Specify ```-listen=:3520``` to listen for remote connections. IPv6 is supported.

* Proxy mode, where the image is fetched from the host supplied in the Host header via http port 80. If you want to disable proxy mode and serve files from a local directory instead, pass ```-local_image_directory=/some/path```.

* Only allocating image buffers that are at most 6,500,000 pixels (width * height). It can read larger JPEGs than this because it scale them down by a factor of 8 when decoding.

* Allowing as many VIPS threads to be running as the machine has CPU cores. Raising this probably won't increase throughput, but lowering it may reduce memory usage.

* Allowing output images to be up to 2048 x 2048. Raising this will allow larger images, eat more RAM, and be slower.

* Limiting a single VIPS operation to 1 minute, after which it assumes it has hit a VIPS bug and crashes the process.  Raise this if actual image operations take longer.
