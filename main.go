package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/tajtiattila/basedir"
)

func main() {
	addr := flag.String("addr", ":7267", "listen address")
	rescan := flag.Bool("rescan", false, "rescan project at startup")
	ui := flag.String("ui", filepath.Join(modulePath(), "ui"), "ui directory")
	img := flag.String("img", "", "optional photo directory")
	flag.Parse()

	var project *Project
	if *img != "" {
		cacheDir, err := basedir.Cache.EnsureDir("phototrack", 0666)
		verify(err)

		s, err := OpenStore(filepath.Join(cacheDir, "index.leveldb"))
		verify(err)

		project, err = s.GetProject(filepath.Clean(*img), *rescan)
		verify(err)

		http.Handle("/thumb/", http.StripPrefix("/thumb/", serveThumbs(s)))
	}

	if flag.NArg() != 1 {
		log.Fatal("need exactly 1 track file argument")
	}
	trk := loadTrack(flag.Arg(0))

	td := struct {
		GoogleMapsAPIKey string
	}{
		GoogleMapsAPIKey: os.Getenv("GOOGLEMAPS_APIKEY"),
	}
	http.Handle("/", http.FileServer(&templateDir{*ui, td}))
	http.Handle("/api/appdata", serveAppData(trk, project))

	verify(serveGopherJS(54321, stripGoSrcPath(*ui), "main"))

	log.Println("listening on", *addr)
	go func() {
		time.Sleep(time.Second)
		logErr(openbrowser(*addr))
	}()
	logErr(http.ListenAndServe(*addr, nil))
}
