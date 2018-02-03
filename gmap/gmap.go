package gmap

import "github.com/gopherjs/gopherjs/js"

// New creates a new google.maps.Map object
func New(mapElement *js.Object, attr js.M) *js.Object {
	return gmo("Map").New(mapElement, attr)
}

func NewMapObj(name string, attrs js.M) *js.Object {
	cls := js.Global.Get("google").Get("maps").Get(name)
	return cls.New(attrs)
}

func gmo(name string) *js.Object {
	return js.Global.Get("google").Get("maps").Get(name)
}
