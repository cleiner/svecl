package svelte

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"rogchap.com/v8go"
)

//go:embed resources
var resources embed.FS

const resourcePath = "resources/"

var globalScripts = []string{
	"env.js", "base64.js", "url-polyfill.js", "main.js",
}

// Compiler compiles Svelte components to JavaScript
type Compiler struct {
	ctx *v8go.Context
}

// Location represents a source file location
type Location struct {
	Line      int
	Column    int
	Character int
}

// Message is a warning or error from the Svelte compiler
type Message struct {
	Type     string
	Code     string
	Message  string
	Filename string
	Frame    string
	Start    Location
	End      Location
}

// CompileResult contains the generated code and any warnings or errors
type CompileResult struct {
	Code     *string
	Messages *[]Message
}

// NewCompiler creates a new compiler instance
func NewCompiler(svelteCompilerPath string) (*Compiler, error) {
	ctx, err := newJavaScriptEngine()
	if err != nil {
		return nil, err
	}
	err = initGlobalScope(ctx)
	if err != nil {
		return nil, err
	}
	err = loadSvelteCompiler(ctx, svelteCompilerPath)
	if err != nil {
		return nil, err
	}
	return &Compiler{ctx: ctx}, nil
}

func newJavaScriptEngine() (*v8go.Context, error) {
	iso, err := v8go.NewIsolate()
	if err != nil {
		return nil, err
	}
	ctx, err := v8go.NewContext(iso)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func initGlobalScope(ctx *v8go.Context) error {
	for _, file := range globalScripts {
		_, err := evalScript(resourcePath+file, ctx, resources.ReadFile)
		if err != nil {
			return fmt.Errorf("%+v", err)
		}
	}
	return nil
}

func evalScript(path string, ctx *v8go.Context, readFile func(path string) ([]byte, error)) (*v8go.Value, error) {
	bytes, err := readFile(path)
	if err != nil {
		return nil, err
	}
	return ctx.RunScript(string(bytes), filepath.Base(path))
}

func loadSvelteCompiler(ctx *v8go.Context, svelteCompilerPath string) error {
	basePath := resourcePath
	readFile := resources.ReadFile
	if svelteCompilerPath != "" {
		basePath = appendSeparator(svelteCompilerPath)
		readFile = os.ReadFile
	}
	_, err := evalScript(basePath+"compiler.js", ctx, readFile)
	if err != nil {
		return fmt.Errorf("%+v", err)
	}
	return nil
}

func appendSeparator(path string) string {
	separator := string(filepath.Separator)
	if strings.HasSuffix(path, separator) {
		return path
	}
	return path + separator
}

// Compile translates a Svelte component into vanilla JavaScript
func (svc *Compiler) Compile(source string, filename string) (*CompileResult, error) {
	svc.ctx.Global().Set("source", source)
	svc.ctx.Global().Set("filename", filename)
	val, err := svc.ctx.RunScript("compile(source, { filename });", "compile_call")
	if err != nil {
		return nil, fmt.Errorf("%+v", err)
	}
	result := CompileResult{}
	err = json.Unmarshal([]byte(val.String()), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Version returns the version of the embedded Svelte compiler
func (svc *Compiler) Version() string {
	val, err := svc.ctx.RunScript("svelte.VERSION", "version_call")
	if err != nil {
		panic(err)
	}
	return val.String()
}
