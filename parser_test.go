package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProgram_LoadFile(t *testing.T) {
	tests := []struct {
		Name string
		Pro  *Program
		Path string
	}{
		{
			Name: "custom code",
			Pro: NewProgram(map[string]string{
				"hello": `
			package main
			import "math"
			func main() {
				var _ = 2 * math.Pi
			}`,
				"math": `
			package math
			const Pi = 3.1415926`,
			}),
			Path: "math",
		}, {
			Name: "a.go",
			Pro:  NewProgram(nil),
			Path: "./a/a.go",
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			pkg, f, err := tt.Pro.LoadFile(tt.Path)
			_, _ = pkg, f
			assert.NoError(t, err)
		})
	}
}

func TestProgram_LoadPackage(t *testing.T) {
	tests := []struct {
		Name string
		Pro  *Program
		Path string
	}{
		{
			Name: "service",
			Pro:  NewProgram(nil),
			Path: "./test_data/service",
		}, {
			Name: "multi_pkg",
			Pro:  NewProgram(nil),
			Path: "./test_data/multi_pkg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			pkg, err := tt.Pro.LoadPackage(tt.Path)
			_ = pkg
			assert.NoError(t, err)
		})
	}
}
