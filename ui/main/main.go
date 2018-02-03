package main

import "github.com/gopherjs/gopherjs/js"

func main() {
	// Create the Google Map using our element and options defined above
	jsglobal("google.maps.Map").New(
		doc.GetElementByID("map"),
		js.M{
			"minZoom":      3,
			"scaleControl": true,
			"styles":       decodeJSON(mapStyles),
		})
}

// How you would like to style the map.
// This is where you would paste any style found on Snazzy Maps.
var mapStyles = `
  [{"featureType":"landscape","stylers":[{"saturation":-100},{"lightness":65},{"visibility":"on"}]},{"featureType":"poi","stylers":[{"saturation":-100},{"lightness":51},{"visibility":"simplified"}]},{"featureType":"road.highway","stylers":[{"saturation":-100},{"visibility":"simplified"}]},{"featureType":"road.arterial","stylers":[{"saturation":-100},{"lightness":30},{"visibility":"on"}]},{"featureType":"road.local","stylers":[{"saturation":-100},{"lightness":40},{"visibility":"on"}]},{"featureType":"transit","stylers":[{"saturation":-100},{"visibility":"simplified"}]},{"featureType":"administrative.province","stylers":[{"visibility":"off"}]},{"featureType":"water","elementType":"labels","stylers":[{"visibility":"on"},{"lightness":-25},{"saturation":-100}]},{"featureType":"water","elementType":"geometry","stylers":[{"hue":"#ffff00"},{"lightness":-25},{"saturation":-97}]}]
`
