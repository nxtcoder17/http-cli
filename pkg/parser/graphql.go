package parser

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/nxtcoder17/http-cli/pkg/template"
	"sigs.k8s.io/yaml"
)

type GqlBlock struct {
	Label     string         `json:"label,omitempty"`
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

func ParseGqlQuery(yql *YamlQueryBlock, env *EnvFile) (*http.Request, error) {
	var gqlBlock GqlBlock
	if err := yaml.Unmarshal(yql.YAMLQuery, &gqlBlock); err != nil {
		return nil, err
	}

	vBytes, err := json.Marshal(gqlBlock.Variables)
	if err != nil {
		return nil, err
	}

	vars := make(map[string]any, len(env.Map[env.Mode].Vars)+len(gqlBlock.Variables))
	for k, v := range env.Map[env.Mode].Vars {
		vars[k] = v
	}
	for k, v := range yql.Global {
		vars[k] = v
	}

	pvBytes, err := template.ParseBytes(vBytes, vars)
	if err != nil {
		return nil, err
	}

	var pVars map[string]any
	if err := json.Unmarshal(pvBytes, &pVars); err != nil {
		return nil, err
	}

	body := map[string]any{
		"query":     gqlBlock.Query,
		"variables": pVars,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, env.Map[env.Mode].Url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range env.Map[env.Mode].Headers {
		req.Header.Set(k, v)
	}
	req.Header.Add("Accept-Charset", "utf-8")

	return req, nil
}
