#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>

int 
cgo_vips_gaussblur(VipsImage *in, VipsImage **out, double sigma) {
    return vips_gaussblur(in, out, sigma, NULL);
}

int
cgo_vips_sharpen(VipsImage *in, VipsImage **out, int radius, double m1, double m2) {
    return vips_sharpen(in, out, "radius", radius, "m1", m1, "m2", m2, NULL);
}
