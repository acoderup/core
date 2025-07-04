package signal

import (
	"os"

	"github.com/acoderup/core/logger"
	"github.com/acoderup/core/module"
)

type InterruptSignalHandler struct {
}

func (ish *InterruptSignalHandler) Process(s os.Signal, ud interface{}) error {
	logger.Logger.Warn("Receive Interrupt signal, process start quit.")
	module.Stop()
	return nil
}

func init() {
	SignalHandlerModule.RegisteHandler(os.Interrupt, &InterruptSignalHandler{}, nil)
}
