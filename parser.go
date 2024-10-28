package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"ultratype/models"
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

func parseKind(kindString string) (*models.Type, error) {
	// Remove all spaces
	kindString = strings.ReplaceAll(kindString, " ", "")
	if len(kindString) == 0 {
		return nil, fmt.Errorf("empty type string")
	}

	// Check if nullable
	nullable := false
	if kindString[len(kindString)-1] == '?' {
		nullable = true
		kindString = kindString[:len(kindString)-1]
	}

	// Handle List type
	if strings.HasSuffix(kindString, "[]") {
		baseType, err := parseKind(kindString[:len(kindString)-2])
		if err != nil {
			return nil, err
		}
		return &models.Type{
			Kind:     models.List,
			Nullable: nullable,
			Generics: []models.Type{*baseType},
		}, nil
	}

	// Handle Map type
	if strings.HasPrefix(kindString, "map[") {
		// Find matching closing bracket
		bracketCount := 1
		closingIdx := -1
		for i := 4; i < len(kindString); i++ {
			if kindString[i] == '[' {
				bracketCount++
			} else if kindString[i] == ']' {
				bracketCount--
				if bracketCount == 0 {
					closingIdx = i
					break
				}
			}
		}
		if closingIdx == -1 || closingIdx == len(kindString)-1 {
			return nil, fmt.Errorf("invalid map type format")
		}

		keyType, err := parseKind(kindString[4:closingIdx])
		if err != nil {
			return nil, err
		}

		valueType, err := parseKind(kindString[closingIdx+1:])
		if err != nil {
			return nil, err
		}

		return &models.Type{
			Kind:     models.Map,
			Nullable: nullable,
			Generics: []models.Type{*keyType, *valueType},
		}, nil
	}

	// Handle generic types
	if idx := strings.Index(kindString, "["); idx != -1 {
		if !strings.HasSuffix(kindString, "]") {
			return nil, fmt.Errorf("invalid generic type format")
		}

		baseTypeName := kindString[:idx]
		genericParamsStr := kindString[idx+1 : len(kindString)-1]

		// Split generic parameters
		var genericParams []string
		bracketCount := 0
		current := ""
		for i := 0; i < len(genericParamsStr); i++ {
			if genericParamsStr[i] == '[' {
				bracketCount++
			} else if genericParamsStr[i] == ']' {
				bracketCount--
			}

			if genericParamsStr[i] == ',' && bracketCount == 0 {
				genericParams = append(genericParams, current)
				current = ""
				continue
			}
			current += string(genericParamsStr[i])
		}
		if current != "" {
			genericParams = append(genericParams, current)
		}

		// Parse each generic parameter
		var generics []models.Type
		for _, param := range genericParams {
			genericType, err := parseKind(param)
			if err != nil {
				return nil, err
			}
			generics = append(generics, *genericType)
		}

		// Handle custom generic type
		customName := baseTypeName
		return &models.Type{
			Kind:           models.Custom,
			Nullable:       nullable,
			Generics:       generics,
			CustomTypeName: &customName,
		}, nil
	}

	// Handle basic types
	switch kindString {
	case "int":
		return &models.Type{Kind: models.Integer, Nullable: nullable}, nil
	case "long":
		return &models.Type{Kind: models.Long, Nullable: nullable}, nil
	case "float":
		return &models.Type{Kind: models.Float, Nullable: nullable}, nil
	case "double":
		return &models.Type{Kind: models.Double, Nullable: nullable}, nil
	case "bool":
		return &models.Type{Kind: models.Bool, Nullable: nullable}, nil
	case "string":
		return &models.Type{Kind: models.String, Nullable: nullable}, nil
	default:
		// Custom type
		return &models.Type{
			Kind:           models.Custom,
			Nullable:       nullable,
			CustomTypeName: &kindString,
		}, nil
	}
}

func ParseField(field string) (*models.SchemaField, error) {
	tokens := parseFieldTokens(field)
	f := new(models.SchemaField)
	f.FieldName = tokens[0]
	kind, err := parseKind(tokens[1])
	if err != nil {
		log.Fatal(err)
	}
	f.Kind = *kind

	for i := 2; i < len(tokens); {
		var fieldFor string = tokens[i]
		var fieldValues string = tokens[i+1]

		i += 2

		nullable := fieldFor[len(fieldFor)-1] == '?'
		if nullable {
			fieldFor = fieldFor[:len(fieldFor)-1]
		}
		tagValues := strings.Trim(fieldValues, "\"")

		f.Tags = append(f.Tags, models.Tag{
			TagFor:   fieldFor,
			Nullable: nullable,
			Values:   []string{tagValues},
		})
	}

	return f, nil
}

func ParseFile(path string) (*models.SchemaModel, error) {
	dataBytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	data := strings.TrimSpace(string(dataBytes))
	lines := strings.Split(data, "\n")

	model := new(models.SchemaModel)
	className := strings.TrimSuffix(lines[0], ":")
	model.ClassName = className
	model.Fields = make([]models.SchemaField, 0)

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
