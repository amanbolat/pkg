package i18n

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Localizer is a wrapper around i18n.Localizer with a better API.
type Localizer struct {
	localizers       map[language.Tag]*i18n.Localizer
	defaultLang      language.Tag
	defaultLocalizer *i18n.Localizer
}

// NewLocalizer creates a new Localizer from map with a key as filename and
// value as its content in toml format. Filename should satisfy the pattern
// <domain/description>.<language>.toml.
func NewLocalizer(defaultLang language.Tag, messagesMap map[string][]byte) (*Localizer, error) {
	l := Localizer{
		defaultLang: defaultLang,
		localizers:  map[language.Tag]*i18n.Localizer{},
	}

	bundle := i18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	langs := make([]language.Tag, len(messagesMap))
	var i int
	for fileName, data := range messagesMap {
		arr := strings.Split(fileName, ".")
		if len(arr) != 3 {
			return nil, fmt.Errorf("translation file [%s] does not satisfy pattern: <domain>.<language>.toml", fileName)
		}
		lang := arr[1]
		langTag, err := language.Parse(lang)
		if err != nil {
			return nil, fmt.Errorf("language [%s] for file [%s] is not supported: %w", lang, fileName, err)
		}
		langs[i] = langTag
		i++

		bundle.MustParseMessageFileBytes(data, fileName)

		l.localizers[langTag] = i18n.NewLocalizer(bundle, langTag.String())
	}

	var hasDefaultLang bool
	for _, l := range langs {
		if l == defaultLang {
			hasDefaultLang = true
		}
	}
	if !hasDefaultLang {
		return nil, fmt.Errorf("bundle has no messages for default lang [%s]", defaultLang)
	}

	l.defaultLocalizer = l.localizers[defaultLang]

	return &l, nil
}

// Localize returns localized string for default language.
func (l *Localizer) Localize(msgID string, params ...map[string]interface{}) string {
	return localize(l.defaultLocalizer, msgID, params...)
}

// LocalizeFor returns localized string for specified language.
func (l *Localizer) LocalizeFor(lang language.Tag, msgID string, params ...map[string]interface{}) string {
	localizer, ok := l.localizers[lang]
	if !ok {
		return ""
	}

	return localize(localizer, msgID, params...)
}

func localize(l *i18n.Localizer, msgID string, params ...map[string]interface{}) string {
	var tmplData interface{}
	if len(params) > 0 {
		tmplData = params[0]
	}

	str, _ := l.Localize(&i18n.LocalizeConfig{
		MessageID:    msgID,
		TemplateData: tmplData,
	})
	if str == "" {
		return msgID
	}
	return str
}
