// Command testvectors regenerates protocol/testvectors deterministically (-check fails if stale).
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tv "trivia.al/backend/internal/platform/testvectors"
)

func encode(v any) []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func main() {
	check := flag.Bool("check", false, "fail if committed vectors are stale")
	flag.Parse()

	files := map[string][]byte{tv.CueFile: encode(cueSuite()), tv.ShuffleFile: encode(shuffleSuite())}
	root, err := tv.OpenDir() // confine all file access to protocol/testvectors
	if err != nil {
		panic(err)
	}
	defer func() { _ = root.Close() }()
	stale := false
	for name, want := range files {
		path := filepath.Join(tv.Dir(), name)
		if *check {
			got, err := root.ReadFile(name)
			if err != nil || !bytes.Equal(got, want) {
				fmt.Fprintf(os.Stderr, "stale: %s (run: go run ./cmd/testvectors)\n", path)
				stale = true
			}
			continue
		}
		if err := root.WriteFile(name, want, 0o600); err != nil {
			panic(err)
		}
		fmt.Println("wrote", path)
	}
	if stale {
		os.Exit(1)
	}
}
