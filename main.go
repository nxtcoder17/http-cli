package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/nxtcoder17/http-cli/pkg/template"
	"github.com/urfave/cli/v2"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

type GqlBlock struct {
	Label     string         `json:"label,omitempty"`
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type GqlQuery struct {
	Global   map[string]any `json:"global,omitempty"`
	GqlBlock GqlBlock       `json:"query"`
}

type EnvFile struct {
	Mode string `json:"mode"`
	Map  map[string]struct {
		Url     string            `json:"url"`
		Headers map[string]string `json:"headers"`
	} `json:"map"`
}

func showOutput(msg any) {
	switch t := msg.(type) {
	case io.ReadCloser:
		b, err := io.ReadAll(t)
		if err != nil {
			panic(err)
		}
		nb := new(bytes.Buffer)
		if err := json.Indent(nb, b, "", "  "); err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", nb.Bytes())
	default:
		b, err := json.MarshalIndent(msg, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", b)
	}
}

func main() {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		{
			Name:    "graphql",
			Aliases: []string{"g", "gq", "gql"},
			Usage:   "graphql query",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "file",
					Required: true,
					Usage:    "filename with yaml queries",
				},
				&cli.StringFlag{
					Name:     "envFile",
					Required: true,
					Usage:    "gqlenv file",
				},
				&cli.UintFlag{
					Name:     "lineNo",
					Required: true,
					Usage:    "lineNo for yaml block to be executed",
				},
			},
			Action: func(cctx *cli.Context) error {
				yamlFile := cctx.String("file")
				lineNo := cctx.Uint("lineNo")

				file, err := os.Open(yamlFile)

				if err != nil {
					return err
				}

				reader := bufio.NewReader(file)

				var currLine uint = 0
				var blockStart uint = 0
				var isInBlock bool

				var gqlQuery GqlQuery

				lines := make([]string, 0, lineNo+10)
				for {
					readString, err := reader.ReadString('\n')
					if err != nil {
						if errors.Is(err, io.EOF) {
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
								return err
							}
							blockStart = currLine
							gqlQuery.Global = m.Global
							// fmt.Printf("%+v\n", gqlQuery.Global)
							continue
						}

						if currLine > lineNo {
							s := strings.Join(lines[blockStart:currLine], "")
							if err := yamlutil.Unmarshal([]byte(s), &gqlQuery.GqlBlock); err != nil {
								return err
							}
							// fmt.Printf("%+v\n", gqlQuery.GqlBlock)
							break
						}
					}
				}

				// here i have gqlQuery

				envFile := cctx.String("envFile")
				b, err := os.ReadFile(envFile)
				if err != nil {
					return err
				}
				var gqlEnv EnvFile
				if err := yaml.Unmarshal(b, &gqlEnv); err != nil {
					return err
				}

				vBytes, err := json.Marshal(gqlQuery.GqlBlock.Variables)
				if err != nil {
					return err
				}

				parsedVars, err := template.ParseBytes(vBytes, gqlQuery.Global)
				if err != nil {
					return err
				}

				fmt.Printf("parsed :) %s\n", parsedVars)

				var nVariables map[string]any
				if err := json.Unmarshal(parsedVars, &nVariables); err != nil {
					return err
				}

				fmt.Println("### Request Body")
				body := map[string]any{
					"query":     gqlQuery.GqlBlock.Query,
					"variables": nVariables,
				}

				bodyBytes, err := json.Marshal(body)
				if err != nil {
					return err
				}

				showOutput(body)

				req, err := http.NewRequest(http.MethodPost, gqlEnv.Map[gqlEnv.Mode].Url, bytes.NewBuffer(bodyBytes))
				if err != nil {
					return err
				}
				req.Header.Set("Content-Type", "application/json")
				for k, v := range gqlEnv.Map[gqlEnv.Mode].Headers {
					req.Header.Set(k, v)
				}

				fmt.Println("### Request Headers")
				showOutput(req.Header)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return err
				}

				fmt.Println("### Response Headers")
				showOutput(resp.Header)

				fmt.Println("### Response Body")
				showOutput(resp.Body)

				return nil
			},
		},
	}

	app.Name = "http-cli"
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
