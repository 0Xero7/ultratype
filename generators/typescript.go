package generators

import (
	"fmt"
	"slices"
	"strings"
	"ultratype/models"
)

type TypeScript struct {
	allTags []string
}

func (_ TypeScript) GetTypeDefaultInitialization(t *models.Type) string {
	if t.Nullable {
		return "null"
	}

	switch t.Kind {
	case models.Integer, models.Long, models.Float, models.Double:
		return "0"
	case models.Bool:
		return "false"
	case models.String:
		return "\"\""
	case models.List:
		return "[]"
	case models.Map:
		return "new Map<" + createTypeScriptType(&t.Generics[0]) + ", " + createTypeScriptType(&t.Generics[1]) + ">()"
	case models.Custom:
		// params := make([]string, 0)
		return "NOT IMPLEMENTED!!!"
	}

	panic("Unknown type!")
}

// func (_ TypeScript) GetDefaultInitialization(field *models.SchemaField) string {

// }

func (_ TypeScript) GenerateField(field *models.SchemaField) any {
	return ""
}

func createTypeScriptType(t *models.Type) string {
	v := ""

	switch t.Kind {
	case models.Integer, models.Long, models.Float, models.Double:
		v = "number"
	case models.Bool:
		v = "boolean"
	case models.String:
		v = "string"
	case models.List:
		v = createTypeScriptType(&t.Generics[0]) + "[]"
	case models.Map:
		v = "Map<" + createTypeScriptType(&t.Generics[0]) + ", " + createTypeScriptType(&t.Generics[1]) + ">"
	case models.Custom:
		v = *t.CustomTypeName
	}

	if t.Nullable {
		v = v + " | null"
	}

	return v
}

func (g TypeScript) GenerateType(kind *models.Type) any {
	return createTypeScriptType(kind)
}

func (g TypeScript) GenerateTag(tags []models.Tag) any {
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

func (g TypeScript) Generate(model *models.SchemaModel) (*string, error) {
	lines := make([]string, 0)
	lines = append(lines, fmt.Sprintf("export default class %s {", model.ClassName))

	// names := make([]string, 0)
	// maxNameLength := 0

	// types := make([]string, 0)
	// maxTypeLength := 0

	// tags := make([]string, 0)
	g.allTags = make([]string, 0)

	// Collect all the tags
	for _, field := range model.Fields {
		for _, tag := range field.Tags {
			g.allTags = append(g.allTags, tag.TagFor)
		}
	}
	slices.Sort(g.allTags)
	g.allTags = slices.Compact(g.allTags)

	fields := make([]string, 0)

	for _, field := range model.Fields {
		item := field.FieldName + ": " + g.GenerateType(&field.Kind).(string)
		lines = append(lines, "    "+item+";")
		fields = append(fields, item)
	}

	lines = append(lines, "")

	lines = append(lines, "    constructor("+strings.Join(fields, ", ")+") {")
	for _, field := range model.Fields {
		lines = append(lines, "        this."+field.FieldName+" = "+field.FieldName+";")
	}
	lines = append(lines, "    }")

	lines = append(lines, "")

	fieldHasTag := make([][]bool, len(g.allTags))
	for j, tag := range g.allTags {
		fieldHasTag[j] = make([]bool, len(model.Fields))

		for i, field := range model.Fields {
			if slices.ContainsFunc(field.Tags, func(e models.Tag) bool { return e.TagFor == tag }) {
				fieldHasTag[j][i] = true
			}
		}
	}

	for j, tag := range g.allTags {
		names := make([]string, 0)
		kinds := make([]string, 0)
		fields := make([]string, 0)
		for i := 0; i < len(model.Fields); i++ {
			if !fieldHasTag[j][i] {
				continue
			}

			names = append(names, model.Fields[i].FieldName)
			kinds = append(kinds, g.GenerateType(&model.Fields[i].Kind).(string))
			fields = append(fields, model.Fields[i].FieldName+": "+g.GenerateType(&model.Fields[i].Kind).(string))
		}

		// To Tag
		tagFunctionName := strings.ToUpper(tag[0:1]) + tag[1:]
		lines = append(lines, "    public To"+tagFunctionName+"(): Object {")
		lines = append(lines, fmt.Sprintf("        const out: any = {};"))
		lines = append(lines, "    ")

		for i := range len(names) {
			lineString := "        "
			fieldIndex := slices.IndexFunc(model.Fields, func(e models.SchemaField) bool {
				return e.FieldName == names[i]
			})
			tagIndex := slices.IndexFunc(model.Fields[fieldIndex].Tags, func(e models.Tag) bool {
				return e.TagFor == tag
			})

			tagFieldName := model.Fields[fieldIndex].Tags[tagIndex].Values[0]
			if tagFieldName == "-" {
				continue
			}
			if model.Fields[fieldIndex].Tags[tagIndex].Nullable {
				lineString += fmt.Sprintf("if (this.%s) ", names[i])
			}

			lineString += fmt.Sprintf("out[\"%s\"] = this.%s;", tagFieldName, names[i])
			lines = append(lines, lineString)
		}

		lines = append(lines, fmt.Sprintf("\n        return out;"))
		lines = append(lines, "    }\n")

		// From Tag
		lines = append(lines, fmt.Sprintf("    public static From%s(data: any): %s {", tagFunctionName, model.ClassName))
		lines = append(lines, fmt.Sprintf("        return new %s(", model.ClassName))
		for k, field := range model.Fields {
			nullable := field.Kind.Nullable
			nullableString := "!"
			if nullable {
				nullableString = " ?? null"
			}

			if fieldHasTag[j][k] {
				lines = append(lines, fmt.Sprintf("            data[\"%s\"]%s,", field.FieldName, nullableString))
			} else {
				if nullable {
					lines = append(lines, fmt.Sprintf("            null,"))
				} else {
					lines = append(lines, fmt.Sprintf("            %s,", g.GetTypeDefaultInitialization(&field.Kind)))
				}
			}
		}
		lines = append(lines, fmt.Sprintf("        );"))
		lines = append(lines, "    }\n")
	}

	lines = append(lines, "}")

	output := strings.Join(lines, "\n")
	return &output, nil
}
