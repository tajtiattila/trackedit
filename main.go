package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var addr, ui string
	flag.StringVar(&addr, "addr", ":7267", "listen address")
	flag.StringVar(&ui, "ui", filepath.Join(modulePath(), "ui"), "ui directory")
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("need exactly 1 track file argument")
	}
	trk := loadTrack(flag.Arg(0))

	td := struct {
		GoogleMapsAPIKey string
	}{
		GoogleMapsAPIKey: os.Getenv("GOOGLEMAPS_APIKEY"),
	}
	http.Handle("/", http.FileServer(&templateDir{ui, td}))
	http.Handle("/api/track", serveTrack(trk))

	verify(serveGopherJS(54321, stripGoSrcPath(ui), "main"))

	log.Println("listening on", addr)
	go func() {
		time.Sleep(time.Second)
		logErr(openbrowser(addr))
	}()
	logErr(http.ListenAndServe(addr, nil))
}
