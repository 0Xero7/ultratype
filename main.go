package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Need atleast 2 args.")
	}

	switch os.Args[1] {
	case "tidy":
		Tidy(os.Args)

	case "gen":
		Generate(os.Args)
	}
}
