// Package bench_walkdir
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-20
package bench_walkdir

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/karrick/godirwalk"
)

const benchRoot = "D:/Temp/go-samples"

var scratch []byte
var largeDirectory string

func init() {
	scratch = make([]byte, godirwalk.MinimumScratchBufferSize)
	largeDirectory = filepath.Join(benchRoot, "tmp")
}

func BenchmarkGodirwalk(b *testing.B) {
	tests := []struct {
		desc string
		fn   func(*testing.B)
	}{
		{
			desc: "ReadDirents",
			fn: func(b *testing.B) {
				_, err := godirwalk.ReadDirents(largeDirectory, scratch)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
		{
			desc: "ReadDirnames",
			fn: func(b *testing.B) {
				_, err := godirwalk.ReadDirnames(largeDirectory, scratch)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
	}

	for _, t := range tests {
		b.Run(t.desc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				t.fn(b)
			}
		})
	}
}

func BenchmarkWalkDirectory(b *testing.B) {
	tests := []struct {
		desc string
		fn   func(*testing.B)
	}{
		{
			desc: "filepath.WalkDir",
			fn: func(b *testing.B) {
				err := filepath.WalkDir(benchRoot, walk)
				if err != nil {
					b.Error(err)
				}
			},
		},
		{
			desc: "godirwalk.Walk Unsorted",
			fn: func(b *testing.B) {
				err := godirwalk.Walk(benchRoot, &godirwalk.Options{
					Callback: walkCallback,
					Unsorted:      true,
				})
				if err != nil {
					b.Errorf("GOT: %v; WANT: nil", err)
				}
			},
		},
		{
			desc: "godirwalk.Walk Sorted",
			fn: func(b *testing.B) {
				err := godirwalk.Walk(benchRoot, &godirwalk.Options{
					Callback: walkCallback,
					Unsorted:      false,
				})
				if err != nil {
					b.Errorf("GOT: %v; WANT: nil", err)
				}
			},
		},
		{
			desc: "godirwalk.Walk Unsorted Scratch",
			fn: func(b *testing.B) {
				err := godirwalk.Walk(benchRoot, &godirwalk.Options{
					Callback: walkCallback,
					ScratchBuffer: scratch,
					Unsorted:      true,
				})
				if err != nil {
					b.Errorf("GOT: %v; WANT: nil", err)
				}
			},
		},
		{
			desc: "godirwalk.Walk Sorted Scratch",
			fn: func(b *testing.B) {
				err := godirwalk.Walk(benchRoot, &godirwalk.Options{
					Callback: walkCallback,
					ScratchBuffer: scratch,
					Unsorted:      false,
				})
				if err != nil {
					b.Errorf("GOT: %v; WANT: nil", err)
				}
			},
		},
	}

	for _, t := range tests {
		b.Run(t.desc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				t.fn(b)
			}
		})
	}
}

func walkCallback(name string, f *godirwalk.Dirent) error {
	if f.IsDir() && f.Name() == ".git" {
		return filepath.SkipDir
	}
	return nil
}

func walk(path string, f fs.DirEntry, e error) error {
	if e != nil {
		return e
	}
	if f.IsDir() && f.Name() == ".git" {
		return filepath.SkipDir
	}
	return nil
}