#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>

int
cgo_vips_cast(VipsImage *in, VipsImage **out, VipsBandFormat format) {
    return vips_cast(in, out, format, NULL);
}

int
cgo_vips_copy(VipsImage *in, VipsImage **out) {
    return vips_copy(in, out, NULL);
}

int
cgo_vips_embed(VipsImage *in, VipsImage **out, int left, int top, int width, int height, int extend) {
    return vips_embed(in, out, left, top, width, height, "extend", extend, NULL);
}

int
cgo_vips_extract_area(VipsImage *in, VipsImage **out, int left, int top, int width, int height) {
    return vips_extract_area(in, out, left, top, width, height, NULL);
}

int
cgo_vips_extract_band(VipsImage *in, VipsImage **out, int band, int n) {
    return vips_extract_band(in, out, band, "n", n, NULL);
}

double
cgo_max_alpha(VipsImage *in) {
    switch (in->BandFmt) {
    case VIPS_FORMAT_USHORT:
        return 65535;
    case VIPS_FORMAT_FLOAT:
    case VIPS_FORMAT_DOUBLE:
        return 1.0;
    default:
        return 255;
    }
}

int
cgo_vips_flatten(VipsImage *in, VipsImage **out) {
    return vips_flatten(in, out, "max_alpha", cgo_max_alpha(in), NULL);
}

int
cgo_vips_flip(VipsImage *in, VipsImage **out, VipsDirection direction) {
    return vips_flip(in, out, direction, NULL);
}

int
cgo_vips_premultiply(VipsImage *in, VipsImage **out) {
    return vips_premultiply(in, out, "max_alpha", cgo_max_alpha(in), NULL);
}

int
cgo_vips_rot(VipsImage *in, VipsImage **out, VipsAngle angle) {
    return vips_rot(in, out, angle, NULL);
}

int
cgo_vips_unpremultiply(VipsImage *in, VipsImage **out) {
    // Assumes we're converting to uchar and uses default max_alpha of 255.
    return vips_unpremultiply(in, out, NULL);
}
