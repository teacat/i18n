package i18n

import (
	"bytes"
	"fmt"
)

// Locale represents a translated locale.
type Locale struct {
	parent *I18n

	locale string
}

// Locale returns the current locale name.
func (l *Locale) Locale() string {
	return l.locale
}

// String returns a translated string.
func (l *Locale) String(name string, data ...any) string {
	selectedTrans := l.lookup(name)
	return l.render(selectedTrans.texts[0], data...)
}

// StringX returns a translated string with a specified context.
func (l *Locale) StringX(name, context string, data ...any) string {
	return l.String(fmt.Sprintf("%s <%s>", name, context), data...)
}

// Number returns a translated string based on the `count`.
func (l *Locale) Number(name string, count int, data ...any) string {
	selectedTrans := l.lookup(name)
	selectedIndex := selectedTrans.pluralizor(count, len(selectedTrans.texts))
	return l.render(selectedTrans.texts[selectedIndex], data...)
}

// NumberX returns a translated string based on the `count` with a specified context.
func (l *Locale) NumberX(name string, context string, count int, data ...any) string {
	return l.Number(fmt.Sprintf("%s <%s>", name, context), count, data...)
}

// lookup
func (l *Locale) lookup(name string) *compiledTranslation {
	if selectedTrans, ok := l.parent.compiledTranslations[l.locale][name]; ok {
		return selectedTrans
	}
	runtimeTrans, ok := l.parent.runtimeCompiledTranslations[name]
	if !ok {
		runtimeTrans = l.parent.compileTranslation(l.parent.defaultLocale, name, trimContext(name))
	}
	l.parent.runtimeCompiledTranslations[name] = runtimeTrans
	return runtimeTrans
}

// render
func (l *Locale) render(text *compiledText, data ...any) string {
	if text.tmpl != nil {
		var tpl bytes.Buffer
		if len(data) > 0 {
			text.tmpl.Execute(&tpl, data[0])
		} else {
			text.tmpl.Execute(&tpl, nil)
		}
		return tpl.String()
	}
	return text.text
}
