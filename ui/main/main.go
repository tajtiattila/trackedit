package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/tajtiattila/track/geomath"
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

		sbc := doc.GetElementByID("sidebar-content")

		go showTrack(sbc, "/api/track")

		doc.Call("addEventListener", "mousedown", clickTrackPt)
	})
}

// How you would like to style the map.
// This is where you would paste any style found on Snazzy Maps.
var mapStyles = `
  [{"featureType":"landscape","stylers":[{"saturation":-100},{"lightness":65},{"visibility":"on"}]},{"featureType":"poi","stylers":[{"saturation":-100},{"lightness":51},{"visibility":"simplified"}]},{"featureType":"road.highway","stylers":[{"saturation":-100},{"visibility":"simplified"}]},{"featureType":"road.arterial","stylers":[{"saturation":-100},{"lightness":30},{"visibility":"on"}]},{"featureType":"road.local","stylers":[{"saturation":-100},{"lightness":40},{"visibility":"on"}]},{"featureType":"transit","stylers":[{"saturation":-100},{"visibility":"simplified"}]},{"featureType":"administrative.province","stylers":[{"visibility":"off"}]},{"featureType":"water","elementType":"labels","stylers":[{"visibility":"on"},{"lightness":-25},{"saturation":-100}]},{"featureType":"water","elementType":"geometry","stylers":[{"hue":"#ffff00"},{"lightness":-25},{"saturation":-97}]}]
`

var appData types.AppData

func showTrack(el *js.Object, u string) {
	removeChildren(el)
	ad, err := fetchAppData(u)
	if err != nil {
		el.Call("appendChild", doc.CreateElement("div", js.M{
			"className": "error",
		}, err.Error()))
		return
	}

	appData = ad

	ents := make([]entry, 0, len(appData.Track)+len(appData.Image))
	for i, p := range appData.Track {
		ents = append(ents, entry{
			t: p.Time,
			v: p,
			i: i,
		})
	}
	for i, im := range appData.Image {
		ents = append(ents, entry{
			t: im.Time,
			v: im,
			i: i,
		})
	}

	sort.Slice(ents, func(i, j int) bool {
		return ents[i].t.Before(ents[j].t)
	})

	var last types.Point
	lastok := false
	for _, e := range ents {
		switch x := e.v.(type) {

		case types.Point:
			var dist string
			if lastok {
				a3 := geomath.Pt3(last.Lat, last.Long)
				b3 := geomath.Pt3(x.Lat, x.Long)
				dist = distm(a3.Sub(b3).Mag())
			} else {
				dist = "start"
			}
			last, lastok = x, true
			div := doc.CreateElement("div", js.M{
				"id":        fmt.Sprintf("trkpt-%d", e.i),
				"className": "trkpt",
			})
			div.Call("appendChild", doc.CreateElement("div", js.M{
				"className": "timestamp",
			}, x.Time.Local().Format("2006-01-02 15:04:05")))
			div.Call("appendChild", doc.CreateElement("div", js.M{
				"className": "distance",
			}, dist))
			el.Call("appendChild", div)

		case types.Image:
			div := doc.CreateElement("div", js.M{
				"className": "photowrap",
			})
			div.Call("appendChild", doc.CreateElement("div", js.M{
				"className": "timestamp",
			}, x.Time.Local().Format("2006-01-02 15:04:05")))
			p := doc.CreateElement("div", js.M{
				"className": "photo",
			})
			p.Call("appendChild", doc.CreateElement("img", js.M{
				"src": x.Thumb,
			}))
			div.Call("appendChild", p)
			el.Call("appendChild", div)
		}
	}
}

type entry struct {
	t time.Time
	v interface{} // types.Track or types.Image
	i int         // index into appData.Track or appData.Image
}

func distm(v float64) string {
	if v < 1000 {
		return fmt.Sprintf("%.0f m", v)
	}
	v /= 1000
	if v < 10 {
		return fmt.Sprintf("%.1f km", v)
	}
	return fmt.Sprintf("%.0f km", v)
}

