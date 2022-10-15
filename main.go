package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"text/template"
)

type TemplateSpec struct {
	// В этом словаре основные шаблоны. Ключ соответствует ключу в metadata VM куда нужно будет положить результат
	Templates map[string]string `json:"templates"`
	// Это дополнительные шаблоны. Их можно использовать в основных
	// {{ template "includeName" }} — простая вставка
	// {{ base64_template "includeName" }} — вставить результат в Base64 кодировке
	// {{ gzip_template "includeName" }} — вставить результат пожатый Gzip и закодированный в Base64
	Includes map[string]string `json:"includes"`
}

func main() {
	// Читаем шаблоны.
	file, err := os.ReadFile(`./templates.json`)
	if err != nil {
		fmt.Printf("failed to read templates file: %s", err)
		return
	}
	var spec TemplateSpec
	err = json.Unmarshal(file, &spec)
	if err != nil {
		fmt.Printf("templates loading err: %s", err)
		return
	}

	// Читаем значения для подстановки. Можно считать, что они результат заполнения формы.
	var formData map[string]interface{}
	file, err = os.ReadFile(`./formdata.json`)
	if err != nil {
		fmt.Printf("failed to read form data file: %s", err)
		return
	}
	err = json.Unmarshal(file, &formData)
	if err != nil {
		fmt.Printf("templates loading err: %s", err)
		return
	}

	root := template.New("$root")
	funcs := template.FuncMap{
		"base64_template": func(tempName string) string {
			var buf bytes.Buffer
			err := root.Lookup(tempName).Execute(&buf, formData)
			if err != nil {
				return ""
			}
			b := buf.Bytes()
			println(string(b))
			return base64.StdEncoding.EncodeToString(b)
		},
		"base64": func(val string) string {
			println(val)
			return base64.StdEncoding.EncodeToString([]byte(val))
		},
		"gzip_template": func(tempName string) string {
			var buf bytes.Buffer
			zw := gzip.NewWriter(&buf)
			err := root.Lookup(tempName).Execute(zw, formData)
			if err != nil {
				return ""
			}
			err = zw.Close()
			if err != nil {
				return ""
			}
			return base64.StdEncoding.EncodeToString(buf.Bytes())
		},
	}
	for t, s := range spec.Includes {
		_, err := root.New(t).Funcs(funcs).Parse(s)
		if err != nil {
			log.Fatal(err)
		}
	}

	parsedTpls := make(map[string]*template.Template)
	for t, s := range spec.Templates {
		parsed, err := root.New(t).Funcs(funcs).Parse(s)
		if err != nil {
			log.Fatal(err)
		}
		parsedTpls[t] = parsed
	}

	for tempName := range spec.Templates {
		fmt.Printf("%s\n", tempName)
		f, err := os.OpenFile("./result/"+tempName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return
		}
		if err := root.Lookup(tempName).Execute(f, formData); err != nil {
			log.Fatal(err)
		}
		err = f.Close()
		if err != nil {
			return
		}
	}
}
