package generators

import (
	"log"
	"ultratype/models"
)

type Generator interface {
	Generate(model *models.SchemaModel) (*string, error)
	GenerateType(kind *models.Type) any
	GenerateTag(tags []models.Tag) any
	GenerateField(field *models.SchemaField) any
}

func GetGenerator(language string) Generator {
	switch language {
	case "go":
		return GoLang{}

	case "typescript":
		return TypeScript{}
	}

	log.Panic("Language", language, "not supported.")
	return nil
}
