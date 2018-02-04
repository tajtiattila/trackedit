package main

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)

// leftPane, splitter, rightPane are DOM elements
func SetupHSplitter(leftPane, splitter, rightPane *js.Object, cb func()) {
	var dragging bool
	var dx int

	move := func(e *js.Object) {
		const leftLimit = 10
		const rightLimit = 90

		px := e.Get("pageX").Int()
		iw := js.Global.Get("innerWidth").Int()

		var cur = float64(px-dx) / float64(iw) * 100
		if cur < leftLimit {
			cur = leftLimit
		}
		if cur > rightLimit {
			cur = rightLimit
		}

		//right := (100 - cur - 2)
		ls := leftPane.Get("style")
		ls.Set("width", fmt.Sprintf("%f%%", cur))

		cb()
	}

	startDragging := func(e *js.Object) {
		dragging = true
		left := splitter.Call("getBoundingClientRect").Get("left").Int()
		cx := e.Get("clientX").Int()
		dx = cx - left
		js.Global.Call("addEventListener", "mousemove", move)
	}

	splitter.Call("addEventListener", "mousedown", startDragging)
	js.Global.Call("addEventListener", "mouseup", func(e *js.Object) {
		if dragging {
			dragging = false
			js.Global.Call("removeEventListener", "mousemove", move)
		}
	})
}
