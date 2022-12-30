package parser

import (
	"bytes"
	"encoding/json"
	"github.com/nxtcoder17/http-cli/pkg/template"
	"io"
	"net/http"
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
	b, err := json.Marshal(url)
	if err != nil {
		return "", err
	}

	pb, err := template.ParseBytes(b, gbl)
	if err != nil {
		return "", err
	}

	return string(pb), nil
}

func ParseRestQuery(yql *YamlQueryBlock, env *EnvFile) (*http.Request, error) {
	var restBlock RestBlock
	if err := json.Unmarshal(yql.YAMLQuery, &restBlock); err != nil {
		return nil, err
	}

	url, err := parseUrl(restBlock.Query.Url, yql.Global)
	if err != nil {
		return nil, err
	}
	body, err := parseBody(restBlock.Body, yql.Global)
	req, err := http.NewRequest(restBlock.Query.Method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
