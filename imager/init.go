package imager

import (
        "github.com/gographics/imagick/imagick"    
)

func init() {
	imagick.Initialize()
	// imagick.Terminate() is never called. We leak at exit.
}
