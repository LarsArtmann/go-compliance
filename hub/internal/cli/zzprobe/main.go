package main

import (
	"context"
	"fmt"
	"os"

	"github.com/larsartmann/go-compliance/hub/internal/discovery"
)

func main() {
	c := discovery.New(discovery.ResolveSocketPath())
	p, err := c.Discover(context.Background(), []string{os.Getenv("HOME") + "/projects"})
	fmt.Println(len(p), err)
	n := 0
	for _, x := range p {
		if _, e := os.Stat(x.Path + "/go.mod"); e != nil {
			n++
			if n <= 6 {
				fmt.Printf("%+v\n", x)
			}
		}
	}
	fmt.Println("no go.mod:", n)
}
