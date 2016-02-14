#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>

int
cgo_vips_jpegload_buffer(void *buf, size_t len, VipsImage **out, int shrink)
{
    return vips_jpegload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, "shrink", shrink, NULL);
};

int
cgo_vips_jpegsave_buffer(VipsImage *in, void **buf, size_t *len, int strip, int q, int optimize_coding, int interlace)
{
    return vips_jpegsave_buffer(in, buf, len, "strip", strip, "Q", q, "optimize_coding", optimize_coding, "interlace", interlace, NULL);
}

int
cgo_vips_pngload_buffer(void *buf, size_t len, VipsImage **out)
{
    return vips_pngload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
};


int
cgo_vips_pngsave_buffer(VipsImage *in, void **buf, size_t *len, int compression, int interlace)
{
    return vips_pngsave_buffer(in, buf, len, "compression", compression, "interlace", interlace, NULL);
}

int
cgo_vips_webpload_buffer(void *buf, size_t len, VipsImage **out)
{
    return vips_webpload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
};

int
cgo_vips_webpsave_buffer(VipsImage *in, void **buf, size_t *len, int q)
{
    return vips_webpsave_buffer(in, buf, len, "Q", q, NULL);
}
