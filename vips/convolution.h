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
    int e = vips_conv(in, out, sharpen, NULL);
    g_object_unref(sharpen);

    return e;
}

int
cgo_vips_sharpen(VipsImage *in, VipsImage **out, int radius, double x1, double y2, double y3, double m1, double m2) {
    return vips_sharpen(in, out, "radius", radius, "x1", x1, "y2", y2, "y3", y3, "m1", m1, "m2", m2, NULL);
}

int
cgo_sobel(VipsImage *in, VipsImage **out) {
    // TODO: Sobel is sqrt(x^2 + y^2). This is (x/2 + y/2).

    VipsImage *band;
    if (vips_colourspace(in, &band, VIPS_INTERPRETATION_B_W, NULL)) {
        return -1;
    }

    if (vips_image_get_bands(band) > 1) {
        VipsImage *single;
        int e = vips_extract_band(band, &single, 0, NULL);
        g_object_unref(band);
        if (e) {
            return -1;
        }
        band = single;
    }

    VipsImage *x = vips_image_new_matrixv(3, 3,
        -1.0, 0.0, 1.0,
        -2.0, 0.0, 2.0,
        -1.0, 0.0, 1.0);
    vips_image_set_double(x, "scale", 2.0);

    VipsImage *y = vips_image_new_matrixv(3, 3,
        1.0, 2.0, 1.0,
        0.0, 0.0, 0.0,
        -1.0, -2.0, -1.0);
    vips_image_set_double(y, "scale", 2.0);

    VipsImage *sx = NULL;
    VipsImage *sy = NULL;
    VipsImage *ax = NULL;
    VipsImage *ay = NULL;
    VipsImage *add = NULL;

    int ret = -1;
    if (!vips_conv(band, &sx, x, NULL)) {
        if (!vips_abs(sx, &ax, NULL)) {
            if (!vips_conv(band, &sy, y, NULL)) {
                if (!vips_abs(sy, &ay, NULL)) {
                    if (!vips_add(ax, ay, &add, NULL)) {
                        if (!vips_cast(add, out, VIPS_FORMAT_UCHAR, NULL)) {
                            ret = 0;
                        }
                        g_object_unref(add);
                    }
                    g_object_unref(ay);
                }
                g_object_unref(sy);
            }
            g_object_unref(ax);
        }
        g_object_unref(sx);
    }

    g_object_unref(y);
    g_object_unref(x);
    g_object_unref(band);

    return ret;
}

static int
longest_run(VipsImage *in, uint thresh) {
    if (vips_image_wio_input(in) ||
        vips_image_get_bands(in) != 1 ||
        in->BandFmt != VIPS_FORMAT_UINT ||
        in->Ysize != 1 || thresh <= 0) {
        return -1;
    }

    int xsize = in->Xsize;
    int longest = 0;
    int x = 0;
    while (x < xsize) {
        for (; x < xsize && *(uint *)VIPS_IMAGE_ADDR(in, x, 0) < thresh; x++) {
        }
        int first = x;
        for (; x < xsize && *(uint *)VIPS_IMAGE_ADDR(in, x, 0) >= thresh; x++) {
        }
        int length = x - first;
        if (length > longest) {
            longest = length;
        }
    }

    return longest;
}

int
cgo_photo_metric(VipsImage *in, double threshold, int *out) {
    int e = -1;

    VipsImage *copy;
    if (vips_copy(in, &copy, NULL)) {
        return -1;
    }

    VipsImage *sobel;
    if (!cgo_sobel(in, &sobel)) {
        VipsImage *hist = NULL;
        if (!vips_hist_find(sobel, &hist, "band", 0, NULL)) {
            double max;
            if (!vips_max(hist, &max, NULL)) {
                *out = longest_run(hist, (uint)(max * threshold));
                e = 0;
            }
            g_object_unref(hist);
        }
        g_object_unref(sobel);
    }

    g_object_unref(copy);

    return e;
}
