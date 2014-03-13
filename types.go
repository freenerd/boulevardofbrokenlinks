package main

import (
	"fmt"
)

type Aggregates map[string][]Down

type Down struct {
	Origin string
	Url    string
	Status int
}

func (d Down) String() string {
	return fmt.Sprintf("%s: %d %s", d.Origin, d.Status, d.Url)
}
