package importmap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

type specifierMap map[string]string

type importMap struct {
	Imports specifierMap
	Scopes  map[string]specifierMap
}

func newImportMapFromFile(filename string) (*importMap, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	basePath, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return nil, fmt.Errorf("Unable to determine base path: %w", err)
	}
	basePath = basePath + string(filepath.Separator)
	return newImportMapFromJSON(bytes, basePath)
}

func newImportMapFromJSON(content []byte, basePath string) (*importMap, error) {
	var importMap importMap
	err := json.Unmarshal(content, &importMap)
	if err != nil {
		return nil, err
	}
	if len(importMap.Scopes) > 0 {
		return nil, fmt.Errorf("Scopes are not supported")
	}
	normalizeResolutionResults(importMap.Imports, basePath)
	return &importMap, nil
}

func normalizeResolutionResults(specifierMap specifierMap, basePath string) {
	for k, v := range specifierMap {
		specifierMap[k] = basePath + filepath.FromSlash(v)
	}
}

func (m *importMap) resolve(specifier string) (string, error) {
	if result, found := m.Imports[specifier]; found {
		return result, nil
	}
	result := ""
	for k, v := range m.Imports {
		if strings.HasSuffix(k, "/") && strings.HasPrefix(specifier, k) && len(result) < len(k) {
			result = strings.Replace(specifier, k, v, 1)
		}
	}
	if len(result) == 0 {
		return "", fmt.Errorf("Unable to resolve specifier '%s'", specifier)
	}
	return filepath.Clean(result), nil
}

func (m *importMap) isRelativeOrAbsolute(specifier string) bool {
	return strings.HasPrefix(specifier, "../") || strings.HasPrefix(specifier, "./") || filepath.IsAbs(specifier)
}

// NewResolver creates a new resolver plugin using the given import map
func NewResolver(importMapPath string) api.Plugin {
	return api.Plugin{
		Name: "import-map",
		Setup: func(build api.PluginBuild) {
			importMap, setupErr := newImportMapFromFile(importMapPath)

			callback := func(args api.OnResolveArgs) (api.OnResolveResult, error) {
				if setupErr != nil {
					return api.OnResolveResult{}, setupErr
				}
				if importMap.isRelativeOrAbsolute(args.Path) {
					return api.OnResolveResult{}, nil
				}
				resolvedPath, err := importMap.resolve(args.Path)
				if err != nil {
					return api.OnResolveResult{}, err
				}
				if len(filepath.Ext(resolvedPath)) == 0 {
					// heuristic concession to existing code with NodeJS-style imports
					resolvedPath = resolvedPath + "/index.js"
				}
				return api.OnResolveResult{
					Path:      resolvedPath,
					Namespace: "file",
				}, nil
			}

			build.OnResolve(api.OnResolveOptions{Filter: `.*`}, callback)
		},
	}
}
