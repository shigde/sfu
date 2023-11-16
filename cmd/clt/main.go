package main

import (
	"fmt"

	"github.com/shigde/sfu/internal/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil && err.Error() != "" {
		fmt.Println(err)
	}
}
