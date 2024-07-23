package processor

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/itchyny/gojq"
	"gopkg.in/yaml.v3"
)

func FindFiles(fileRoot string) []string {
	var files []string
	filepath.WalkDir(fileRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println(err.Error())
		}
		baseName := filepath.Base(path)
		if baseName[0] == '.' {
			fmt.Printf("skipping dotfile %s\n", path)
			return nil
		}
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	return files
}

func FindFilesWithName(fileRoot string, targetName string) []string {
	var files []string
	filepath.WalkDir(fileRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println(err.Error())
		}
		baseName := filepath.Base(path)
		if baseName == targetName {
			files = append(files, path)
		}
		return nil
	})

	return files
}

func Printfln(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

func Errorfln(format string, args ...any) bool {
	fmt.Printf(format+"\n", args...)
	return true
}

// From https://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file/74107689#74107689
//
// Copy copies the contents of the file at srcpath to a regular file
// at dstpath. If the file named by dstpath already exists, it is
// truncated. The function does not copy the file mode, file
// permission bits, or file attributes.
func Copy(srcpath, dstpath string) (err error) {
	r, err := os.Open(srcpath)
	if err != nil {
		return err
	}
	defer r.Close() // ignore error: file was opened read-only.

	w, err := os.Create(dstpath)
	if err != nil {
		return err
	}

	defer func() {
		// Report the error, if any, from Close, but do so
		// only if there isn't already an outgoing error.
		if c := w.Close(); err == nil {
			err = c
		}
	}()

	_, err = io.Copy(w, r)
	return err
}

func TrimExt(path string) (string, string) {
	ext := filepath.Ext(path)
	noExt, hasExt := strings.CutSuffix(path, ext)
	if !hasExt {
		panic("Whaaa?")
	}
	return noExt, ext

}

func SafeCutPrefix(s string, prefix string) string {
	s, hasPrefix := strings.CutPrefix(s, prefix)
	if !hasPrefix {
		panic("Whaa?")
	}
	return s
}

func AssertNonEmpty(s string) {
	if s == "" {
		panic("unexpected empty string")
	}
}

/*
func MakeItemSort(sortKey string) func(Item, Item) int {
	return func(a, b Item) int {
		keyA, hasKey := a.Data[sortKey]
		if !hasKey {
			panic("missing sort key in item")
		}

		keyB, hasKey := b.Data[sortKey]
		if !hasKey {
			panic("missing sort key in item")
		}

		valueA := reflect.ValueOf(keyA)
		valueB := reflect.ValueOf(keyB)

		if valueA.Type() != valueB.Type() {
			panic("mismatched compare types")
		}

		switch {
		case valueA.CanUint():
			return compare(valueA.Uint(), valueB.Uint())
		case valueA.CanInt():
			return compare(valueA.Int(), valueB.Int())
		case valueA.CanFloat():
			return compare(valueA.Float(), valueB.Float())
		case valueA.Kind() == reflect.String:
			return compare(valueA.String(), valueB.String())
		default:
			panic("unsupported sortkey type")
		}
	}
}
*/

type sortable interface {
	uint | uint8 | uint16 | uint32 | uint64 | int | int8 | int16 | int32 | int64 | float32 | float64 | string
}

func compare[K sortable](v1 K, v2 K) int {
	if v1 < v2 {
		return -1
	}
	if v1 > v2 {
		return 1
	}
	return 0
}

func FilterBySelector(selector string, siteData any) []any {
	query, err := gojq.Parse(selector)
	if err != nil {
		panic(err.Error())
	}

	var matches []any
	iter := query.Run(siteData)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		err, isErr := v.(error)
		if isErr {
			haltErr, isHalt := err.(*gojq.HaltError)
			if isHalt && haltErr.Value() == nil {
				break
			}

			panic(fmt.Sprintf("gojq error: %q", err.Error()))
		}

		matches = append(matches, v)
	}

	Printfln("SELECTOR %q found %d matches", selector, len(matches))
	return matches
}

func EvalOutputBase(expr string, itemData any) string {
	parts := strings.SplitN(expr, ":", 2)
	if len(parts) != 2 {
		panic("Output requires a colon-delimited expression")
	}

	exprType := parts[0]
	outputExpr := parts[1]

	switch exprType {
	case "jq":
		query, err := gojq.Parse(outputExpr)
		if err != nil {
			panic(err.Error())
		}

		var matches []any
		iter := query.Run(itemData)
		for {
			v, ok := iter.Next()
			if !ok {
				break
			}

			err, isErr := v.(error)
			if isErr {
				haltErr, isHalt := err.(*gojq.HaltError)
				if isHalt && haltErr.Value() == nil {
					break
				}

				panic(fmt.Sprintf("gojq error: %q", err.Error()))
			}

			matches = append(matches, v)
		}

		if len(matches) != 1 {
			panic(fmt.Sprintf("bad matches for query %q -> %+v", expr, matches))
		}
		return matches[0].(string)
	default:
		panic(fmt.Sprintf("unexpected expr type %s", exprType))
	}
}

func LoadYamlFile(filepath string, output any) error {
	fileBytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(fileBytes, output)
}
