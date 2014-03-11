package main

import (
	"fmt"
)

type Aggregates map[string][]down

type down struct {
	Origin string
	Url    string
	Status int
}

func (d down) String() string {
	return fmt.Sprintf("%s: %d %s", d.Origin, d.Status, d.Url)
}
