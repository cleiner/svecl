# svecl

A single executable to compile and bundle [Svelte](https://svelte.dev) applications, with
support for [import maps](https://github.com/wicg/import-maps).

## Getting Started

svecl is derived from [esbuild](https://esbuild.github.io/), but only implements its `build` command.

Example usage:
`svecl js/main.js --import-map=js/importmap.json --sourcemap=inline --outfile=bundle.js`

Run with `--help` or `-h` for options.

## Development

Note: A C++ compiler is required to build and link with the v8go dependency. On Windows you can follow
the instructions for MSYS2 as described here: https://github.com/rogchap/v8go#windows

Build: `go build ./cmd/...`

Run tests: `go test ./...`

## TODOs

- add support for Svelte compiler options (--svelte-opt=dev:true,key:value ?)
