package main

import (
	"fmt"
	"log"
	"os"

	esbuild "github.com/evanw/esbuild/pkg/api"
)

type Spec struct {
	Infile, Outfile string
}

func main() {
	specs := []Spec{
		{
			Infile:  "./ts/exports/parser.ts",
			Outfile: "./build/parser.js",
		},
		{
			Infile:  "./ts/exports/compiler.ts",
			Outfile: "./build/compiler.js",
		},
	}
	for _, s := range specs {
		generateFile(s)
	}
}

func generateFile(s Spec) {
	res := esbuild.Build(esbuild.BuildOptions{
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifySyntax:      true,
		MinifyIdentifiers: true,
		Write:             true,
		Define:            map[string]string{"SUPIBOT": "true"},
		EntryPoints:       []string{s.Infile},
		Outfile:           s.Outfile,
		Tsconfig:          "./tsconfig.build.json",
		GlobalName:        "sbl",
		Drop:              esbuild.DropConsole,
		Target:            esbuild.ES2016,
		Platform:          esbuild.PlatformNode,
		LegalComments:     esbuild.LegalCommentsNone,
		Format:            esbuild.FormatIIFE,
		// Format:            esbuild.FormatCommonJS,
		TreeShaking: esbuild.TreeShakingTrue,
		Charset:     esbuild.CharsetASCII,
	})
	if len(res.Errors) > 0 || len(res.Warnings) > 0 {
		locationToString := func(loc *esbuild.Location) string {
			if loc != nil {
				return loc.File + ":" + fmt.Sprint(loc.Line) + ":" + fmt.Sprint(loc.Column)
			}
			return "<unknown>:?:?"
		}
		for _, m := range res.Warnings {
			log.Printf("Minify JS (warning): %s: %s\n", locationToString(m.Location), m.Text)
			for _, n := range m.Notes {
				log.Printf("Minify JS (warning): %s: Note: %s\n", locationToString(n.Location), n.Text)
			}
		}
		for _, m := range res.Errors {
			log.Printf("Minify JS: %s: %s\n", locationToString(m.Location), m.Text)
			for _, n := range m.Notes {
				log.Printf("Minify JS: %s: Note: %s\n", locationToString(n.Location), n.Text)
			}
		}
		if len(res.Errors) > 0 {
			os.Exit(1)
		}
	}
	// repr.Println(res, repr.Indent("  "), repr.OmitEmpty(true))

	fmt.Println("xd", res.OutputFiles[0].Path+":", len(res.OutputFiles[0].Contents))
}
