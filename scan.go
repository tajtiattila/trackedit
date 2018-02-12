package main

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tajtiattila/metadata"
)

type Project struct {
	Root string
	Im   []Img
}

type Img struct {
	Path  string    // relative path
	Thumb string    // db key of thumb
	Time  time.Time // (Exif) time stamp
}

type Store struct {
	db *leveldb.DB
}

func OpenStore(path string) (*Store, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) GetProject(root string, rescan bool) (*Project, error) {
	const pfx = "project:"
	var p Project
	if !rescan {
		raw, err := s.get(pfx + root)
		if err == nil {
			if err := json.Unmarshal(raw, &p); err != nil {
				// non-fatal, rescan
				log.Println(err)
			}
		}
	}

	p.Root = root

	if err := p.scan(s); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(p)
	if err == nil {
		err = s.put(pfx+root, raw)
	}

	return &p, err
}

func (s *Store) get(key string) ([]byte, error) {
	return s.db.Get([]byte(key), nil)
}

func (s *Store) put(key string, value []byte) error {
	return s.db.Put([]byte(key), value, nil)
}

func (p *Project) scan(s *Store) error {
	m := make(map[string]struct{})
	for _, im := range p.Im {
		m[im.Path] = struct{}{}
	}
	return filepath.Walk(p.Root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return nil
		}
		if fi.IsDir() {
			fmt.Println(path)
			return nil
		}

		rel, err := filepath.Rel(p.Root, path)
		if err != nil {
			return err
		}

		if _, ok := m[rel]; ok {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			log.Println(path, err)
			return nil
		}
		defer f.Close()

		m, err := metadata.Parse(f)
		if err != nil {
			if err != metadata.ErrUnknownFormat {
				log.Println(path, err)
			}
			return nil
		}

		t := metadataTime(fi, m)

		r, err := Mkthumb(path, 256, 128)
		if err != nil {
			log.Println(path, err)
			return nil
		}

		raw, err := ioutil.ReadAll(r)
		if err != nil {
			log.Println(path, err)
			return nil
		}

		h := hashBytes(raw)

		s.put("thumb:"+h, raw)
		p.Im = append(p.Im, Img{
			Path:  rel,
			Thumb: h,
			Time:  t,
		})

		return nil
	})
}

func hashBytes(p []byte) string {
	h := sha512.New512_224()
	h.Write(p)
	return "sha512/224-" + hex.EncodeToString(h.Sum(nil))
}

func metadataTime(fi os.FileInfo, m *metadata.Metadata) time.Time {
	to := metadata.ParseTime(m.Get(metadata.DateTimeOriginal)).In(time.Local)
	tc := metadata.ParseTime(m.Get(metadata.DateTimeCreated)).In(time.Local)
	if to.Prec > 0 {
		return to.Time
	}
	if tc.Prec > 0 {
		return tc.Time
	}
	return fi.ModTime()
}
