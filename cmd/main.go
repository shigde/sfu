package main

import (
	"fmt"

	"github.com/google/uuid"
)

func main() {

	// name := "chiara"

	sagMeinNamen("Enrico")
	fmt.Println(uuid.New().String())

}

func sagMeinNamen(name string) {
	if name == "chiara" {
		fmt.Println(name + " Ich kenne dich!")
	} else {
		fmt.Println(name + " Ich kenne dich nicht. Wer bist du?")
	}

}
