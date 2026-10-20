// Command generate writes the Go code of the declarations of the Ergani API.
//
// The wire structs, the public models, the translation between them and the
// test that compares every payload with the reference payload of the
// Ministry are written out of one table, so that they cannot drift apart.
//
// The code it writes is committed, so the build of the module never needs
// it. Run it with:
//
//	go generate ./...
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	root := flag.String("root", "", "the directory of the module (the working directory by default)")
	samples := flag.String("samples", "", "the directory of the reference payloads of the Ministry")
	schemas := flag.String("schemas", "", "the directory of the schemas of the Ministry")
	flag.Parse()

	if err := run(*root, *samples, *schemas); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run writes the Go code of the declarations.
func run(root, samples, schemas string) error {
	module, err := moduleRoot(root)
	if err != nil {
		return err
	}

	if samples == "" {
		samples = filepath.Join(module, "testdata", "declarations")
	}
	if schemas == "" {
		schemas = filepath.Join(module, "internal", "schema", "xsds")
	}

	if err := checkNames(); err != nil {
		return err
	}

	generator := &generator{
		root:    module,
		samples: samples,
		schemas: schemas,
	}

	if err := generator.read(); err != nil {
		return err
	}

	if err := generator.emitWire(); err != nil {
		return err
	}
	if err := generator.emitPublic(); err != nil {
		return err
	}
	if err := generator.emitTranslation(); err != nil {
		return err
	}
	if err := generator.emitGoldenTest(); err != nil {
		return err
	}

	fmt.Println("wrote the wire structs, the public models and the translation")
	for _, entry := range generator.groups {
		fmt.Printf("  %-20s %3d shared, %3d optional, %d variants\n",
			entry.spec.kind, len(entry.shared), len(entry.optional), len(entry.variants))
	}

	return nil
}

// moduleRoot returns the directory of the module: the one that holds the
// go.mod, found by walking up from the working directory.
func moduleRoot(root string) (string, error) {
	if root != "" {
		return root, nil
	}

	directory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("ergani: the working directory could not be read: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory, nil
		}

		parent := filepath.Dir(directory)
		if parent == directory {
			return "", fmt.Errorf("ergani: the module could not be found: no go.mod above the working directory")
		}
		directory = parent
	}
}
