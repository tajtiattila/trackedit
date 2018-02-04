package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/tajtiattila/trackedit/gmap"
)

var mapui *js.Object

func main() {
	jq := jquery.NewJQuery()
	jq.Ready(func() {

		// Create the Google Map using our element and options defined above
		mapui = gmap.New(
			doc.GetElementByID("map"),
			js.M{
				"minZoom":      3,
				"scaleControl": true,
				"styles":       decodeJSON(mapStyles),
			})
		mapui.Call("fitBounds", js.M{
			"north": 48,
			"south": 46,
			"west":  18,
			"east":  20,
		})

		SetupHSplitter(
			doc.GetElementByID("sidebarcontainer"),
			doc.GetElementByID("separator"),
			doc.GetElementByID("map"),
			func() {
				// google.maps.event.trigger(mapui, 'resize')
				js.Global.Get("google").Get("maps").Get("event").Call("trigger", mapui, "resize")
			})

	})
}

// How you would like to style the map.
// This is where you would paste any style found on Snazzy Maps.
var mapStyles = `
  [{"featureType":"landscape","stylers":[{"saturation":-100},{"lightness":65},{"visibility":"on"}]},{"featureType":"poi","stylers":[{"saturation":-100},{"lightness":51},{"visibility":"simplified"}]},{"featureType":"road.highway","stylers":[{"saturation":-100},{"visibility":"simplified"}]},{"featureType":"road.arterial","stylers":[{"saturation":-100},{"lightness":30},{"visibility":"on"}]},{"featureType":"road.local","stylers":[{"saturation":-100},{"lightness":40},{"visibility":"on"}]},{"featureType":"transit","stylers":[{"saturation":-100},{"visibility":"simplified"}]},{"featureType":"administrative.province","stylers":[{"visibility":"off"}]},{"featureType":"water","elementType":"labels","stylers":[{"visibility":"on"},{"lightness":-25},{"saturation":-100}]},{"featureType":"water","elementType":"geometry","stylers":[{"hue":"#ffff00"},{"lightness":-25},{"saturation":-97}]}]
`
