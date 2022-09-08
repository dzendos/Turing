package locales

import (
	"encoding/json"
	"log"
	"os"
)

func localesFile() string {
	return "config/locales/locales.json"
}

type Localizer struct {
	dict map[string]map[string]string
}

func (l *Localizer) Get(local string, field string) string {
	// In case if required localization does not exist
	// we use English localization as a default one.
	if local != "en" && local != "ru" {
		local = "en"
	}

	return l.dict[local][field]
}

func NewLocalizer() *Localizer {
	var local Localizer

	jsonDict, errFile := os.ReadFile(localesFile())

	if errFile != nil {
		log.Print(errFile.Error())
	}

	err := json.Unmarshal(jsonDict, &local.dict)

	if err != nil {
		log.Print(err.Error())
	}

	return &local
}
