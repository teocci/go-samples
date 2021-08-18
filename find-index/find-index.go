// Package main
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-18
package main

import (
	"fmt"
	"sort"
)

type Person struct {
	Name string
}

func main() {
	crowd := []Person{{"Zoey"}, {"Anna"}, {"Benni"}, {"Chris"}}

	sort.Slice(crowd, func(i, j int) bool {
		return crowd[i].Name <= crowd[j].Name
	})

	needle := "Benni"
	idx := sort.Search(len(crowd), func(i int) bool {
		return string(crowd[i].Name) >= needle
	})

	if crowd[idx].Name == needle {
		fmt.Println("Found:", idx, crowd[idx])
	} else {
		fmt.Println("Found noting: ", idx)
	}
}
