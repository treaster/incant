package content_file

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/treaster/golist"
	"gopkg.in/yaml.v3"
)

type context struct {
	fileLoader func(string) ([]byte, error)

	inProgress *golist.Set[string]
	allResults map[string]any
	stack      []string
	errors     []error
}

func (ctx *context) addError(s string, args ...any) {
	errorStr := fmt.Sprintf(s, args...)
	stackStr := fmt.Sprintf("stack: [%s]", strings.Join(ctx.stack, " -> "))
	ctx.errors = append(ctx.errors, fmt.Errorf("%s (%s)", errorStr, stackStr))
}

func EvalContentFile(
	fileLoader func(string) ([]byte, error),
	filePath string) (
	any, []error) {

	ctx := context{
		fileLoader,
		golist.NewSet[string](),
		map[string]any{},
		[]string{},
		nil,
	}

	result := evalOneFile(&ctx, filePath)

	if len(ctx.errors) > 0 {
		for _, err := range ctx.errors {
			fmt.Println(err.Error())
		}
		return nil, ctx.errors
	}

	return result, ctx.errors
}

func evalOneFile(ctx *context, contentPath string) any {
	if ctx.inProgress.Has(contentPath) {
		ctx.addError("circular reference with %q", contentPath)
		return nil
	}

	fileContent, isProcessed := ctx.allResults[contentPath]
	if isProcessed {
		return fileContent
	}

	origContentBytes, err := ctx.fileLoader(contentPath)
	if err != nil {
		ctx.addError("unable to load content file %q: %s", contentPath, err.Error())
		return nil
	}

	var origContent any
	err = yaml.Unmarshal(origContentBytes, &origContent)
	if err != nil {
		ctx.addError("error decoding file %s: %s", contentPath, err.Error())
		return nil
	}

	ctx.inProgress.Add(contentPath)
	value := evalValue(ctx, fmt.Sprintf("file:%s", contentPath), reflect.ValueOf(origContent))
	ctx.inProgress.Remove(contentPath)

	ctx.allResults[contentPath] = value

	return value
}

func evalValue(ctx *context, stackKey string, contentValue reflect.Value) any {
	ctx.stack = append(ctx.stack, stackKey)
	defer func() {
		ctx.stack = ctx.stack[:len(ctx.stack)-1]
	}()

	switch contentValue.Kind() {
	case reflect.Map:
		subMap := map[string]any{}
		iter := contentValue.MapRange()
		for iter.Next() {
			var newValue any
			if iter.Key().Kind() != reflect.String {
				ctx.addError("non-string key in data map.")
			} else {
				newValue = evalValue(ctx, iter.Key().String(), iter.Value())
			}
			subMap[iter.Key().String()] = newValue
		}
		return subMap
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		numElements := contentValue.Len()
		subArr := make([]any, 0, numElements)
		for i := 0; i < numElements; i++ {
			newValue := evalValue(ctx, fmt.Sprintf("[%d]", i), contentValue.Index(i))
			subArr = append(subArr, newValue)
		}
		return subArr
	case reflect.Interface:
		fallthrough
	case reflect.Pointer:
		return evalValue(ctx, "*", contentValue.Elem())
	case reflect.String:
		s := contentValue.String()
		switch {
		case strings.HasPrefix(s, "file:"):
			filePath := safeCutPrefix(s, "file:")
			return evalOneFile(ctx, filePath)
		default:
			return s
		}
	default:
		return contentValue.Interface()
	}
}

// TODO(treaster): Rely on the implementation in processor/util.go
func safeCutPrefix(s string, prefix string) string {
	s, hasPrefix := strings.CutPrefix(s, prefix)
	if !hasPrefix {
		panic("Whaa?")
	}
	return s
}
