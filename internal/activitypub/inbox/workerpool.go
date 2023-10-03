package inbox

import (
	"runtime"

	"github.com/shigde/sfu/internal/activitypub/remote"
	"golang.org/x/exp/slog"
)

// workerPoolSize defines the number of concurrent ActivityPub handlers.
var workerPoolSize = runtime.GOMAXPROCS(0)

// Job struct bundling the ActivityPub and the payload in one struct.
type Job struct {
	request InboxRequest
}

var queue chan Job

// InitInboxWorkerPool starts n go routines that await ActivityPub jobs.
func InitInboxWorkerPool(resolver *remote.Resolver) {
	queue = make(chan Job)

	// start workers
	for i := 1; i <= workerPoolSize; i++ {
		go worker(i, queue, resolver)
	}
}

// AddToQueue will queue up an outbound http request.
func AddToQueue(req InboxRequest) {
	slog.Info("Queued request for ActivityPub inbox handler")
	queue <- Job{req}
}

func worker(workerID int, queue <-chan Job, resolver *remote.Resolver) {
	slog.Debug("Started ActivityPub inbox worker", "workerId", workerID)

	for job := range queue {
		handle(job.request, resolver)

		slog.Info("Done with ActivityPub inbox handler using worker", "workerId", workerID)
	}
}
