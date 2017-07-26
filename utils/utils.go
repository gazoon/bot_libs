package utils

import (
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"

	"github.com/gazoon/bot_libs/logging"
)

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
