package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/tajtiattila/trackedit/gmap"
	"github.com/tajtiattila/trackedit/types"
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

		go showTrack(doc.GetElementByID("sidebar-content"), "/api/track")
	})
}

// How you would like to style the map.
// This is where you would paste any style found on Snazzy Maps.
var mapStyles = `
  [{"featureType":"landscape","stylers":[{"saturation":-100},{"lightness":65},{"visibility":"on"}]},{"featureType":"poi","stylers":[{"saturation":-100},{"lightness":51},{"visibility":"simplified"}]},{"featureType":"road.highway","stylers":[{"saturation":-100},{"visibility":"simplified"}]},{"featureType":"road.arterial","stylers":[{"saturation":-100},{"lightness":30},{"visibility":"on"}]},{"featureType":"road.local","stylers":[{"saturation":-100},{"lightness":40},{"visibility":"on"}]},{"featureType":"transit","stylers":[{"saturation":-100},{"visibility":"simplified"}]},{"featureType":"administrative.province","stylers":[{"visibility":"off"}]},{"featureType":"water","elementType":"labels","stylers":[{"visibility":"on"},{"lightness":-25},{"saturation":-100}]},{"featureType":"water","elementType":"geometry","stylers":[{"hue":"#ffff00"},{"lightness":-25},{"saturation":-97}]}]
`

var track types.Track

func showTrack(el *js.Object, u string) {
	removeChildren(el)
	t, err := fetchTrack(u)
	if err != nil {
		el.Call("appendChild", doc.CreateElement("div", js.M{
			"className": "error",
		}, err.Error()))
		return
	}

	for i, p := range t {
		div := doc.CreateElement("div", js.M{
			"id":        fmt.Sprintf("trkpt%d", i),
			"className": "trkpt",
		}, p.Time.Local().Format("2006-01-02 15:04:05"))
		el.Call("appendChild", div)
	}
}

func fetchTrack(u string) (types.Track, error) {
	resp, err := http.Get("/api/track")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var t types.Track
	err = json.NewDecoder(resp.Body).Decode(&t)
	return t, err
}
