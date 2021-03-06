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

int
cgo_vips_icc_transform(VipsImage *in, VipsImage **out, const char *output_profile, const char *input_profile, VipsIntent intent) {
    if (input_profile != NULL) {
        return vips_icc_transform(in, out, output_profile, "intent", intent, "input_profile", input_profile, "embedded", FALSE, NULL);
    }
    return vips_icc_transform(in, out, output_profile, "intent", intent, "embedded", TRUE, NULL);
}
