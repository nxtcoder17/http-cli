package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nxtcoder17/http-cli/pkg/parser"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var debug = false

func showOutput(label string, msg any) {
	fmt.Printf("\n# %s\n", label)
	switch t := msg.(type) {
	case io.ReadCloser:
		b, err := io.ReadAll(t)
		if err != nil {
			panic(err)
		}

		nb := new(bytes.Buffer)
		if err := json.Indent(nb, b, "", "  "); err != nil {
			fmt.Println(errors.Wrapf(err, "[ERROR]: failed to decode response body into json format"))
			fmt.Print("\n### RAW body:\n")
			fmt.Printf("```\n%s\n```\n", b)
			return
		}
		fmt.Printf("%s\n", nb.String())
	case []byte:
		fmt.Printf("%s\n", t)
	case string:
		fmt.Printf("%s\n", t)
	case http.Header:
		m := make(map[string]any, len(t))
		for k, v := range t {
			m[k] = strings.Join(v, ", ")
		}

		b, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("```json\n%s\n```\n", b)
	default:
		b, err := json.MarshalIndent(msg, "", "  ")
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("```json\n%s\n```\n", b)
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

				if debug {
					showOutput("raw query", string(queryBlock.YAMLQuery))
				}

				ef, err := parser.ParseEnvFile(envFile)
				if err != nil {
					return err
				}
				if debug {
					showOutput("Parsed Env", *ef)
				}

				req, err := parser.ParseGqlQuery(queryBlock, ef)
				if err != nil {
					return err
				}

				showOutput("request", map[string]any{
					"url":     req.URL.String(),
					"headers": req.Header,
				})

				start := time.Now()

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return err
				}

				d := time.Since(start)

				showOutput("time taken", fmt.Sprintf("%.2fs", d.Seconds()))
				showOutput("response headers", resp.Header)
				showOutput("response body\n```json", resp.Body)
				fmt.Println("```")
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

				if debug {
					showOutput("global block", queryBlock.Global)
					showOutput("query", string(queryBlock.YAMLQuery))
				}

				ef, err := parser.ParseEnvFile(envFile)
				if err != nil {
					return err
				}

				if debug {
					showOutput("env file", *ef)
				}

				req, err := parser.ParseRestQuery(queryBlock, ef)
				if err != nil {
					return err
				}

				showOutput("request", map[string]any{
					"url":     req.URL.String(),
					"headers": req.Header,
				})

				start := time.Now()

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return err
				}

				d := time.Since(start)

				showOutput("time taken", fmt.Sprintf("%.2fs", d.Seconds()))
				showOutput("response headers", resp.Header)

				showOutput("response status", resp.Status)

				showOutput("response body\n```json", resp.Body)
				fmt.Println("```")
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
