package types

import "time"

type Image struct {
	Thumb string    // thumb image src
	Time  time.Time // (Exif) time stamp
}
