#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>

int
cgo_vips_colourspace(VipsImage *in, VipsImage **out, VipsInterpretation space) {
    return vips_colourspace(in, out, space, NULL);
}

int
cgo_vips_icc_import(VipsImage *in, VipsImage **out) {
    return vips_icc_import(in, out, "embedded", TRUE, NULL);
}
