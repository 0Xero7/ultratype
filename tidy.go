package main

import (
	"log"
	"os"
)

func Tidy(args []string) {
	filepath := args[2]

	v, err := ParseFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	tidy := v.Pretty()
	os.WriteFile(filepath, []byte(tidy), os.ModePerm)
}
