package content_file

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/treaster/golist"
)

type Content map[string]any

type context struct {
	allContents map[string]Content
	inProgress  *golist.Set[string]
	allResults  map[string]Content
	stack       []string
	errors      []error
}

func (ctx *context) addError(s string, args ...any) {
	errorStr := fmt.Sprintf(s, args...)
	stackStr := fmt.Sprintf("stack: [%s]", strings.Join(ctx.stack, " -> "))
	ctx.errors = append(ctx.errors, fmt.Errorf("%s (%s)", errorStr, stackStr))
}

func EvalContents(allContents map[string]Content) (map[string]Content, []error) {
	allPaths := make([]string, 0, len(allContents))
	for contentPath, _ := range allContents {
		allPaths = append(allPaths, contentPath)
	}
	sort.Strings(allPaths)

	ctx := context{
		allContents,
		golist.NewSet[string](),
		map[string]Content{},
		[]string{},
		nil,
	}

	for _, contentPath := range allPaths {
		evalOneFile(&ctx, contentPath)
	}

	if len(ctx.errors) > 0 {
		for _, err := range ctx.errors {
			fmt.Println(err.Error())
		}
		return nil, ctx.errors
	}

	return ctx.allResults, ctx.errors
}

func evalOneFile(ctx *context, contentPath string) map[string]any {
	if ctx.inProgress.Has(contentPath) {
		ctx.addError("circular reference with %q", contentPath)
		return nil // map[string]any{}
	}

	fileContent, isProcessed := ctx.allResults[contentPath]
	if isProcessed {
		return map[string]any(fileContent)
	}

	origContent, hasFile := ctx.allContents[contentPath]
	if !hasFile {
		ctx.addError("unrecognized content file %q", contentPath)
		return nil // map[string]any{}
	}

	ctx.inProgress.Add(contentPath)
	value := evalValue(ctx, fmt.Sprintf("file:%s", contentPath), reflect.ValueOf(origContent))
	ctx.inProgress.Remove(contentPath)

	mapValue := value.(map[string]any)
	ctx.allResults[contentPath] = Content(mapValue)

	return mapValue
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
