package workerpool

import (
	"log"
)

// WorkerPool is a contract for Worker Pool implementation
type WorkerPool interface {
	Run()
	AddTask(task func())
}

type workerPool struct {
	maxWorker   int
	queuedTaskC chan func()
	QuitChan    chan bool
}

// NewWorkerPool will create an instance of WorkerPool.
func NewWorkerPool(maxWorker int) WorkerPool {
	wp := &workerPool{
		maxWorker:   maxWorker,
		queuedTaskC: make(chan func()),
		QuitChan:    make(chan bool),
	}

	return wp
}

func (wp *workerPool) Run() {
	wp.run()
}

func (wp *workerPool) AddTask(task func()) {
	wp.queuedTaskC <- task
}

func (wp *workerPool) GetTotalQueuedTask() int {
	return len(wp.queuedTaskC)
}

func (wp *workerPool) run() {
	for i := 0; i < wp.maxWorker; i++ {
		wID := i + 1
		log.Printf("[WorkerPool] Worker %d has been spawned", wID)

		go func(workerID int) {
			for task := range wp.queuedTaskC {
				log.Printf("[WorkerPool] Worker %d start processing task", wID)
				task()
				log.Printf("[WorkerPool] Worker %d finish processing task", wID)
			}
		}(wID)
	}
}

// Stop tells the worker to stop listening for work requests.
//
// Note that the worker will only stop *after* it has finished its work.
func (w *workerPool) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}
