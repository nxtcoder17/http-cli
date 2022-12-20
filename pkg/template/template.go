package template

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

func txtFuncs(t *template.Template) template.FuncMap {
	funcs := sprig.TxtFuncMap()

	// inspired by helm include
	funcs["include"] = func(templateName string, templateData any) (string, error) {
		buf := bytes.NewBuffer(nil)
		if err := t.ExecuteTemplate(buf, templateName, templateData); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	funcs["toYAML"] = func(txt any) (string, error) {
		if txt == nil {
			return "", nil
		}

		a, ok := funcs["toPrettyJson"].(func(any) string)
		if !ok {
			panic("could not convert sprig.TxtFuncMap[toPrettyJson] into func(any) string")
		}

		x := a(txt)
		if x == "null" {
			return "", nil
		}

		ys, err := yaml.JSONToYAML([]byte(x))
		if err != nil {
			return "", err
		}
		return string(ys), nil
	}

	funcs["md5"] = func(txt string) string {
		hash := md5.New()
		hash.Write([]byte(txt))
		return hex.EncodeToString(hash.Sum(nil))
	}

	funcs["K8sAnnotation"] = func(cond any, key string, value any) string {
		if cond == reflect.Zero(reflect.TypeOf(cond)).Interface() {
			return ""
		}
		return fmt.Sprintf("%s: '%v'", key, value)
	}

	funcs["K8sLabel"] = funcs["K8sAnnotation"]

	funcs["Iterate"] = func(count int) []int {
		var i int
		var Items []int
		for i = 0; i < count; i++ {
			Items = append(Items, i)
		}
		return Items
	}

	return funcs
}

func New() *template.Template {
	t := template.New("inline")
	t.Funcs(txtFuncs(t))
	return t
}

func ParseBytes(b []byte, values any) ([]byte, error) {
	t := template.New("parse-bytes")
	t.Funcs(txtFuncs(t))
	if _, err := t.Parse(string(b)); err != nil {
		return nil, err
	}

	out := new(bytes.Buffer)
	if err := t.ExecuteTemplate(out, "parse-bytes", values); err != nil {
		return nil, errors.Wrap(err, "could not execute template")
	}
	return out.Bytes(), nil
}
