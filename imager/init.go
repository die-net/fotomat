// Copyright 2013-2014 Aaron Hopkins. All rights reserved.
// Use of this source code is governed by the GPL v2 license
// license that can be found in the LICENSE file.

package imager

import (
	"gopkg.in/gographics/imagick.v2/imagick"
)

var white *imagick.PixelWand

func init() {
	imagick.Initialize()
	// imagick.Terminate() is never called. We leak at exit.

	white = imagick.NewPixelWand()
	white.SetColor("white")
}
