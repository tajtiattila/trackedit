package main

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)

// left, split, right are DOM elements
// uses requires https://github.com/lingtalfi/simpledrag.git
func SetupHSplitter(left, split, right *js.Object, cb func()) {
	split.Call("sdrag", func(el *js.Object, px, sx, py, sy int, fix *js.Object) {
		fix.Set("skipX", true)

		// The script below constrains the target to move horizontally between a left and a right virtual boundaries.
		// - the left limit is positioned at 10% of the screen width
		// - the right limit is positioned at 90% of the screen width
		const leftLimit = 10
		const rightLimit = 90

		iw := js.Global.Get("innerWidth").Int()
		if px < iw*leftLimit/100 {
			px = iw * leftLimit / 100
			fix.Set("px", px)
		}
		if px > iw*rightLimit/100 {
			px = iw * rightLimit / 100
			fix.Set("px", px)
		}

		var cur = float64(px) / float64(iw) * 100
		if cur < 0 {
			cur = 0
		}
		if cur > float64(iw) {
			cur = float64(iw)
		}

		//right := (100 - cur - 2)
		ls := left.Get("style")
		ls.Set("width", fmt.Sprintf("%f%%", cur))

		cb()

	}, nil, "horizontal")
}
