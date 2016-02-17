#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>

int
cgo_vips_image_get_as_string(const VipsImage *image, const char *field, const char **out )
{
    if (vips_image_get_typeof(image, field) && !vips_image_get_string(image, field, out)) {
        return 0;
    }
    return -1;
}
