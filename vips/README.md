# Fotomat VIPS Wrapper

A simple cgo wrapper around the C-based
[libvips](http://www.vips.ecs.soton.ac.uk/index.php?title=Libvips) that
handles the conversion from Go semantics to C semantics.

To the greatest extent possible, VIPS function, argument, and constant names
have been preserved.  However, the "vips" prefix has been removed, since
that's covered by the package name.  And in accordance with Go style, Go
method and variable names use CamelCase instead of snake_case.

Also see:

* [Godoc API documentation](https://godoc.org/github.com/die-net/fotomat/vips) for this API
* The original [libvips API documentation](http://www.vips.ecs.soton.ac.uk/supported/current/doc/html/libvips/index.html)
* Fotomat's [format](https://github.com/die-net/fotomat/tree/master/format) and [thumbnail](https://github.com/die-net/fotomat/tree/master/thumbnail) libraries and the Fotomat [server](https://github.com/die-net/fotomat), which use this API.

#### Limitations

Only the subset of the VIPS functionality that Fotomat uses is currently
supported.

Many VIPS functions take optional arguments via the C varargs mechanism,
which isn't compatible with cgo.  To work around this, there are .h header
files that expose a cgo_vips_*() version of the function that uses a fixed
list of parameters.
