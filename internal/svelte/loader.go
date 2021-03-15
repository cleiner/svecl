package svelte

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/evanw/esbuild/pkg/api"
)

type compilerPool chan *Compiler

func newCompilerPool(size int) compilerPool {
	pool := make(compilerPool, size)
	var wg sync.WaitGroup
	for i := 0; i < cap(pool); i++ {
		wg.Add(1)
		go initPoolEntry(&pool, &wg)
	}
	wg.Wait()
	return pool
}

func initPoolEntry(pool *compilerPool, wg *sync.WaitGroup) {
	defer wg.Done()
	svc, err := NewCompiler()
	if err != nil {
		panic(err)
	}
	pool.put(svc)
}

func (p compilerPool) get() *Compiler {
	return <-p
}

func (p compilerPool) put(svc *Compiler) {
	p <- svc
}

// Loader compiles .svelte files
var Loader = api.Plugin{
	Name: "svelte",
	Setup: func(build api.PluginBuild) {
		sveltePool := newCompilerPool(runtime.NumCPU())

		callback := func(args api.OnLoadArgs) (api.OnLoadResult, error) {
			bytes, err := os.ReadFile(args.Path)
			if err != nil {
				return api.OnLoadResult{}, err
			}
			svelte := sveltePool.get()
			defer sveltePool.put(svelte)
			result, err := svelte.Compile(string(bytes), filepath.Base(args.Path))
			if err != nil {
				return api.OnLoadResult{}, err
			}
			return api.OnLoadResult{
				Contents: result.Code,
				Loader:   api.LoaderJS,
				Errors:   convertMessages(result.Messages, "error"),
				Warnings: convertMessages(result.Messages, "warning"),
			}, nil
		}

		build.OnLoad(api.OnLoadOptions{Filter: `\.svelte$`}, callback)
	},
}

func convertMessages(messages *[]Message, kind string) []api.Message {
	result := []api.Message{}
	for _, src := range *messages {
		if src.Type != kind {
			continue
		}
		dst := api.Message{
			Text: src.Message,
			Location: &api.Location{
				Line:     src.Start.Line,
				Column:   src.Start.Column,
				LineText: findLineText(src.Frame, src.Start.Line),
				File:     src.Filename,
				Length:   src.End.Character - src.Start.Character,
			},
		}
		result = append(result, dst)
	}
	return result
}

func findLineText(frame string, line int) string {
	for _, srcLine := range strings.Split(frame, "\n") {
		prefix := fmt.Sprint(line) + ": "
		if strings.HasPrefix(srcLine, prefix) {
			return srcLine[len(prefix):]
		}
	}
	return frame
}
