package fucker

import "errors"

var (
	errQueueIsFull     = errors.New("this queue is full")
	errQueueIsReleased = errors.New("the length of this queue is zero")
)

type workerQueue struct {
	items  []*goWorker
	head   int
	tail   int
	size   int
	isFull bool
}

func newWorkerQueue(size int) *workerQueue {
	return &workerQueue{
		items: make([]*goWorker, size),
		size:  0,
	}
}

// len 返回队列中有多少 goWorker 在工作
func (wq *workerQueue) len() int {
	if wq.size == 0 {
		return 0
	}

	if wq.head == wq.tail {
		if wq.isFull {
			return wq.size
		}
		return 0
	}

	if wq.tail > wq.head {
		return wq.tail - wq.head
	}

	return wq.size - wq.head + wq.tail
}

func (wq *workerQueue) isEmpty() bool {
	return wq.head == wq.tail && !wq.isFull
}

func (wq *workerQueue) insert(w *goWorker) error {
	if wq.size == 0 {
		return errQueueIsReleased
	}

	if wq.isFull {
		return errQueueIsFull
	}
	wq.items[wq.tail] = w
	wq.tail++

	wq.tail %= wq.size
	if wq.tail == wq.head {
		wq.isFull = true
	}

	return nil
}

func (wq *workerQueue) detach() *goWorker {
	if wq.isEmpty() {
		return nil
	}

	w := wq.items[wq.head]
	wq.head++

	wq.head %= wq.size
	wq.isFull = false

	return w
}

func (wq *workerQueue) reset() {
	if wq.isEmpty() {
		return
	}

retry:
	if w := wq.detach(); w != nil {
		w.finish()
		goto retry
	}
	wq.items = wq.items[:0]
	wq.size = 0
	wq.head = 0
	wq.tail = 0
}
