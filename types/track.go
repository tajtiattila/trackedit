package types

import "time"

type Point struct {
	Lat  float64 `json:"latitude"`
	Long float64 `json:"longitude"`

	Time time.Time `json:"time"`
}

type Track []Point
