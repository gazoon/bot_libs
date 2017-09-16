package utils

import (
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"

	"sort"
	"strings"

	"github.com/gazoon/bot_libs/logging"
	"gopkg.in/go-playground/validator.v9"
)

var (
	Validate *validator.Validate
)

func init() {
	Validate = validator.New()
}

var (
	gLogger = logging.WithPackage("utils")
)

func WaitingForShutdown() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	gLogger.Infof("Received shutdown signal: %s", <-ch)
}

func FunctionName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func FilterDuplicatesInsensitive(texts []string) []string {
	set := make(map[string]bool, len(texts))
	filtered := make([]string, 0, len(texts))
	for _, text := range texts {
		lowered := strings.ToLower(text)
		if inSet := set[lowered]; inSet {
			continue
		}
		set[lowered] = true
		filtered = append(filtered, text)
	}
	return filtered
}

func SortByOccurrence(texts []string, occurredText string) {
	occurredIndexes := make([]int, len(texts))
	for i, text := range texts {
		occurredIndexes[i] = strings.Index(text, occurredText)
	}
	sort.Slice(texts, func(i, k int) bool {
		firstIndex := occurredIndexes[i]
		secondIndex := occurredIndexes[k]
		if firstIndex < 0 {
			return false
		}
		if secondIndex < 0 {
			return true
		}
		return firstIndex < secondIndex
	})
}
