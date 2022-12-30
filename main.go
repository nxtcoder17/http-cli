package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nxtcoder17/http-cli/pkg/parser"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"net/http"
	"os"
)

type RestBlock struct {
	Label   string            `json:"label,omitempty"`
	Method  string            `json:"method,omitempty"`
	Url     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    map[string]any    `json:"body,omitempty"`
	Params  map[string]any    `json:"params,omitempty"`
}

type RestQuery struct {
	Global    map[string]any `json:"global,omitempty"`
	RestBlock RestBlock      `json:"rest"`
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

var flags = []cli.Flag{
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
}

func main() {
	app := cli.NewApp()
	app.Name = "http-cli"
	app.Commands = []*cli.Command{
		{
			Name:    "graphql",
			Aliases: []string{"g", "gq", "gql"},
			Usage:   "graphql query",
			Flags:   flags,
			Action: func(cctx *cli.Context) error {
				yamlFile := cctx.String("file")
				lineNo := cctx.Uint("lineNo")
				envFile := cctx.String("envFile")

				queryBlock, err := parser.ReadQueryFile(yamlFile, lineNo)
				if err != nil {
					return err
				}

				ef, err := parser.ParseEnvFile(envFile)
				if err != nil {
					return err
				}

				req, err := parser.ParseGqlQuery(queryBlock, ef)
				if err != nil {
					return err
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

		{
			Name:    "rest",
			Aliases: []string{"r", "rest"},
			Usage:   "rest api calls",
			Flags:   flags,
			Action: func(cctx *cli.Context) error {
				yamlFile := cctx.String("file")
				lineNo := cctx.Uint("lineNo")
				envFile := cctx.String("envFile")

				queryBlock, err := parser.ReadQueryFile(yamlFile, lineNo)
				if err != nil {
					return err
				}

				ef, err := parser.ParseEnvFile(envFile)
				if err != nil {
					return err
				}

				req, err := parser.ParseRestQuery(queryBlock, ef)
				if err != nil {
					return err
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

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
