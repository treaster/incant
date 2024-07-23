package processor

import (
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/hjson/hjson-go/v4"
	"github.com/treaster/ssg/processor/json5"
	"gopkg.in/yaml.v3"
)

type FileLoader struct {
	// e.g. os.ReadFile
	ReadFileFn func(string) ([]byte, error)
}

func (l FileLoader) LoadFile(s string, output any) error {
	ext := filepath.Ext(s)

	fileBytes, err := l.ReadFileFn(s)
	if err != nil {
		return err
	}

	switch ext {
	case ".yaml":
		return yaml.Unmarshal(fileBytes, output)
	case ".toml":
		_, err := toml.Decode(string(fileBytes), output)
		return err
	case ".json5":
		return json5.Unmarshal(fileBytes, output)
	case ".hjson":
		return hjson.Unmarshal(fileBytes, output)
	default:
		panic(fmt.Sprintf("Unknown extension on file path %q", s))
	}
}
