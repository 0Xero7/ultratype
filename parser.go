package main

import (
	"log"
	"os"
	"regexp"
	"strings"
)

func parseFieldTokens(field string) []string {
	tokens := make([]string, 0)
	field = strings.TrimSpace(field)

	fieldNameEndIndex := strings.IndexAny(field, "\t ")
	tokens = append(tokens, field[:fieldNameEndIndex])
	field = strings.TrimSpace(field[fieldNameEndIndex+1:])

	// TODO: have proper type parsing if needed
	fieldTypeEndIndex := strings.IndexAny(field, "\t ")
	if fieldTypeEndIndex == -1 {
		fieldTypeEndIndex = len(field)
	}
	tokens = append(tokens, field[:fieldTypeEndIndex])
	field = strings.TrimSpace(field[fieldTypeEndIndex+1:])

	pattern := regexp.MustCompile("^[ \t]*[a-zA-Z]+[a-zA-Z_0-9\\-]*\\??[ \t]*:[ \t]*\"[^\"]*\"")

	for len(field) > 0 {
		if field[0] == '-' {
			field = strings.TrimSpace(field[1:])
			continue
		}

		tagIndex := pattern.FindStringIndex(field)
		if tagIndex == nil {
			break
		}

		tagString := strings.TrimSpace(field[tagIndex[0]:tagIndex[1]])

		splits := strings.Split(tagString, ":")
		tagFor := strings.TrimSpace(splits[0])
		tagValue := strings.TrimSpace(strings.Join(splits[1:], ""))
		tokens = append(tokens, tagFor, tagValue)

		if tagIndex[1]+1 < len(field) {
			field = strings.TrimSpace(field[tagIndex[1]+1:])
		} else {
			break
		}
	}

	return tokens
}

func ParseField(field string) (*SchemaField, error) {
	tokens := parseFieldTokens(field)
	f := new(SchemaField)
	f.FieldName = tokens[0]
	f.Type = tokens[1]

	for i := 2; i < len(tokens); {
		var fieldFor string = tokens[i]
		var fieldValues string = tokens[i+1]

		i += 2

		nullable := fieldFor[len(fieldFor)-1] == '?'
		if nullable {
			fieldFor = fieldFor[:len(fieldFor)-1]
		}
		tagValues := strings.Trim(fieldValues, "\"")

		f.Tags = append(f.Tags, Tag{
			TagFor:   fieldFor,
			Nullable: nullable,
			Values:   []string{tagValues},
		})
	}

	return f, nil
}

func ParseFile(path string) (*SchemaModel, error) {
	dataBytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	data := strings.TrimSpace(string(dataBytes))
	lines := strings.Split(data, "\n")

	model := new(SchemaModel)
	className := strings.TrimSuffix(lines[0], ":")
	model.ClassName = className
	model.Fields = make([]SchemaField, 0)

	for _, fieldLine := range lines[1:] {
		field, err := ParseField(fieldLine)
		if err != nil {
			log.Fatal(err)
		}
		if field == nil {
			continue
		}

		model.Fields = append(model.Fields, *field)
	}

	return model, nil
}
