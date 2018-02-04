package main

import (
	"html"
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

// decodeJSON decodes json and returns the javascript object
// it panics if json is invalid
func decodeJSON(json string) *js.Object {
	return js.Global.Get("JSON").Call("parse", json)
}

var doc = &Document{
	Object: js.Global.Get("document"),
}

type Document struct {
	*js.Object
}

func (d *Document) CreateElement(name string, attr js.M, content ...string) *js.Object {
	e := d.Call("createElement", name)
	for k, v := range attr {
		e.Set(k, v)
	}
	if len(content) != 0 {
		e.Set("innerHTML", html.EscapeString(strings.Join(content, "")))
	}
	return e
}

func (d *Document) GetElementByID(id string) *js.Object {
	return d.Call("getElementById", id)
}
