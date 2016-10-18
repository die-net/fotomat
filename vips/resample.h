#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>


int
cgo_vips_affine(VipsImage *in, VipsImage **out, double a, double b, double c, double d, VipsInterpolate *interpolate) {
    return vips_affine(in, out, a, b, c, d, "interpolate", interpolate, NULL);
}

int
cgo_vips_resize(VipsImage *in, VipsImage **out, double xscale, double yscale) {
#if VIPS_MAJOR_VERSION > 8 || VIPS_MINOR_VERSION >= 4
    return vips_resize(in, out, xscale, "vscale", yscale, "centre", TRUE, NULL);
#else
    return vips_resize(in, out, xscale, "vscale", yscale, NULL);
#endif
}

int
cgo_vips_shrink(VipsImage *in, VipsImage **out, double xshrink, double yshrink) {
    return vips_shrink(in, out, xshrink, yshrink, NULL);
}
