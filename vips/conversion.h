#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>

int
cgo_vips_copy(VipsImage *in, VipsImage **out)
{
    return vips_copy(in, out, NULL);
}

int
cgo_vips_embed(VipsImage *in, VipsImage **out, int left, int top, int width, int height, int extend)
{
    return vips_embed(in, out, left, top, width, height, "extend", extend, NULL);
}

int
cgo_vips_extract_area(VipsImage *in, VipsImage **out, int left, int top, int width, int height)
{
    return vips_extract_area(in, out, left, top, width, height, NULL);
}
