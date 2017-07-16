package utils

import (
	"github.com/gazoon/bot_libs/logging"
	"os"
	"os/signal"
	"syscall"
)

var (
	gLogger = logging.WithPackage("utils")
)

func WaitingForShutdown() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	gLogger.Infof("Received shutdown signal: %s", <-ch)
}
