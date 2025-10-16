package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"maps"
	"net/http"

	"github.com/seanomeara96/gates/config"
	"github.com/seanomeara96/gates/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Render struct {
	cfg      *config.Config
	template *template.Template
}

func DefaultRender(cfg *config.Config) *Render {
	return &Render{cfg: cfg}
}

func (r *Render) tmpl() *template.Template {
	if r.cfg.Mode == config.Production && r.template != nil {
		return r.template
	}
	r.template = template.Must(template.New("").Funcs(template.FuncMap{
		"add":      func(a, b int) int { return a + b },
		"toString": structToString,
		"sizeRange": func(width, tolerance float32) float32 {
			return width - tolerance
		},
		"title": func(str string) string {
			return cases.Title(language.AmericanEnglish).String(str)
		},
		// expect two different number types int and float32
		"mul": func(a any, b any) float32 {
			var f1, f2 float32

			switch v := a.(type) {
			case int:
				f1 = float32(v)
			case float32:
				f1 = v
			case float64:
				f1 = float32(v)
			default:
				return 0.0
			}

			switch v := b.(type) {
			case int:
				f2 = float32(v)
			case float32:
				f2 = v
			case float64:
				f2 = float32(v)
			default:
				return 0.0
			}

			return f1 * f2
		},
	}).ParseGlob("templates/**/*.tmpl"))
	return r.template
}

func (r *Render) Partial(w http.ResponseWriter, templateName string, templateData any) error {
	var buffer bytes.Buffer
	if err := r.tmpl().ExecuteTemplate(&buffer, templateName, templateData); err != nil {
		return fmt.Errorf("problem executing partial template %s: %w", templateName, err)
	}
	if _, err := w.Write(buffer.Bytes()); err != nil {
		return err
	}
	return nil
}

func (r *Render) Page(w http.ResponseWriter, templateName string, templateData map[string]any) error {
	data := map[string]any{
		"MetaDescription": "default meta description",
		"PageTitle":       "default page title",
		"NavItems": []models.NavItem{
			{Href: "/", Text: "Home"},
			{Href: "/gates", Text: "Gates"},
			{Href: "/extensions", Text: "Extensions"},
			{Href: "/contact", Text: "Contact"},
			{Href: "/cart", Text: "Cart"},
		},
	}
	maps.Copy(data, templateData)

	var buffer bytes.Buffer
	if err := r.tmpl().ExecuteTemplate(&buffer, templateName, data); err != nil {
		return fmt.Errorf("problem executing template %s: %w", templateName, err)
	}
	if _, err := w.Write(buffer.Bytes()); err != nil {
		return err
	}
	return nil
}

func structToString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "error: marshalling struct to string"
	}

	return string(b)
}
