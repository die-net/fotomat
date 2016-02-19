#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>

int 
cgo_vips_gaussblur(VipsImage *in, VipsImage **out, double sigma) {
    return vips_gaussblur(in, out, sigma, NULL);
}

int
cgo_vips_mild_sharpen(VipsImage *in, VipsImage **out) {

    VipsImage *sharpen = vips_image_new_matrixv(3, 3,
        -1.0, -1.0, -1.0,
        -1.0, 32.0, -1.0,
        -1.0, -1.0, -1.0);
    vips_image_set_double(sharpen, "scale", 24.0);
    return vips_conv(in, out, sharpen, NULL);
}

int
cgo_vips_sharpen(VipsImage *in, VipsImage **out, int radius, double x1, double y2, double y3, double m1, double m2) {
    return vips_sharpen(in, out, "radius", radius, "x1", x1, "y2", y2, "y3", y3, "m1", m1, "m2", m2, NULL);
}
