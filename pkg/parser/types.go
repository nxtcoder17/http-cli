package parser

import (
	"bufio"
	"errors"
	"io"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
)

type GlobalVars map[string]any

type YamlQueryBlock struct {
	Global    GlobalVars `json:"global"`
	YAMLQuery []byte     `json:"yamlQuery"`
}

type EnvFile struct {
	Mode string `json:"mode"`
	Map  map[string]struct {
		Url     string            `json:"url"`
		Headers map[string]string `json:"headers"`
		Vars    GlobalVars        `json:"vars"`
	} `json:"map"`
}

func ReadQueryFile(file string, lineNo uint) (*YamlQueryBlock, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)

	var currLine uint = 0
	var blockStart uint = 0
	var isInBlock bool

	var yQuery YamlQueryBlock

	lines := make([]string, 0, lineNo+10)
	for {
		readString, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				readString = "---"
				break
			}
		}

		currLine += 1
		lines = append(lines, readString)
		// fmt.Println("[ READ ]", readString, len(readString))
		if strings.TrimSpace(readString) == "---" {
			if !isInBlock {
				isInBlock = true
				blockStart = currLine
				continue
			}

			if strings.HasPrefix(lines[blockStart], "global:") {
				s := strings.Join(lines[blockStart:currLine], "")
				var m struct {
					Global map[string]any `json:"global"`
				}
				if err := yamlutil.Unmarshal([]byte(s), &m); err != nil {
					return nil, err
				}
				blockStart = currLine
				yQuery.Global = m.Global
				continue
			}

			if currLine > lineNo {
				yQuery.YAMLQuery = []byte(strings.Join(lines[blockStart:currLine-1], ""))
				break
			}

			blockStart = currLine
		}
	}

	return &yQuery, nil
}

func ParseEnvFile(f string) (*EnvFile, error) {
	b, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	var ev EnvFile
	if err := yaml.Unmarshal(b, &ev); err != nil {
		return nil, err
	}
	return &ev, nil
}
