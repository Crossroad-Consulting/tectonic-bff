package models

import (
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/example/placeholder-webapp/translations"
)

var translationFields = [...]string{
	"title",
	"logout",
	"download",
	"general",
	"safety",
	"schedule",
	"nofilesfound",
}

type LanguageTranslation map[string]string

var Translations = map[string]LanguageTranslation{}

func LoadTranslation(lang, path string) error {
	data, err := fs.ReadFile(translations.FS, path)
	if err != nil {
		return fmt.Errorf("failed to open translation file: %v", err)
	}

	var translation LanguageTranslation
	err = json.Unmarshal(data, &translation)
	if err != nil {
		return fmt.Errorf("failed to unmarshal translation file: %v", err)
	}

	for _, f := range translationFields {
		v, ok := translation[f]
		if !ok {
			return fmt.Errorf("key '%s' does not exist in translation file", f)
		}
		if v == "" {
			return fmt.Errorf("key '%s' is empty in translation file", f)
		}
	}

	Translations[lang] = translation

	return nil
}
