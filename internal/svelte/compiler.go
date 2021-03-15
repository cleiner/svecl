package svelte

import (
	"embed"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/cleiner/quickjs"
)

//go:embed resources
var resources embed.FS

const resourcePath = "resources/"

var globalScripts = []string{
	"url-polyfill.js", "env.js", "base64.js", "compiler.js", "main.js",
}

// Compiler compiles Svelte components to JavaScript
type Compiler struct {
	ctx *quickjs.Context
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
func NewCompiler() (*Compiler, error) {
	ctx, err := newJavaScriptEngine()
	if err != nil {
		return nil, err
	}
	err = initGlobalScope(ctx)
	if err != nil {
		return nil, err
	}
	return &Compiler{ctx: ctx}, nil
}

func newJavaScriptEngine() (*quickjs.Context, error) {
	runtime := quickjs.NewRuntime()
	ctx := runtime.NewContext()
	// TODO runtime + ctx are never freed
	return ctx, nil
}

func initGlobalScope(ctx *quickjs.Context) error {
	for _, file := range globalScripts {
		result, err := evalScript(resourcePath+file, ctx)
		if err != nil {
			return fmt.Errorf("%+v", err)
		}
		defer result.Free()
	}
	return nil
}

func evalScript(path string, ctx *quickjs.Context) (*quickjs.Value, error) {
	bytes, err := resources.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result, err := ctx.EvalFile(string(bytes), filepath.Base(path))
	if err != nil {
		return nil, formatJSError(err)
	}
	return &result, nil
}

func formatJSError(err error) error {
	var evalErr *quickjs.Error
	if errors.As(err, &evalErr) {
		return fmt.Errorf("JS: %s\n%s\n'%+v'", evalErr.Cause, evalErr.Stack, evalErr.Error())
	}
	return err
}

// Compile translates a Svelte component into vanilla JavaScript
func (svc *Compiler) Compile(source string, filename string) (*CompileResult, error) {
	svc.ctx.Globals().Set("source", svc.ctx.String(source))
	svc.ctx.Globals().Set("filename", svc.ctx.String(filename))
	val, err := svc.ctx.EvalFile("compile(source, { filename });", "compile_call")
	if err != nil {
		return nil, formatJSError(err)
	}
	defer val.Free()
	code := val.Get("code").String()
	result := CompileResult{
		Code:     &code,
		Messages: convertMessagesFromJS(val.Get("messages")),
	}
	return &result, nil
}

func convertMessagesFromJS(messages quickjs.Value) *[]Message {
	result := make([]Message, messages.Len())
	for i := 0; i < int(messages.Len()); i++ {
		msg := messages.GetByUint32(uint32(i))
		result[i] = Message{
			Type:     msg.Get("type").String(),
			Code:     msg.Get("code").String(),
			Message:  msg.Get("message").String(),
			Filename: msg.Get("filename").String(),
			Frame:    msg.Get("frame").String(),
			Start:    *convertLocationFromJS(msg.Get("start")),
			End:      *convertLocationFromJS(msg.Get("end")),
		}
	}
	return &result
}

func convertLocationFromJS(loc quickjs.Value) *Location {
	return &Location{
		Line:      int(loc.Get("line").Int32()),
		Column:    int(loc.Get("column").Int32()),
		Character: int(loc.Get("character").Int32()),
	}
}

// Version returns the version of the embedded Svelte compiler
func (svc *Compiler) Version() string {
	val, err := svc.ctx.EvalFile("svelte.VERSION", "version_call")
	if err != nil {
		panic(formatJSError(err))
	}
	defer val.Free()
	return val.String()
}
