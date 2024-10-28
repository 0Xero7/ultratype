package generators

import (
	"fmt"
	"slices"
	"strings"
	"ultratype/models"
	"ultratype/utils"
)

type GoLang struct {
	allTags []string
}

func (_ GoLang) GenerateField(field *models.SchemaField) any {
	return ""
}

func createGoType(t *models.Type) string {
	v := ""

	switch t.Kind {
	case models.Integer:
		v = "int"
	case models.Long:
		v = "long"
	case models.Float:
		v = "float"
	case models.Double:
		v = "double"
	case models.Bool:
		v = "bool"
	case models.String:
		v = "string"
	case models.List:
		v = createGoType(&t.Generics[0]) + "[]"
	case models.Map:
		v = "map[" + createGoType(&t.Generics[0]) + "]" + createGoType(&t.Generics[1])
	case models.Custom:
		v = *t.CustomTypeName
	}

	return v
}

func (g GoLang) GenerateType(kind *models.Type) any {
	return createGoType(kind)
}

func (g GoLang) GenerateTag(tags []models.Tag) any {
	tagStrings := make([]string, 0)
	tagsFound := make([]string, 0)
	for _, tag := range tags {
		item := tag.TagFor + ":\""
		tagsFound = append(tagsFound, tag.TagFor)
		tagValues := tag.Values
		if tag.Nullable {
			tagValues = append(tagValues, "omitempty")
		}
		item += strings.Join(tagValues, ",")
		item += "\""
		tagStrings = append(tagStrings, item)
	}

	for _, v := range g.allTags {
		if !slices.Contains(tagsFound, v) {
			tagStrings = append(tagStrings, v+":\"-\"")
		}
	}

	return fmt.Sprintf("`%s`", strings.Join(tagStrings, " "))
}

func (g GoLang) Generate(model *models.SchemaModel) (*string, error) {
	lines := make([]string, 0)
	lines = append(lines, fmt.Sprintf("type %s struct {", model.ClassName))

	names := make([]string, 0)
	maxNameLength := 0

	types := make([]string, 0)
	maxTypeLength := 0

	tags := make([]string, 0)
	g.allTags = make([]string, 0)

	// Collect all the tags
	for _, field := range model.Fields {
		for _, tag := range field.Tags {
			g.allTags = append(g.allTags, tag.TagFor)
		}
	}
	slices.Sort(g.allTags)
	g.allTags = slices.Compact(g.allTags)

	for _, field := range model.Fields {
		names = append(names, field.FieldName)
		maxNameLength = max(maxNameLength, len(field.FieldName))

		types = append(types, g.GenerateType(&field.Kind).(string))
		maxTypeLength = max(maxTypeLength, len(types[len(types)-1]))

		tags = append(tags, g.GenerateTag(field.Tags).(string))
	}

	for _, field := range model.Fields {
		lines = append(lines, "    "+utils.PadRight(field.FieldName, maxNameLength)+" "+utils.PadRight(g.GenerateType(&field.Kind).(string), maxTypeLength)+" "+g.GenerateTag(field.Tags).(string))
	}

	lines = append(lines, "}")

	output := strings.Join(lines, "\n")
	return &output, nil
}
