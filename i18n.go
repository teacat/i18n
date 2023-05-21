package i18n

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// Pluralizor decides which translation string to use by the returned index.
type Pluralizor func(number, choices int) int

// Unmarshaler unmarshals the translation files, can be `json.Unmarshal` or `yaml.Unmarshal`.
type Unmarshaler func(data []byte, v any) error

// I18n is the main internationalization core.
type I18n struct {
	defaultLocale               string
	pluralizors                 map[string]Pluralizor
	unmarshaler                 Unmarshaler
	fallbacks                   map[string][]string
	translations                map[string]map[string]string
	runtimeCompiledTranslations map[string]*compiledTranslation
	compiledTranslations        map[string]map[string]*compiledTranslation
}

// WithUnmarshaler replaces the default translation file unmarshaler.
func WithUnmarshaler(u Unmarshaler) func(*I18n) {
	return func(i *I18n) {
		i.unmarshaler = u
	}
}

// WithFallback changes fallback settings.
func WithFallback(f map[string][]string) func(*I18n) {
	return func(i *I18n) {
		i.fallbacks = f
	}
}

// WithPluralizor changes pluralizors.
func WithPluralizor(p map[string]Pluralizor) func(*I18n) {
	return func(i *I18n) {
		i.pluralizors = p
	}
}

// New creates a new internationalization.
func New(defaultLocale string, options ...func(*I18n)) *I18n {
	i := &I18n{
		defaultLocale:               defaultLocale,
		unmarshaler:                 json.Unmarshal,
		pluralizors:                 make(map[string]Pluralizor),
		fallbacks:                   make(map[string][]string),
		translations:                make(map[string]map[string]string),
		runtimeCompiledTranslations: make(map[string]*compiledTranslation),
		compiledTranslations:        make(map[string]map[string]*compiledTranslation),
	}
	for _, o := range options {
		o(i)
	}
	return i
}

// LoadMap loads the translations from the map.
func (i *I18n) LoadMap(languages map[string]map[string]string) error {
	for locale, translations := range languages {
		locale = nameInsenstive(locale)
		i.compiledTranslations[locale] = make(map[string]*compiledTranslation)

		for name, text := range translations {
			trans := i.compileTranslation(locale, name, text)
			i.compiledTranslations[locale][name] = trans
		}
	}
	i.compileFallbacks()
	return nil
}

// LoadFiles loads the translations from the files.
func (i *I18n) LoadFiles(filenames ...string) error {
	data := make(map[string]map[string]string)

	for _, v := range filenames {
		b, err := os.ReadFile(v)
		if err != nil {
			return err
		}
		var trans map[string]string
		if err := i.unmarshaler(b, &trans); err != nil {
			return err
		}
		locale := nameInsenstive(v)
		_, ok := data[locale]
		if !ok {
			data[locale] = make(map[string]string)
		}
		for name, text := range trans {
			data[locale][name] = text
		}
	}
	return i.LoadMap(data)
}

// LoadGlob loads the translations from the files that matches specified patterns.
func (i *I18n) LoadGlob(pattern ...string) error {
	var files []string

	for _, pattern := range pattern {
		v, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		files = append(files, v...)
	}

	return i.LoadFiles(files...)
}

// LoadFS loads the translation from a `fs.FS`, useful for `go:embed`.
func (i *I18n) LoadFS(fsys fs.FS, patterns ...string) error {
	var files []string
	data := make(map[string]map[string]string)

	for _, pattern := range patterns {
		v, err := fs.Glob(fsys, pattern)
		if err != nil {
			return err
		}
		files = append(files, v...)
	}

	for _, v := range files {
		b, err := fs.ReadFile(fsys, v)
		if err != nil {
			return err
		}
		var trans map[string]string
		if err := i.unmarshaler(b, &trans); err != nil {
			return err
		}

		locale := nameInsenstive(v)

		_, ok := data[locale]
		if !ok {
			data[locale] = make(map[string]string)
		}
		for name, text := range trans {
			data[locale][name] = text
		}
	}
	return i.LoadMap(data)
}

// NewLocale reads a locale from the internationalization core.
func (i *I18n) NewLocale(locales ...string) *Locale {
	selectedLocale := i.defaultLocale
	for _, v := range locales {
		v = nameInsenstive(v)
		if _, ok := i.compiledTranslations[v]; ok {
			selectedLocale = v
			break
		}
	}
	return &Locale{
		parent: i,
		locale: selectedLocale,
	}
}

var contextRegExp = regexp.MustCompile("<(.*?)>$")

// compiledTranslation
type compiledTranslation struct {
	locale     string
	name       string
	pluralizor Pluralizor
	texts      []*compiledText
}

// compiledText
type compiledText struct {
	text string
	tmpl *template.Template
}

// defaultPluralizor
func defaultPluralizor(number, choices int) int {
	switch choices {
	case 2:
		switch number {
		case 0, 1:
			return 0
		default:
			return 1
		}
	default:
		switch number {
		case 0:
			return 0
		case 1:
			return 1
		default:
			return 2
		}
	}
}

// pluralizor
func (i *I18n) pluralizor(lang string) Pluralizor {
	v, ok := i.pluralizors[lang]
	if !ok {
		return defaultPluralizor
	}
	return v
}

// trimContext
func trimContext(v string) string {
	return contextRegExp.ReplaceAllString(v, "")
}

// compileTranslation
func (i *I18n) compileTranslation(locale, name, text string) *compiledTranslation {
	compTrans := &compiledTranslation{
		name: name,
	}
	compTrans.locale = locale
	compTrans.pluralizor = i.pluralizor(locale)
	compTrans.texts = compileText(text)

	return compTrans
}

// compileText
func compileText(text string) (compTexts []*compiledText) {
	texts := strings.Split(text, " | ")

	for _, v := range texts {
		compText := &compiledText{}

		if strings.Contains(v, "{{") {
			t, _ := template.New("").Parse(v)
			compText.tmpl = t
		} else {
			compText.text = v
		}
		compTexts = append(compTexts, compText)
	}
	return
}

// nameInsenstive converts `zh_TW.music.json`, `zh_TW` and `zh-TW` to `zh-tw`.
func nameInsenstive(v string) string {
	v = filepath.Base(v)
	v = strings.Split(v, ".")[0]
	v = strings.ToLower(v)
	v = strings.ReplaceAll(v, "_", "-")
	return v
}

// compileFallbacks
func (i *I18n) compileFallbacks() {
	for _, grandTrans := range i.compiledTranslations[i.defaultLocale] {
		for locale, trans := range i.compiledTranslations {
			//
			if locale == i.defaultLocale {
				continue
			}
			//
			if _, ok := trans[grandTrans.name]; !ok {
				if bestfit := i.lookupBestFallback(locale, grandTrans.name); bestfit != nil {
					i.compiledTranslations[locale][grandTrans.name] = bestfit
				}
			}
		}
	}
}

// lookupBestFallback
func (i *I18n) lookupBestFallback(locale, name string) *compiledTranslation {
	fallbacks, ok := i.fallbacks[locale]
	if !ok {
		if v, ok := i.compiledTranslations[i.defaultLocale][name]; ok {
			return v
		}
	}
	for _, fallback := range fallbacks {
		if v, ok := i.compiledTranslations[fallback][name]; ok {
			return v
		}
		if j := i.lookupBestFallback(fallback, name); j != nil {
			return j
		}
	}
	return nil
}
