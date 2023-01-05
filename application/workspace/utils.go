package workspace

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig"
)

func parseEnvTemplate(envTemplates map[string]string) (map[string]*template.Template, error) {
	var envs = make(map[string]*template.Template, len(envTemplates))
	for k, v := range envTemplates {
		t, err := parseTemplate(v)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", k, err)
		}
		envs[k] = t
	}
	return envs, nil
}

func parseTemplate(text string) (*template.Template, error) {
	return template.New("").Option("missingkey=zero").Funcs(sprig.TxtFuncMap()).Parse(text)
}

func renderTemplate(t *template.Template, dataContext any) (string, error) {
	var buf bytes.Buffer
	err := t.Execute(&buf, dataContext)
	return buf.String(), err
}
