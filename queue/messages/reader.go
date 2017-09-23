package msgsqueue

import (
	"context"
	"sync"

	"github.com/gazoon/bot_libs/logging"
	"github.com/gazoon/bot_libs/utils"
)

type MsgHandler func(ctx context.Context, msg *Message)
type Reader struct {
	queue      ReadQueue
	workersNum int
	onMessage  MsgHandler
	wg         sync.WaitGroup
}

func NewReader(queue ReadQueue, workersNum int, onMessage MsgHandler) *Reader {
	return &Reader{queue: queue, workersNum: workersNum, onMessage: onMessage}
}

func (r *Reader) Start() {
	gLogger.WithField("workers_num", r.workersNum).Info("Listening for incoming messages")
	for i := 0; i < r.workersNum; i++ {
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			for {
				gLogger.Info("Fetching new msg from incoming queue")
				msg, processingID, ok := r.queue.GetNext()
				if !ok {
					return
				}
				ctx := utils.PrepareContext(msg.RequestID)
				logger := logging.FromContextAndBase(ctx, gLogger).WithField("processing_id", processingID)
				logger.WithField("msg", msg).Info("Message received from incoming queue")
				r.onMessage(ctx, msg)
				logger.Info("Finish processing incoming message")
				r.queue.FinishProcessing(ctx, processingID)
			}
		}()
	}
}

func (r *Reader) Stop() {
	gLogger.Info("Close incoming queue for reading")
	r.queue.StopGivingMsgs()
	gLogger.Info("Waiting until all workers will process the remaining messages")
	r.wg.Wait()
	gLogger.Info("All workers've been stopped")
}