func fetchAppData(u string) (types.AppData, error) {
	resp, err := http.Get("/api/appdata")
	if err != nil {
		return types.AppData{}, err
	}
	defer resp.Body.Close()

	var d types.AppData
	err = json.NewDecoder(resp.Body).Decode(&d)
	return d, err
}

type MarkerSet struct {
	mapui *js.Object
	o     []*js.Object
}

func (s *MarkerSet) Clear() {
	for _, o := range s.o {
		o.Call("setMap", nil)
	}
	s.o = s.o[:0]
}

func (s *MarkerSet) Add(o *js.Object) {
	s.o = append(s.o, o)
}

func (s *MarkerSet) SimpleMarker(pt types.Point, title string) {
	gm := js.Global.Get("google").Get("maps")
	m := js.M{
		"position": ptLatLong(pt),
		"map":      s.mapui,
	}
	if title != "" {
		m["title"] = title
	}
	s.Add(gm.Get("Marker").New(m))
}

func (s *MarkerSet) Marker(attrs js.M, pt types.Point) {
	gm := js.Global.Get("google").Get("maps")
	m := js.M{
		"position": ptLatLong(pt),
		"map":      s.mapui,
	}
	for k, v := range attrs {
		m[k] = v
	}
	s.Add(gm.Get("Marker").New(m))
}

func (s *MarkerSet) Polyline(attrs js.M, line ...types.Point) {
	gm := js.Global.Get("google").Get("maps")
	pts := make([]js.M, len(line))
	for i, p := range line {
		pts[i] = ptLatLong(p)
	}
	m := js.M{
		"path": pts,
		"map":  s.mapui,
	}
	for k, v := range attrs {
		m[k] = v
	}
	s.Add(gm.Get("Polyline").New(m))
}

var clickMarker MarkerSet

func clickTrackPt(e *js.Object) {
	const pfx = "trkpt-"

	clickMarker.mapui = mapui

	node := e.Get("target")
	for node != nil && !strings.HasPrefix(node.Get("id").String(), pfx) {
		node = node.Get("parentNode")
	}

	if node == nil {
		return
	}

	i, err := strconv.Atoi(strings.TrimPrefix(node.Get("id").String(), pfx))
	if err != nil {
		println(err)
		return
	}

	track := appData.Track

	clickMarker.Clear()

	circleSymbol := js.M{
		"path":          "M 0 0 m -1 0 a 1 1 0 1,0 2 0 a 1 1 0 1, 0 -2 0",
		"scale":         5,
		"fillColor":     "#f00",
		"fillOpacity":   1,
		"strokeOpacity": 0,
	}
	clickMarker.Marker(js.M{
		"icon": circleSymbol,
	}, track[i])

	if i > 0 {
		symbolIncoming := js.M{
			"path":        "M -3,-3 0,3 3,-3 z",
			"fillOpacity": 1,
		}
		a := js.M{
			"icons": []js.M{
				{
					"icon":   symbolIncoming,
					"offset": "20px",
				},
			},
		}
		clickMarker.Polyline(a, track[i], track[i-1])
	}

	if i < len(track)-1 {
		symbolOutgoing := js.M{
			"path":          "M -3,3 0,-3 3,3 z",
			"strokeWeight":  2,
			"strokeOpacity": 1,
		}
		lineSymbol := js.M{
			"path":          "M 0,-1 0,1",
			"strokeOpacity": 1,
			"scale":         4,
		}
		a := js.M{
			"icons": []js.M{
				{
					"icon":   symbolOutgoing,
					"offset": "20px",
				},
				{
					"icon":   lineSymbol,
					"offset": 0,
					"repeat": "20px",
				},
			},
			"strokeOpacity": 0,
		}
		clickMarker.Polyline(a, track[i], track[i+1])
	}

	mapui.Call("panTo", ptLatLong(track[i]))
}

func ptLatLong(p types.Point) js.M {
	return js.M{
		"lat": p.Lat,
		"lng": p.Long,
	}
}
