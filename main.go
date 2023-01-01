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

	"github.com/nxtcoder17/http-cli/pkg/parser"
	"github.com/urfave/cli/v2"
)

func printHeaders(headers http.Header) {
	for k, v := range headers {
		fmt.Printf("%-20s: %-30s\n", k, strings.Join(v, ","))
	}
}

var debug = true

func showOutput(msg any) {
	switch t := msg.(type) {
	case io.ReadCloser:
		b, err := io.ReadAll(t)
		if err != nil {
			panic(err)
		}

		if debug {
      fmt.Printf("[RAW RESPONSE]\n%s\n", string(b))
		}

		nb := new(bytes.Buffer)
		if err := json.Indent(nb, b, "", "  "); err != nil {
		  fmt.Println(err)
		  return
		}
    fmt.Printf("\n[JSON RESPONSE]\n%s\n", nb.String())
	default:
		if debug {
      fmt.Printf("[RAW RESPONSE]\n%v\n", msg)
		}
		b, err := json.MarshalIndent(msg, "", "  ")
		if err != nil {
		  fmt.Println(err)
		  return
		}
		fmt.Printf("%s\n", b)
	}
}

func printHttpResponse(resp io.ReadCloser) {
		b, err := io.ReadAll(resp)
		if err != nil {
			panic(err)
		}

		if debug {
      fmt.Printf("[RAW RESPONSE]\n%s\n", string(b))
		}

		nb := new(bytes.Buffer)
		if err := json.Indent(nb, b, "", "  "); err != nil {
		  fmt.Println(err)
		  return
		}
    fmt.Printf("\n[JSON RESPONSE]\n%s\n", nb.String())
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

				debug := true

				if debug {
					fmt.Println("global")
					showOutput(queryBlock.Global)

					fmt.Println("query")
					showOutput(string(queryBlock.YAMLQuery))
				}

				ef, err := parser.ParseEnvFile(envFile)
				if err != nil {
					return err
				}

				if debug {
					fmt.Println("env file")
					showOutput(ef)
				}

				req, err := parser.ParseRestQuery(queryBlock, ef)
				if err != nil {
					return err
				}

				fmt.Println("### Request Headers")
				printHeaders(req.Header)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return err
				}

				fmt.Println("\n### Response Headers")
				printHeaders(resp.Header)

        printHttpResponse(resp.Body)
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
