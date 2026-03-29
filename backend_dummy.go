package engine

import "fmt"

type backendDummy struct{}

func (b *backendDummy) Run(w *World) {
	fmt.Println("I'm a dummy backend")
}
