package main

import (
	"fmt"
)

type Down struct {
	Origin string
	Url    string
	Status int
}

func (d Down) String() string {
	return fmt.Sprintf("%s: %d %s", d.Origin, d.Status, d.Url)
}

type Downs chan Down
type Checked chan Downs
