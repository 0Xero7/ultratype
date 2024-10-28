package main

import (
	"fmt"
	"log"
	"ultratype/generators"
)

func Generate(args []string) {
	language := args[2]
	filepath := args[3]

	model, err := ParseFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	generator := generators.GetGenerator(language)
	out, _ := generator.Generate(model)
	fmt.Println(*out)
}
