package main

import (
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/matt3o12/gowork"
)

func workon(c *cli.Context) {
	distros, err := gowork.AllDistributors()
	if err != nil {
		fmt.Println("Could not load distros.")
		fmt.Println("Error:", err)
	} else {
		d := make([]string, len(distros))
		for t, dist := range distros {
			d[t] = string(dist)
		}

		fmt.Printf("All available repos: %v\n", strings.Join(d, ", "))
	}
}
