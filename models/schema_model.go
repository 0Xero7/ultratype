package models

import (
	"fmt"
	"slices"
	"strings"
	"ultratype/utils"
)

type Tag struct {
	TagFor   string
	Nullable bool
	Values   []string
}

func (t *Tag) String(maybeNullable bool) string {
	f := t.TagFor
	if t.Nullable {
		f += "?"
	} else if maybeNullable {
		f += " "
	}

	f += ":" + "\"" + strings.Join(t.Values, ",") + "\""
	return f
}

type InternalType int

const (
	Integer InternalType = iota
	Long
	Float
	Double
	Bool
	String

	List
	Map

	Custom
)

type Type struct {
	Kind           InternalType
	Nullable       bool
	Generics       []Type
	CustomTypeName *string
}

func (t Type) String() string {
	v := ""

	switch t.Kind {
	case Integer:
		v = "int"
	case Long:
		v = "long"
	case Float:
		v = "float"
	case Double:
		v = "double"
	case Bool:
		v = "bool"
	case String:
		v = "string"
	case List:
		v = t.Generics[0].String() + "[]"
	case Map:
		v = "map[" + t.Generics[0].String() + "]" + t.Generics[1].String()
	case Custom:
		v = *t.CustomTypeName
	}

	if t.Nullable {
		v += "?"
	}

	return v
}

type SchemaField struct {
	FieldName string
	Kind      Type
	Tags      []Tag
}

type SchemaModel struct {
	ClassName string
	Fields    []SchemaField
}

func (s *SchemaModel) Pretty() string {
	lines := make([]string, 0)
	lines = append(lines, s.ClassName+":")

	names := make([]string, 0)
	types := make([]string, 0)
	fieldTags := make(map[string][]string)
	tags := make([]string, 0)
	hasNullVariant := make(map[string]bool)
	byTags := make(map[string]int)

	longestName := 0
	longestType := 0

	for _, field := range s.Fields {
		names = append(names, field.FieldName)
		types = append(types, field.Kind.String())

		longestName = max(longestName, len(field.FieldName))
		longestType = max(longestType, len(field.Kind.String()))

		for _, tag := range field.Tags {
			forTag := tag.TagFor
			tags = append(tags, forTag)
			fieldTags[field.FieldName] = append(fieldTags[field.FieldName], forTag)

			if tag.Nullable {
				hasNullVariant[forTag] = true
			}
		}
	}

	for _, field := range s.Fields {
		for _, tag := range field.Tags {
			forTag := tag.TagFor
			byTags[forTag] = max(byTags[forTag], len(tag.String(hasNullVariant[forTag])))
		}
	}

	slices.Sort(tags)
	tags = slices.Compact(tags)

	for _, field := range s.Fields {
		f := fmt.Sprintf("  %s %s", utils.PadRight(field.FieldName, longestName), utils.PadRight(field.Kind.String(), longestType))

		for _, tag := range tags {
			if slices.Contains(fieldTags[field.FieldName], tag) {
				t := slices.IndexFunc(field.Tags, func(tx Tag) bool {
					return tx.TagFor == tag
				})
				f += "  " + utils.PadRight(field.Tags[t].String(hasNullVariant[tag]), byTags[tag])
			} else {
				f += "  " + utils.PadRight("-", byTags[tag])
			}
		}

		lines = append(lines, f)
	}

	return strings.Join(lines, "\n")
}
