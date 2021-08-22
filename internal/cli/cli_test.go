package cli

import (
	"bytes"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"testing"
)

func redirectOutput(fn func() int) (string, int) {
	inp, out, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer inp.Close()

	stdout := os.Stdout
	os.Stdout = out
	defer func() { os.Stdout = stdout }()
	stderr := os.Stderr
	os.Stderr = out
	defer func() { os.Stderr = stderr }()

	output := make(chan []byte)
	go func() {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, inp); err != nil {
			panic(err)
		}
		output <- buf.Bytes()
	}()

	rc := fn()
	out.Close()

	return string(<-output), rc
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("%v != %v @%s", a, b, debug.Stack())
	}
}

func assertContains(t *testing.T, s string, parts ...string) {
	for _, p := range parts {
		if !strings.Contains(s, p) {
			t.Errorf("'%s' not in '%s' @%s", p, s, debug.Stack())
		}
	}
}

const version = "0.0.test"

func run(args []string) (string, int) {
	return redirectOutput(func() int {
		return Run(args, version)
	})
}

func TestNoArgs(t *testing.T) {
	output, rc := run([]string{})
	assertEqual(t, rc, 0)
	assertContains(t, output, "Usage:\n  svecl", version)
}

func TestPrintHelp(t *testing.T) {
	output, rc := run([]string{"--help"})
	assertEqual(t, rc, 0)
	assertContains(t, output, "Usage:\n  svecl", version)
}

func TestInvalidArg(t *testing.T) {
	output, rc := run([]string{"--oops"})
	assertEqual(t, rc, 1)
	assertContains(t, output, "Invalid build flag")
}

func TestPrintVersion(t *testing.T) {
	output, rc := run([]string{"--version"})
	assertEqual(t, rc, 0)
	assertContains(t, output, version, "Svelte 3.")
}

func TestImportMapMissing(t *testing.T) {
	output, rc := run([]string{"--import-map=missing.json", "ep.js"})
	assertEqual(t, rc, 1)
	assertContains(t, output, "[import-map] open missing.json")
}

func TestImportMapParseError(t *testing.T) {
	output, rc := run([]string{"--import-map=test/im_syntax.json.txt", "ep.js"})
	assertEqual(t, rc, 1)
	assertContains(t, output, "[import-map] invalid character '}'")
}

func TestImportMapWithScopes(t *testing.T) {
	output, rc := run([]string{"--import-map=test/im_scopes.json", "ep.js"})
	assertEqual(t, rc, 1)
	assertContains(t, output, "[import-map] Scopes are not supported")
}

func TestImportMapResolveError(t *testing.T) {
	output, rc := run([]string{"--import-map=test/im_ok.json", "test/js/root_missing.js"})
	assertEqual(t, rc, 1)
	assertContains(t, output, "[import-map] Unable to resolve specifier", "unmapped")
}

func TestImportMapSuccess(t *testing.T) {
	output, rc := run([]string{"--import-map=test/im_ok.json", "--minify", "test/js/root.js"})
	assertEqual(t, rc, 0)
	assertEqual(t, output, "(()=>{var f=\"@sib\";var e=\"@lib/ref\",b=()=>42,r=()=>b(),t=()=>b()+1;var i=\"@lib\";var o=\"@sub:a\";var m=o+\"+@sub:b\";var p=\"@sub:c\";console.log(f,i,e,o,m,p,r,t);})();\n")
}

func TestSvelteError(t *testing.T) {
	output, rc := run([]string{"--import-map=test/im_svelte.json", "test/js/view_error.svelte"})
	assertEqual(t, rc, 1)
	assertContains(t, output, "[svelte] Unexpected token", "view_error.svelte:7:1")
}

func TestSvelteSuccessWithWarning(t *testing.T) {
	output, rc := run([]string{"--import-map=test/im_svelte.json", "test/js/view.svelte"})
	assertEqual(t, rc, 0)
	assertContains(t, output, "warning: [svelte] A11y: <img> element should have an alt attribute")
	assertContains(t, output, "extends SvelteComponent", "var view_default = View;")
}

func TestExternalSvelte(t *testing.T) {
	output, rc := run([]string{"--import-map=test/im_svelte.json", "--svelte=../svelte/resources", "test/js/view.svelte"})
	assertEqual(t, rc, 0)
	assertContains(t, output, "extends SvelteComponent")
}
