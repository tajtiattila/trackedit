package main

import (
	"os"

	"github.com/tajtiattila/track/trackio"
	"github.com/tajtiattila/trackedit/types"
)

func loadTrack(fn string) types.Track {
	f, err := os.Open(fn)
	verify(err)
	defer f.Close()

	tx, err := trackio.NewDecoder(f).Track()
	verify(err)

	tx = dedupTrack(tx)

	t := make(types.Track, len(tx))
	for i, p := range tx {
		t[i] = types.Point{
			Lat:  p.Lat,
			Long: p.Long,
			Time: p.Time,
		}
	}

	return t
}

func dedupTrack(t trackio.Track) trackio.Track {
	i := 1
	for _, p := range t[1:] {
		q := t[i-1]
		if p.Lat != q.Lat || p.Long != q.Long || p.Time != q.Time {
			t[i], i = p, i+1
		}
	}
	return t[:i]
}
