package processor

import (
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/hjson/hjson-go/v4"
	"github.com/treaster/incant/processor/json5"
	"gopkg.in/yaml.v3"
)

func MakeFileLoader(readFileFn func(string) ([]byte, error)) FileLoader {
	return FileLoader{
		readFileFn: readFileFn,
		typesMap: map[string]func([]byte, any) error{
			".yaml": yaml.Unmarshal,
			".toml": func(fileBytes []byte, output any) error {
				_, err := toml.Decode(string(fileBytes), output)
				return err
			},
			".json5": json5.Unmarshal,
			".hjson": hjson.Unmarshal,
		},
	}
}

type FileLoader struct {
	// e.g. os.ReadFile
	readFileFn func(string) ([]byte, error)
	typesMap   map[string]func([]byte, any) error
}

func (l FileLoader) SupportsFormat(s string) bool {
	ext := filepath.Ext(s)
	_, hasFormat := l.typesMap[ext]
	return hasFormat
}

func (l FileLoader) LoadFileAsBytes(s string) ([]byte, error) {
	return l.readFileFn(s)
}

func (l FileLoader) LoadFile(s string, output any) error {
	ext := filepath.Ext(s)

	fileBytes, err := l.readFileFn(s)
	if err != nil {
		return err
	}

	fn, hasFormat := l.typesMap[ext]
	if !hasFormat {
		panic(fmt.Sprintf("Unknown extension on file path %q", s))
	}

	return fn(fileBytes, output)
}
