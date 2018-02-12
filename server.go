package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/tajtiattila/trackedit/types"
)

func serveAppData(t types.Track, p *Project) http.Handler {
	var images []types.Image
	if p != nil {
		for _, im := range p.Im {
			images = append(images, types.Image{
				Thumb: path.Join("thumb", im.Thumb),
				Time:  im.Time,
			})
		}
	}
	data := types.AppData{
		Track: t,
		Image: images,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			log.Println(err)
		}
	})
}

func serveTrack(t types.Track) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		err := json.NewEncoder(w).Encode(t)
		if err != nil {
			log.Println(err)
		}
	})
}

func serveThumbs(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := s.get("thumb:" + r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(b))
	})
}

type imByTime []Img

func (t imByTime) Len() int           { return len(t) }
func (t imByTime) Less(i, j int) bool { return t[i].Time.Before(t[j].Time) }
func (t imByTime) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
