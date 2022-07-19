package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProgram_LoadPackage(t *testing.T) {
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
			pkg, f, err := tt.Pro.LoadPackage(tt.Path)
			_, _ = pkg, f
			assert.NoError(t, err)
			for i := 0; i < pkg.Scope().NumChildren(); i++ {
				fmt.Println(pkg.Scope().Child(i).String())
			}
		})
	}
}
