package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/tajtiattila/metadata"
	"github.com/tajtiattila/metadata/orient"

	"golang.org/x/image/draw"
)

var magick struct {
	checked bool
	path    string
	broken  bool
}

func hasMagick() bool {
	return false

	if !magick.checked {
		magick.checked = true
		mp := os.Getenv("MAGICK_PATH")
		if mp != "" {
			magick.path = mp
			return true
		}

		mp, err := exec.LookPath("magick")
		if err == nil {
			magick.path = mp
			return true
		}

		log.Println("magick unavailable")
		magick.broken = true
	}

	return !magick.broken
}

func Mkthumb(fn string, maxw, maxh int) (io.ReadCloser, error) {
	if hasMagick() {
		r, err := magickThumb(fn, maxw, maxh)
		if err == nil {
			return r, nil
		}
	}

	return pureThumb(fn, maxw, maxh)
}

func pureThumb(fn string, maxw, maxh int) (io.ReadCloser, error) {
	raw, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	im, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	m, err := metadata.Parse(bytes.NewReader(raw))
	if err != nil && err != metadata.ErrUnknownFormat {
		log.Println(fn, err)
	}

	var w, h int
	if m != nil && orient.IsTranspose(m.Orientation) {
		// image is going to be transposed,
		// swap target width/height
		w, h = thumbSize(im, maxh, maxw)
	} else {
		w, h = thumbSize(im, maxw, maxh)
	}
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))

	draw.Draw(dst, dst.Bounds(), image.White, image.ZP, draw.Src)
	draw.BiLinear.Scale(dst, dst.Bounds(), im, im.Bounds(), draw.Src, nil)

	var thumb image.Image = dst
	if m != nil {
		thumb = orient.Orient(thumb, m.Orientation)
	}

	b := new(bytes.Buffer)
	err = jpeg.Encode(b, thumb, nil)

	return ioutil.NopCloser(bytes.NewReader(b.Bytes())), err
}

func thumbSize(im image.Image, maxw, maxh int) (w, h int) {
	s := im.Bounds().Size()
	if s.X <= maxw && s.Y <= maxh {
		return s.X, s.Y
	}
	w = s.X * maxh / s.Y
	if w <= maxw {
		return w, maxh
	}
	h = s.Y * maxw / s.X
	return maxw, h
}

func magickThumb(fn string, maxw, maxh int) (io.ReadCloser, error) {
	cmd := exec.Command(magick.path, "convert",
		"-auto-orient",
		"-thumbnail", fmt.Sprintf("%dx%d", maxw, maxh),
		"jpg:-")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		log.Println("magick failed with", err)
		magick.broken = true
		return nil, err
	}
	buf, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(buf)), nil
}
