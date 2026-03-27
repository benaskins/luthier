package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: luthier <prd.md>")
		os.Exit(1)
	}
	fmt.Println("luthier: not yet implemented")
}
