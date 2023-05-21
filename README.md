# i18n [![GoDoc](https://godoc.org/github.com/teacat/i18n?status.svg)](https://godoc.org/github.com/teacat/i18n) [![Coverage Status](https://coveralls.io/repos/github/teacat/i18n/badge.svg?branch=master)](https://coveralls.io/github/teacat/i18n?branch=master) [![Build Status](https://app.travis-ci.com/teacat/i18n.svg?branch=master)](https://app.travis-ci.com/github/teacat/i18n) [![Go Report Card](https://goreportcard.com/badge/github.com/teacat/i18n)](https://goreportcard.com/report/github.com/teacat/i18n)

`teacat/i18n` is a simple, easy i18n package for Golang that helps you translate Go programs into multiple languages.

-   Token-based (`hello_world`) and Text-based (`Hello, world!`) translation.
-   Variables in translation powered by [`text/template`](https://pkg.go.dev/text/template) with Pre-Compiled Techonology™ 😎👍
-   Pluralization and Custom Pluralizor.
-   Load translations from a map, files or even [`fs.FS`](https://pkg.go.dev/io/fs) (`go:embed` supported).
-   Supports any translation file format (e.g. JSON, YAML).

&nbsp;

## Installation

```bash
$ go get github.com/teacat/i18n
```

&nbsp;

## Example

```go
package main

import (
    "github.com/teacat/i18n"
    "fmt"
)

func main() {
    i := i18n.New("zh-tw")
    i.LoadMap(map[string]map[string]string{
        "en-us": map[string]string{
            "hello_world": "Hello, world!"
        }
    })

    l := i.NewLocale("en-us")

    // Output: Hello, world!
    fmt.Println(l.String("hello_world"))

    // Output: What a wonderful world!
    fmt.Println(l.String("What a wonderful world!"))

    // Output: How are you, Yami?
    fmt.Println(l.String("How are you, {{ .Name }}?", map[string]any{
        "Name": "Yami",
    }))

    // Output: 3 Posts
    fmt.Println(l.Number("No Posts | 1 Post | {{ .Count }} Posts", 3, map[string]any{
        "Count": 3,
    }))
}
```

&nbsp;

## Index

-   [Getting Started](#getting-started)
-   [Translations](#translations)
    -   [Passing Data to Translation](#passing-data-to-translation)
-   [Pluralization](#pluralization)
-   [Text-based Translations](#text-based-translations)
    -   [Disambiguation by context](#disambiguation-by-context)
    -   [Act as fallback](#act-as-fallback)
-   [Fallbacks](#fallbacks)
-   [Custom Unmarshaler](#custom-unmarshaler)
-   [Custom Pluralizor](#custom-pluralizor)
-   [Parse Accept-Language](#parse-accept-language)
-   [Load from FS](#load-from-fs)

&nbsp;

## Getting Started

Initialize with a default language, then load the translations from a map or the files.

```go
package main

import "github.com/teacat/i18n"

func main() {
    i := i18n.New("zh-tw")

    // (a) Load the translation from a map.
    i.LoadMap(map[string]map[string]string{
        "zh-tw": map[string]string{
            "hello_world": "早安，世界",
        },
    })

    // (b) Load from "zh-tw.json", "en-us.json", "ja-jp.json".
    i.LoadFiles("zh-tw.json", "en-us.json", "ja-jp.json")

    // (c) Load all json files under `language` folder.
    i.LoadGlob("languages/*.json")
}
```

Filenames like `zh_TW.json`, `zh-tw.json` `zh_tw.user.json`, `zh-TW.music.json` will be combined to a single `zh-tw` translation (case-insenstive and the suffixes are ignored).

&nbsp;

## Translations

Translations named like `welcome_message`, `button_create`, `button_buy` are token-based translations. For text-based, check the chapters below.

```json
{
    "message_basic": "你好，世界"
}
```

```go
locale := i.NewLocale("zh-tw")

// Output: 你好，世界
locale.String("message_basic")

// Output: message_what_is_this
locale.String("message_what_is_this")
```

&nbsp;

### Passing Data to Translation

It's possible to pass the data to translations. `text/template` is used to parse the text, the templates will be parsed and cached after the translation was loaded.

```json
{
    "message_tmpl": "你好，{{ .Name }}"
}
```

```go
// Output: 你好，Yami
locale.String("message_tmpl", map[string]any{
    "Name": "Yami",
})
```

&nbsp;

## Pluralization

Simpliy dividing the translation text into `zero,one | many` (2 options) and `zero | one | many` (3 options) format to use pluralization.

※ Spaces around the `|` separators are **REQUIRED**.

```json
{
    "apples": "我沒有蘋果 | 我只有 1 個蘋果 | 我有 {{ .Count }} 個蘋果"
}
```

```go
// Output: 我沒有蘋果
locale.Number("apples", 0)

// Output: 我只有 1 個蘋果
locale.Number("apples", 1)

// Output: 我有 3 個蘋果
locale.Number("apples", 3, map[string]any{
    "Count": 3,
})
```

&nbsp;

## Text-based Translations

Translations can also be named with sentences so it will act like fallbacks when the translation was not found.

```json
{
    "I'm fine.": "我過得很好。",
    "How about you?": "你如何呢？"
}
```

```go
// Output: 我過得很好。
locale.String("I'm fine.")

// Output: 你如何呢？
locale.String("How about you?")

// Output: Thank you!
locale.String("Thank you!")
```

&nbsp;

### Disambiguation by context

In English a "Post" can be "Post something (verb)" or "A post (noun)". With token-based translation, you can easily separating them to `post_verb` and `post_noun`.

With text-based translation, you will need to use `StringX` (X stands for context), and giving the translation a `<context>` suffix.

The space before the `<` is **REQUIRED**.

```json
{
    "Post <verb>": "發表文章",
    "Post <noun>": "一篇文章"
}
```

```go
// Output: 發表文章
locale.StringX("Post", "verb")

// Output: 一篇文章
locale.StringX("Post", "noun")

// Output: Post
locale.StringX("Post", "adjective")
```

&nbsp;

### Act as fallback

Remember, if a translation was not found, the token name will be output directly. The token name can also be used as template content.

```go
// Output: Hello, World
locale.String("Hello, {{ .Name }}", map[string]any{
    "Name": "World",
})

// Output: 2 Posts
locale.Number("None | 1 Post | {{ .Count }} Posts", 2, map[string]any{
    "Count": 2,
})
```

&nbsp;

## Fallbacks

A fallback language will be used when a translation is missing from the current language. If it's still missing from the fallback language, it will lookup from the default language.

If a translation cannot be found from any language, the token name will be output directly.

```go
// `ja-jp` is the default language
i := i18n.New("ja-jp", WithFallback(map[string][]string{
    // `zh-tw` uses `zh-hk`, `zh-cn` as fallbacks.
    // `en-gb` uses `en-us` as fallback.
    "zh-tw": []string{"zh-hk", "zh-cn"},
    "en-gb": []string{"en-us"},
}))
```

Lookup path looks like this with the example above:

```
zh-tw -> zh-hk -> zh-cn -> ja-jp
en-gb -> en-us -> ja-jp
```

Recursive fallback is also supported. If `zh-tw` has a `zh-hk` fallback, and `zh-hk` has a `zh-cn` fallback, `zh-tw` will have either `zh-hk` and `zh-cn` fallbacks.

Fallback only works if the translation exists in default language.

&nbsp;

## Custom Unmarshaler

Translations are JSON format because `encoding/json` is the default unmarshaler. Change it by calling `WithUnmarshaler`.

The following example uses [`go-yaml/yaml`](https://github.com/go-yaml/yaml) to read the files, so you can write the translation files in YAML format.

```go
package main

import "gopkg.in/yaml.v3"

func main() {
    i := i18n.New("zh-tw", WithUnmarshaler(yaml.Unmarshal))
    i.LoadFiles("zh-tw.yaml")
}
```

Your `zh-tw.yaml` should look like this:

```yaml
hello_world: "你好，世界"
"How are you?": "你過得如何？"
"mobile_interface.button": "按鈕"
```

Nested translations are not supported, you will need to name them like `"mobile_interface.button"` as key and quote them in double quotes.

&nbsp;

## Custom Pluralizor

Languages like Slavic languages (Russian, Ukrainian, etc.) has complex pluralization rules. To change the default `zero | one | many` behaviour, use `WithPluralizor`.

An example translation text like `a | b | c | d`, the `choices` will be `4`, if `0` was returned, then `a` will be used.

```go
i := i18n.New("zh-tw", WithPluralizor(map[string]Pluralizor{
    // A simplified pluralizor for Slavic languages (Russian, Ukrainian, etc.).
    "ru": func(number, choices int) int {
        if number == 0 {
            return 0
        }

        teen := number > 10 && number < 20
        endsWithOne := number % 10 == 1

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
        if !teen && number % 10 >= 2 && number % 10 <= 4 {
            return 2
        }
        if choices < 4 {
            return 2
        }
        return 3
    },
})
```

The `ru.json` file:

```json
{
    "car": "0 машин | {{ .Count }} машина | {{ .Count }} машины | {{ .Count }} машин"
}
```

```go
locale := i.NewLocale("ru")

// Output: 0 машин
i.Number("car", 0, map[string]any{
    "Count": 0,
})
// Output: 1 машина
i.Number("car", 1, map[string]any{
    "Count": 1,
})
// Output: 2 машины
i.Number("car", 2, map[string]any{
    "Count": 2,
})
// Output: 12 машин
i.Number("car", 12, map[string]any{
    "Count": 12,
})
// Output: 21 машина
i.Number("car", 21, map[string]any{
    "Count": 21,
})
```

&nbsp;

## Parse Accept-Language

The built-in `ParseAcceptLanguage` function helps you to parse the `Accept-Language` from HTTP Header.

```go
func(w http.ResponseWriter, r *http.Request) {
    // Initialize i18n.
    i := i18n.New("zh-tw")
    i.LoadFiles("zh-tw.json", "en-us.json")

    // Get `Accept-Language` from request header.
    accept := r.Header.Get("Accept-Language")

    // Use the locale.
    l := i.NewLocale(...i18n.ParseAcceptLanguage(accept))
    l.String("hello_world")
}
```

Orders of the languages that passed to `NewLocale` won't affect the fallback priorities, it will use the first language that was found in loaded translations.

&nbsp;

## Load from FS

Use `LoadFS` if you are using `go:embed` to compile your translations to the program.

```go
package main

import "github.com/teacat/i18n"

//go:embed languages/*.json
var langFS embed.FS

func main() {
    i := i18n.New("zh-tw")

    // Load all json files under `language` folder from the filesystem.
    i.LoadFS(langFS, "languages/*.json")
}
```
