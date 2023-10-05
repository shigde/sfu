package workerpool

import (
	"net/http"
	"runtime"

	"golang.org/x/exp/slog"
)

// workerPoolSize defines the number of concurrent HTTP ActivityPub requests.
var workerPoolSize = runtime.GOMAXPROCS(0)

// Job struct bundling the ActivityPub and the payload in one struct.
type Job struct {
	request *http.Request
}

var queue chan Job

// InitOutboundWorkerPool starts n go routines that await ActivityPub jobs.
func InitOutboundWorkerPool() {
	queue = make(chan Job)

	// start workers
	for i := 1; i <= workerPoolSize; i++ {
		go worker(i, queue)
	}
}

// AddToOutboundQueue will queue up an outbound http request.
func AddToOutboundQueue(req *http.Request) {
	slog.Info("Queued request for ActivityPub", "destination", req.RequestURI)
	queue <- Job{req}
}

func worker(workerID int, queue <-chan Job) {
	slog.Debug("Started ActivityPub worker", "workerId", workerID)

	for job := range queue {
		if err := sendActivityPubMessageToInbox(job); err != nil {
			slog.Error("ActivityPub destination failed to send", "destination", job.request.RequestURI, "err", err)
		}
		slog.Info("Done with ActivityPub destination using worker", "destination", job.request.RequestURI, "workerId", workerID)
	}
}

func sendActivityPubMessageToInbox(job Job) error {
	client := &http.Client{}

	resp, err := client.Do(job.request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}
