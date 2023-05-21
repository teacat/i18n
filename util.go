package i18n

import (
	"strings"
)

// ParseAcceptLanguage parses the `Accept-Language` header content and converts to a slice.
// So you can pass it into `NewLocale(...lang)`.
//
// Source: https://siongui.github.io/2015/02/22/go-parse-accept-language/
func ParseAcceptLanguage(acceptLang string) []string {
	var lqs []string
	langQStrs := strings.Split(acceptLang, ",")
	for _, langQStr := range langQStrs {
		trimedLangQStr := strings.Trim(langQStr, " ")
		langQ := strings.Split(trimedLangQStr, ";")
		lqs = append(lqs, nameInsenstive(langQ[0]))
	}
	return lqs
}
