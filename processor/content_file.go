package processor

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/treaster/gotl"
)

type context struct {
	loader FileLoader

	inProgress *gotl.Set[string]
	allResults map[string]any
	stack      []string
	errors     []error
}

func (ctx *context) addError(s string, args ...any) {
	errorStr := fmt.Sprintf(s, args...)
	stackStr := fmt.Sprintf("stack: [%s]", strings.Join(ctx.stack, " -> "))
	ctx.errors = append(ctx.errors, fmt.Errorf("%s (%s)", errorStr, stackStr))
}

func EvalContentFile(loader FileLoader, filePath string) (any, []error) {
	ctx := context{
		loader,
		gotl.NewSet[string](),
		map[string]any{},
		[]string{},
		nil,
	}

	result := evalOneFile(&ctx, filePath, false)

	if len(ctx.errors) > 0 {
		for _, err := range ctx.errors {
			fmt.Println(err.Error())
		}
		return nil, ctx.errors
	}

	return result, ctx.errors
}

func evalOneFile(ctx *context, contentPath string, allowAsString bool) any {
	if ctx.inProgress.Has(contentPath) {
		ctx.addError("circular reference with %q", contentPath)
		return nil
	}

	fileContent, isProcessed := ctx.allResults[contentPath]
	if isProcessed {
		return fileContent
	}

	if allowAsString && !ctx.loader.SupportsFormat(contentPath) {
		value, err := ctx.loader.LoadFileAsBytes(contentPath)
		if err != nil {
			ctx.addError("unable to load content file *as bytes* %q: %s", contentPath, err.Error())
			return nil
		}
		return string(value)
	} else {
		var origContent any
		err := ctx.loader.LoadFile(contentPath, &origContent)
		if err != nil {
			ctx.addError("unable to load content file %q: %s", contentPath, err.Error())
			return nil
		}

		ctx.inProgress.Add(contentPath)
		value := evalValue(ctx, fmt.Sprintf("file:%s", contentPath), reflect.ValueOf(origContent))
		ctx.inProgress.Remove(contentPath)

		ctx.allResults[contentPath] = value

		Printfln("Loaded data file of kind %q", reflect.ValueOf(value).Kind())

		return value
	}
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
			filePath := SafeCutPrefix(s, "file:")
			return evalOneFile(ctx, filePath, true)
		default:
			return s
		}
	default:
		return contentValue.Interface()
	}
}
