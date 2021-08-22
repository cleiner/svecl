package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/evanw/esbuild/pkg/cli"

	"github.com/cleiner/svecl/internal/importmap"
	"github.com/cleiner/svecl/internal/svelte"
)

type colors struct {
	Bold      string
	Default   string
	Dim       string
	Underline string
}

func helpText(version string, colors colors) string {
	return `
` + colors.Bold + `Usage:` + colors.Default + `
  svecl [options] [entry points]

` + colors.Bold + `Derived from:` + colors.Default + `
  ` + colors.Underline + `https://esbuild.github.io/` + colors.Default + `
  to include support for import maps and Svelte components

  ` + colors.Bold + `Repository:` + colors.Default + `
  ` + colors.Underline + `https://github.com/cleiner/svecl` + colors.Default + `

  ` + colors.Bold + `Simple options:` + colors.Default + `
  --define:K=V          Substitute K with V while parsing
  --external:M          Exclude module M from the bundle (can use * wildcards)
  --format=...          Output format (iife | cjs | esm, no default when not
                        bundling, otherwise default is iife when platform
                        is browser and cjs when platform is node)
  --import-map=...      Use the given import map to resolve imports
                        (https://github.com/wicg/import-maps)
  --loader:X=L          Use loader L to load file extension X, where L is
                        one of: js | jsx | ts | tsx | json | text | base64 |
                        file | dataurl | binary
  --minify              Minify the output (sets all --minify-* flags)
  --outdir=...          The output directory (for multiple entry points)
  --outfile=...         The output file (for one entry point)
  --platform=...        Platform target (browser | node | neutral,
                        default browser)
  --sourcemap           Emit a source map
  --splitting           Enable code splitting (currently only for esm)
  --svelte=...          The path to the Svelte distribution (compiler.js) to
                        use to compile components instead of the built-in one
  --target=...          Environment target (e.g. es2017, chrome58, firefox57,
                        safari11, edge16, node10, default esnext)
` + colors.Bold + `Advanced options:` + colors.Default + `
  --banner:T=...            Text to be prepended to each output file of type T
                            where T is one of: css | js
  --charset=utf8            Do not escape UTF-8 code points
  --log-limit=...           Maximum message count or 0 to disable (default 10)
  --footer:T=...            Text to be appended to each output file of type T
                            where T is one of: css | js
  --global-name=...         The name of the global for the IIFE format
  --inject:F                Import the file F into all input files and
                            automatically replace matching globals with imports
  --jsx-factory=...         What to use for JSX instead of React.createElement
  --jsx-fragment=...        What to use for JSX instead of React.Fragment
  --keep-names              Preserve "name" on functions and classes
  --log-level=...           Disable logging (info | warning | error | silent,
                            default info)
  --metafile=...            Write metadata about the build to a JSON file
  --minify-whitespace       Remove whitespace in output files
  --minify-identifiers      Shorten identifiers in output files
  --minify-syntax           Use equivalent but shorter syntax in output files
  --out-extension:.js=.mjs  Use a custom output extension instead of ".js"
  --outbase=...             The base path used to determine entry point output
                            paths (for multiple entry points)
  --preserve-symlinks       Disable symlink resolution for module lookup
  --public-path=...         Set the base URL for the "file" loader
  --pure:N                  Mark the name N as a pure function for tree shaking
  --resolve-extensions=...  A comma-separated list of implicit extensions
                            (default ".tsx,.ts,.jsx,.js,.css,.json")
  --sourcemap=external      Do not link to the source map with a comment
  --sourcemap=inline        Emit the source map with an inline data URL
  --sources-content=false   Omit "sourcesContent" in generated source maps
  --tree-shaking=...        Set to "ignore-annotations" to work with packages
                            that have incorrect tree-shaking annotations
  --tsconfig=...            Use this tsconfig.json file instead of other ones
  --version                 Print the current version (` + version + `) and exit

` + colors.Bold + `Examples:` + colors.Default + `
  ` + colors.Dim + `# Produces dist/entry_point.js and dist/entry_point.js.map` + colors.Default + `
  svecl entry_point.js --outdir=dist --minify --sourcemap
  ` + colors.Dim + `# Substitute the identifier RELEASE for the literal true` + colors.Default + `
  svecl example.js --outfile=out.js --define:RELEASE=true
  ` + colors.Dim + `# Resolve bare imports using the mappings specified in importmap.json` + colors.Default + `
  svecl js/main.js --import-map=js/importmap.json --outfile=bundle.js
`
}

// Run parses the provided options and executes the build
func Run(osArgs []string, version string) int {
	if len(osArgs) == 0 {
		fmt.Println(helpText(version, colors{}))
		return 0
	}

	importMap := ""
	svelteCompilerPath := ""
	argsEnd := 0
	for _, arg := range osArgs {
		switch {
		case arg == "-h", arg == "--help", arg == "/?":
			fmt.Println(helpText(version, colors{}))
			return 0

		case arg == "--version":
			compiler, _ := svelte.NewCompiler("")
			fmt.Printf("%s (Svelte %s)\n", version, compiler.Version())
			return 0

		case strings.HasPrefix(arg, "--import-map="):
			importMap = arg[len("--import-map="):]

		case strings.HasPrefix(arg, "--svelte="):
			svelteCompilerPath = arg[len("--svelte="):]

		default:
			// remove (overwrite) handled arguments
			osArgs[argsEnd] = arg
			argsEnd++
		}
	}
	osArgs = osArgs[:argsEnd]

	options, err := cli.ParseBuildOptions(osArgs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	options.Bundle = true
	options.Write = true
	options.LogLevel = api.LogLevelInfo
	options.Plugins = []api.Plugin{svelte.NewSvelteLoader(svelteCompilerPath)}

	if len(importMap) > 0 {
		options.Plugins = append(options.Plugins, importmap.NewResolver(importMap))
	}

	result := api.Build(options)

	if len(result.Errors) > 0 {
		return 1
	}

	return 0
}
