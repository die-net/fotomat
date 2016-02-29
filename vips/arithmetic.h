#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>

int
cgo_vips_min(VipsImage *in, double *out) {
    return vips_min(in, out, NULL);
}
