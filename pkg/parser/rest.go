package parser

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/nxtcoder17/http-cli/pkg/template"
	"sigs.k8s.io/yaml"
)

type RestBlock struct {
	Label string `json:"label,omitempty"`
	Query struct {
		Method  string            `json:"method,omitempty"`
		Url     string            `json:"url,omitempty"`
		Headers map[string]string `json:"headers,omitempty"`
	} `json:"query"`
	Body map[string]any `json:"body,omitempty"`
}

func parseBody(body map[string]any, gbl GlobalVars) (io.Reader, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	pb, err := template.ParseBytes(b, gbl)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(pb), nil
}

func parseUrl(url string, gbl GlobalVars) (string, error) {
	pb, err := template.ParseBytes([]byte(url), gbl)
	if err != nil {
		return "", err
	}

	return string(pb), nil
}

func ParseRestQuery(yql *YamlQueryBlock, env *EnvFile) (*http.Request, error) {
	var restBlock RestBlock
	if err := yaml.Unmarshal(yql.YAMLQuery, &restBlock); err != nil {
		return nil, err
	}

	vars := make(map[string]any, len(env.Map[env.Mode].Vars)+len(yql.Global))

	for k, v := range env.Map[env.Mode].Vars {
		vars[k] = v
	}

	for k, v := range yql.Global {
		vars[k] = v
	}

	url, err := parseUrl(restBlock.Query.Url, vars)
	if err != nil {
		return nil, err
	}

	body, err := parseBody(restBlock.Body, vars)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(restBlock.Query.Method, url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range env.Map[env.Mode].Headers {
		req.Header.Set(k, v)
	}
	for k, v := range restBlock.Query.Headers {
		req.Header.Set(k, v)
	}

	return req, nil
}
