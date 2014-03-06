package imager

import (
        "github.com/die-net/fotomat/imager/imagick"
)

func init() {
	imagick.Initialize()
	// imagick.Terminate() is never called. We leak at exit.
}
