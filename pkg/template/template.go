package template

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/autopus/bootstrap/pkg/types"
)

func MapReplace(input types.M, ctx types.M) types.M {
	processed := make(types.M)
	for key, value := range input {
		switch v := value.(type) {
		case string:
			processed[key] = Replace(v, ctx)
		case types.M:
			processed[key] = MapReplace(v, ctx)
		default:
			processed[key] = v
		}
	}
	return processed
}

func Replace(input string, ctx types.M) string {
	re := regexp.MustCompile(`{{(.*?)}}`)
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		fullMatch := match[0]
		path := match[1]
		replacement := resolvePath(ctx, strings.Split(path, "."))
		input = strings.ReplaceAll(input, fullMatch, fmt.Sprint(replacement))
	}

	return input
}

func resolvePath(ctx types.M, path []string) any {
	if len(path) == 0 {
		return ""
	}

	current, exists := ctx[path[0]]
	if !exists {
		return ""
	}

	if len(path) == 1 {
		return current
	}

	if nextCtx, ok := current.(types.M); ok {
		return resolvePath(nextCtx, path[1:])
	}

	return ""
}
