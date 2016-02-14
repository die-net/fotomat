# Fotomat VIPS Wrapper

A simple cgo wrapper around the C-based [libvips](http://www.vips.ecs.soton.ac.uk/index.php?title=Libvips) that handles the conversion from Go semantics to C semantics.

To the greatest extent possible, vips function, argument, and constant names have been preserved.  However, the "vips" prefix has been removed, since that's covered by the package name.  And in accordance with Go style, Go method and variable names use CamelCase instead of snake_case.

#### Limitations

Many vips functions take optional arguments via the C varargs mechanism, which isn't compatible with Go. To work around this, there are .h header files that expose a cgo_vips_*() version of the function that uses a fixed list of parameters.
