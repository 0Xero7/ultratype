package main

import (
	"fmt"
	"slices"
	"strings"
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

type SchemaField struct {
	FieldName string
	Type      string
	Tags      []Tag
}

type SchemaModel struct {
	ClassName string
	Fields    []SchemaField
}

func padRight(s string, width int) string {
	k := string(s)
	for len(k) < width {
		k += " "
	}
	return k
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
		types = append(types, field.Type)

		longestName = max(longestName, len(field.FieldName))
		longestType = max(longestType, len(field.Type))

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
		f := fmt.Sprintf("  %s %s", padRight(field.FieldName, longestName), padRight(field.Type, longestType))

		for _, tag := range tags {
			if slices.Contains(fieldTags[field.FieldName], tag) {
				t := slices.IndexFunc(field.Tags, func(tx Tag) bool {
					return tx.TagFor == tag
				})
				f += "  " + padRight(field.Tags[t].String(hasNullVariant[tag]), byTags[tag])
			} else {
				f += "  " + padRight("-", byTags[tag])
			}
		}

		lines = append(lines, f)
	}

	return strings.Join(lines, "\n")
}
