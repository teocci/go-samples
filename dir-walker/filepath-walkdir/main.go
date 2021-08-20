// Package main
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-20
package main

import (
	"io/fs"
	"path/filepath"
)

func walk(path string, f fs.DirEntry, e error) error {
	if e != nil {
		return e
	}
	if f.IsDir() && f.Name() == ".git" {
		println("dir.Name: ", f.Name())
		return filepath.SkipDir
	}
	if !f.IsDir() {
		println(path)
	}
	return nil
}

func main() {
	filepath.WalkDir(".", walk)
}
