package main

import (
	"fmt"

	. "github.com/ayushmaheshwari768/spartan-go"
)

func main() {
	fakeNet := NewFakeNet(&FakeNet{})
	minnie := NewMiner(&Client{Name: "minnie", Net: fakeNet})
	fmt.Println(minnie.Name)
}
