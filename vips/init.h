#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>

int
cgo_vips_warning_callback(char const* log_domain, GLogLevelFlags log_level, char const* message, void* ignore) {
   // Do nothing
}

int
cgo_vips_init() {
    g_log_set_handler("VIPS", G_LOG_LEVEL_WARNING, (GLogFunc)cgo_vips_warning_callback, NULL);
    return VIPS_INIT("fotomat");
}
