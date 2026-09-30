package views

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/example/placeholder-webapp/models"
	"github.com/gorilla/csrf"
)

type public interface {
	Public() string
}

type Template struct {
	htmlTpl *template.Template
}

func Must(t Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return t
}

func ParseFS(fs fs.FS, patterns ...string) (Template, error) {
	tpl := template.New(patterns[0])
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfField": func() (template.HTML, error) {
				return "", fmt.Errorf("csrfField not implemented")
			},
			"translate": func(key string) (string, error) {
				return "", fmt.Errorf("translate not implemented")
			},
			"errors": func() []string {
				return nil
			},
		},
	)
	tpl, err := tpl.ParseFS(fs, patterns...)
	if err != nil {
		return Template{}, fmt.Errorf("parsing template: %w", err)
	}
	return Template{
		htmlTpl: tpl,
	}, nil
}

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data any, errs ...error) {
	tpl, err := t.htmlTpl.Clone()
	if err != nil {
		log.Printf("cloning template: %v", err)
		http.Error(w, "There was an error rendering the page.", http.StatusInternalServerError)
		return
	}
	errMsgs := errMessages(errs...)
	language := r.URL.Query().Get("lang")
	if language == "" {
		language = getBestLanguage(r.Header.Get("Accept-Language"))
	}
	translation, ok := models.Translations[language]
	if !ok {
		log.Printf("translation for language %s not found\n", language)
		translation = models.Translations["en"]
	}
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfField": func() template.HTML {
				return csrf.TemplateField(r)
			},
			"translate": func(key string) string {
				value, ok := translation[key]
				if !ok {
					log.Printf("translation for key %s not found\n", key)
					return key
				}
				return value
			},
			"errors": func() []string {
				return errMsgs
			},
		},
	)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		log.Printf("executing template: %v", err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buf)
}

func errMessages(errs ...error) []string {
	var msgs []string
	for _, err := range errs {
		var pubErr public
		log.Println(err)
		if errors.As(err, &pubErr) {
			msgs = append(msgs, pubErr.Public())
		} else {
			msgs = append(msgs, "Something went wrong.")
		}
	}
	return msgs
}

func getBestLanguage(acceptLanguageHeader string) string {
	languages := strings.Split(acceptLanguageHeader, ",")
	var bestLanguage string
	maxQValue := -1.0

	for _, lang := range languages {
		languageParts := strings.Split(lang, ";")
		languageCode := strings.TrimSpace(languageParts[0])
		qValue := 1.0
		if len(languageParts) > 1 {
			qValue = getQValue(languageParts[1])
		}

		l, ok := checkLanguage(languageCode)
		if ok && qValue > maxQValue {
			bestLanguage = l
			maxQValue = qValue
		}
	}

	if bestLanguage == "" {
		return "en"
	}

	return bestLanguage
}

func getQValue(qValueStr string) float64 {
	parts := strings.Split(qValueStr, "=")
	if len(parts) < 2 {
		return 1.0
	}

	qValue, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	return qValue
}

func checkLanguage(languageCode string) (string, bool) {
	regex, err := regexp.Compile("(nl|en|fr)(-*)?")
	if err != nil {
		panic(err)
	}

	return languageCode[:2], regex.MatchString(languageCode)
}
