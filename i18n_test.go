package i18n

import (
	"embed"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

//go:embed test/*.json
var testTranslationFS embed.FS

var testTranslations = map[string]map[string]string{
	"en-us": map[string]string{
		"None | 1 Apple | {{ .Count }} Apples": "None | 1 Apple | {{ .Count }} Apples",
	},

	"zh-tw": map[string]string{
		// Token-based Translations
		"test_message":  "這是一則測試訊息。",
		"test_template": "你好，{{ .Name }}！",
		"test_plural":   "沒有 | 只有 1 個 | 有 {{.Count}} 個",

		// Text-based Translations.
		"Hello, world!":             "你好，世界！",
		"How are you, {{ .Name }}?": "過得如何，{{ .Name }}？",
		"Post <verb>":               "發表貼文",
		"Post <noun>":               "文章",

		"None | 1 Apple | {{ .Count }} Apples":         "沒有蘋果 | 1 顆蘋果 | 有 {{.Count}} 顆蘋果",
		"No Post | 1 Post | {{ .Count }} Posts <noun>": "沒有文章 | 1 篇文章 | 有 {{.Count}} 篇文章",
		"No Post | 1 Post | {{ .Count }} Posts <verb>": "沒有發表 | 1 篇發表 | 有 {{.Count}} 篇發表",

		"Post":                                  "THIS_SHOULD_NOT_BE_USED",
		"No Post | 1 Post | {{ .Count }} Posts": "THIS_SHOULD_NOT_BE_USED | THIS_SHOULD_NOT_BE_USED | THIS_SHOULD_NOT_BE_USED",
	},

	"ja-jp": map[string]string{
		// Token-based Translations
		"test_message":  "これはテストメッセージです。",
		"test_template": "こんにちは、{{ .Name }}！",
		"test_plural":   "なし | 1 つだけ | {{.Count}} 個あります",
	},

	"ko-kr": map[string]string{
		// Token-based Translations
		"test_message":  "이것은 테스트 메시지입니다.",
		"test_template": "안녕하세요, {{ .Name }} 님!",
		"test_plural":   "없음 | 1 개 | {{.Count}} 개가 있음",

		// Text-based Translations.
		"Hello, world!":             "안녕하세요, 세상!",
		"How are you, {{ .Name }}?": "{{ .Name }} 님, 어떻게 지내세요?",
		"Post <verb>":               "메시지 게시",
		"Post <noun>":               "기사",
	},
}

func newTestLocale() *Locale {
	i := New("zh-tw")
	i.LoadMap(testTranslations)
	return i.NewLocale("zh-tw")
}

func TestLoadMap(t *testing.T) {
	assert := assert.New(t)

	i := New("zh-tw")
	i.LoadMap(testTranslations)
	l := i.NewLocale("zh-tw")

	assert.Equal("這是一則測試訊息。", l.String("test_message"))
	assert.Equal("not_exists_message", l.String("not_exists_message"))
}

func TestLoadFiles(t *testing.T) {
	assert := assert.New(t)

	i := New("zh-tw")
	assert.NoError(i.LoadFiles("test/zh-tw.json", "test/zh_TW.json", "test/zh_tw.hello.json"))

	l := i.NewLocale("zh-tw")
	assert.Equal("訊息 A", l.String("message_a"))
	assert.Equal("訊息 B", l.String("message_b"))
	assert.Equal("訊息 C", l.String("message_c"))
}

func TestLoadGlob(t *testing.T) {
	assert := assert.New(t)

	i := New("zh-tw")
	assert.NoError(i.LoadGlob("test/*.json"))

	l := i.NewLocale("zh-tw")
	assert.Equal("訊息 A", l.String("message_a"))
	assert.Equal("訊息 B", l.String("message_b"))
	assert.Equal("訊息 C", l.String("message_c"))
}

func TestLoadFS(t *testing.T) {
	assert := assert.New(t)

	i := New("zh-tw")
	assert.NoError(i.LoadFS(testTranslationFS, "test/*.json"))

	l := i.NewLocale("zh-tw")
	assert.Equal("訊息 A", l.String("message_a"))
	assert.Equal("訊息 B", l.String("message_b"))
	assert.Equal("訊息 C", l.String("message_c"))
}

func TestPluralizor(t *testing.T) {
	assert := assert.New(t)

	i := New("ru", WithPluralizor(map[string]Pluralizor{
		"ru": func(number, choices int) int {
			if number == 0 {
				return 0
			}

			teen := number > 10 && number < 20
			endsWithOne := number%10 == 1

			if choices < 4 {
				if !teen && endsWithOne {
					return 1
				} else {
					return 2
				}
			}
			if !teen && endsWithOne {
				return 1
			}
			if !teen && number%10 >= 2 && number%10 <= 4 {
				return 2
			}
			if choices < 4 {
				return 2
			}
			return 3
		},
	}))

	l := i.NewLocale("ru")
	assert.Equal("0 машин", l.Number("0 машин | {{ .Count }} машина | {{ .Count }} машины | {{ .Count }} машин", 0, map[string]any{
		"Count": 0,
	}))
	assert.Equal("1 машина", l.Number("0 машин | {{ .Count }} машина | {{ .Count }} машины | {{ .Count }} машин", 1, map[string]any{
		"Count": 1,
	}))
	assert.Equal("2 машины", l.Number("0 машин | {{ .Count }} машина | {{ .Count }} машины | {{ .Count }} машин", 2, map[string]any{
		"Count": 2,
	}))
	assert.Equal("12 машин", l.Number("0 машин | {{ .Count }} машина | {{ .Count }} машины | {{ .Count }} машин", 12, map[string]any{
		"Count": 12,
	}))
	assert.Equal("21 машина", l.Number("0 машин | {{ .Count }} машина | {{ .Count }} машины | {{ .Count }} машин", 21, map[string]any{
		"Count": 21,
	}))
}

func TestUnmarshaler(t *testing.T) {
	assert := assert.New(t)

	i := New("zh-tw", WithUnmarshaler(yaml.Unmarshal))
	assert.NoError(i.LoadFiles("test/zh_tW.yml"))

	l := i.NewLocale("zh-tw")
	assert.Equal("訊息 A", l.String("message_a"))
}

func TestLocale(t *testing.T) {
	assert := assert.New(t)
	l := newTestLocale()

	assert.Equal("zh-tw", l.Locale())
}

func TestTokenString(t *testing.T) {
	assert := assert.New(t)
	l := newTestLocale()

	assert.Equal("這是一則測試訊息。", l.String("test_message"))
	assert.Equal("not_exists_message", l.String("not_exists_message"))
}

func TestTokenTmpl(t *testing.T) {
	assert := assert.New(t)
	l := newTestLocale()

	assert.Equal("你好，Yami！", l.String("test_template", map[string]string{
		"Name": "Yami",
	}))
}

func TestTokenPlural(t *testing.T) {
	assert := assert.New(t)
	l := newTestLocale()

	assert.Equal("沒有", l.Number("test_plural", 0))
	assert.Equal("只有 1 個", l.Number("test_plural", 1))
	assert.Equal("有 2 個", l.Number("test_plural", 2, map[string]int{
		"Count": 2,
	}))

	// Lazy template
	assert.Equal("沒有", l.Number("test_plural", 0, map[string]int{
		"Count": 2,
	}))
	assert.Equal("只有 1 個", l.Number("test_plural", 1, map[string]int{
		"Count": 2,
	}))
	assert.Equal("有 2 個", l.Number("test_plural", 2, map[string]int{
		"Count": 2,
	}))
}

func TestTextString(t *testing.T) {
	assert := assert.New(t)
	l := newTestLocale()

	assert.Equal("你好，世界！", l.String("Hello, world!"))
}

func TestTextStringRaw(t *testing.T) {
	assert := assert.New(t)
	l := newTestLocale()

	assert.Equal("I'm fine thank you!", l.String("I'm fine thank you!"))
}

func TestTextTmpl(t *testing.T) {
	assert := assert.New(t)
	l := newTestLocale()

	assert.Equal("過得如何，Yami？", l.String("How are you, {{ .Name }}?", map[string]string{
		"Name": "Yami",
	}))
}

func TestTextTmplRaw(t *testing.T) {
	assert := assert.New(t)
	l := newTestLocale()

	assert.Equal("I'm fine, thanks to Yami!", l.String("I'm fine, thanks to {{ .Name }}!", map[string]string{
		"Name": "Yami",
	}))
}

func TestTextPlural(t *testing.T) {
	assert := assert.New(t)
	l := newTestLocale()

	assert.Equal("沒有蘋果", l.Number("None | 1 Apple | {{ .Count }} Apples", 0))
	assert.Equal("1 顆蘋果", l.Number("None | 1 Apple | {{ .Count }} Apples", 1))
	assert.Equal("有 2 顆蘋果", l.Number("None | 1 Apple | {{ .Count }} Apples", 2, map[string]int{
		"Count": 2,
	}))

	// Lazy template
	assert.Equal("沒有蘋果", l.Number("None | 1 Apple | {{ .Count }} Apples", 0, map[string]int{
		"Count": 2,
	}))
	assert.Equal("1 顆蘋果", l.Number("None | 1 Apple | {{ .Count }} Apples", 1, map[string]int{
		"Count": 2,
	}))
	assert.Equal("有 2 顆蘋果", l.Number("None | 1 Apple | {{ .Count }} Apples", 2, map[string]int{
		"Count": 2,
	}))
}

func TestTextPluralRaw(t *testing.T) {
	assert := assert.New(t)
	l := newTestLocale()

	assert.Equal("Zero", l.Number("Zero | 1 Thing | {{ .Count }} Things", 0))
	assert.Equal("1 Thing", l.Number("Zero | 1 Thing | {{ .Count }} Things", 1))
	assert.Equal("2 Things", l.Number("Zero | 1 Thing | {{ .Count }} Things", 2, map[string]int{
		"Count": 2,
	}))
}

func TestTextStringContext(t *testing.T) {
	assert := assert.New(t)
	l := newTestLocale()

	assert.Equal("發表貼文", l.StringX("Post", "verb"))
	assert.Equal("文章", l.StringX("Post", "noun"))
}

func TestTextPluralContext(t *testing.T) {
	assert := assert.New(t)
	l := newTestLocale()

	assert.Equal("沒有文章", l.NumberX("No Post | 1 Post | {{ .Count }} Posts", "noun", 0))
	assert.Equal("1 篇文章", l.NumberX("No Post | 1 Post | {{ .Count }} Posts", "noun", 1))
	assert.Equal("有 2 篇文章", l.NumberX("No Post | 1 Post | {{ .Count }} Posts", "noun", 2, map[string]int{
		"Count": 2,
	}))

	// Lazy template
	assert.Equal("沒有文章", l.NumberX("No Post | 1 Post | {{ .Count }} Posts", "noun", 0, map[string]int{
		"Count": 2,
	}))
	assert.Equal("1 篇文章", l.NumberX("No Post | 1 Post | {{ .Count }} Posts", "noun", 1, map[string]int{
		"Count": 2,
	}))
	assert.Equal("有 2 篇文章", l.NumberX("No Post | 1 Post | {{ .Count }} Posts", "noun", 2, map[string]int{
		"Count": 2,
	}))

	//
	assert.Equal("沒有發表", l.NumberX("No Post | 1 Post | {{ .Count }} Posts", "verb", 0))
	assert.Equal("1 篇發表", l.NumberX("No Post | 1 Post | {{ .Count }} Posts", "verb", 1))
	assert.Equal("有 2 篇發表", l.NumberX("No Post | 1 Post | {{ .Count }} Posts", "verb", 2, map[string]int{
		"Count": 2,
	}))
}

func TestTextFallback(t *testing.T) {
	assert := assert.New(t)
	i := New("zh-tw", WithFallback(map[string][]string{
		"ja-jp": []string{"ko-kr"},
	}))
	i.LoadMap(testTranslations)
	l := i.NewLocale("ja-jp")

	// Test ja-jp
	assert.Equal("これはテストメッセージです。", l.String("test_message"))
	assert.Equal("こんにちは、Yami！", l.String("test_template", map[string]string{
		"Name": "Yami",
	}))
	assert.Equal("なし", l.Number("test_plural", 0))

	// Test ja-jp -> ko-kr fallback
	assert.Equal("안녕하세요, 세상!", l.String("Hello, world!"))
	assert.Equal("Yami 님, 어떻게 지내세요?", l.String("How are you, {{ .Name }}?", map[string]string{
		"Name": "Yami",
	}))
	assert.Equal("메시지 게시", l.StringX("Post", "verb"))

	// Test ja-jp -> zh-tw fallback
	assert.Equal("沒有蘋果", l.Number("None | 1 Apple | {{ .Count }} Apples", 0))
	assert.Equal("1 顆蘋果", l.Number("None | 1 Apple | {{ .Count }} Apples", 1))
	assert.Equal("有 2 顆蘋果", l.Number("None | 1 Apple | {{ .Count }} Apples", 2, map[string]int{
		"Count": 2,
	}))

	// Test nil fallback
	assert.Equal("Ni hao", l.String("Ni hao"))
}

func TestTextFallbackResursive(t *testing.T) {
	assert := assert.New(t)
	i := New("en-us", WithFallback(map[string][]string{
		"ja-jp": []string{"ko-kr"},
		"ko-kr": []string{"zh-tw"},
	}))
	i.LoadMap(testTranslations)
	l := i.NewLocale("ja-jp")

	// Test ja-jp -> ko-kr -> zh-tw fallback
	assert.Equal("1 顆蘋果", l.Number("None | 1 Apple | {{ .Count }} Apples", 1))
}

func TestParseAcceptLanguage(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("[zh-tw zh en-us en ja]", fmt.Sprintf("%+v", ParseAcceptLanguage("zh-TW,zh;q=0.9,en-US;q=0.8,en;q=0.7,ja;q=0.6")))
}
