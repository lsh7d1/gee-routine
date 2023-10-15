package fucker

import (
	"errors"
	"fucker/syncx"
	"sync"
	"sync/atomic"
)

// Pool 用于接收客户端任务，通过回收goroutine限制 worker 总数
type Pool struct {
	capacity    int32
	state       int32
	running     int32
	waiting     int32
	workers     *workerQueue
	lock        sync.Locker
	cond        *sync.Cond
	workerCache sync.Pool
}

func NewPool(size int) (*Pool, error) {
	if size <= 0 {
		size = -1 // 无限容量
	}

	p := &Pool{
		capacity: int32(size),
		lock:     syncx.NewSpinLock(),
	}

	p.workerCache.New = func() any {
		return &goWorker{
			pool: p,
			task: make(chan func(), 1),
		}
	}

	p.workers = newWorkerQueue(size)
	p.cond = sync.NewCond(p.lock)

	return p, nil
}

func (p *Pool) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

func (p *Pool) Submit(task func()) error {
	if w := p.acquireWorker(); w != nil {
		w.inputFunc(task)
		return nil
	}
	return errors.New("submit task failed")
}

func (p *Pool) Free() int {
	c := p.Cap()
	if c < 0 {
		return -1
	}
	return c - p.Running()
}

func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *Pool) addRunning(delta int) {
	atomic.AddInt32(&p.running, int32(delta))
}

func (p *Pool) addWaiting(delta int) {
	atomic.AddInt32(&p.waiting, int32(delta))
}

func (p *Pool) acquireWorker() (w *goWorker) {
	spawnWorker := func() {
		w = p.workerCache.Get().(*goWorker)
		w.run()
	}

	p.lock.Lock()
	w = p.workers.detach()
	if w != nil {
		// workerQueue 不为空
		p.lock.Unlock()
	} else if capacity := p.Cap(); capacity == -1 || capacity > p.Running() {
		// 池子还有容量，那么直接生产一个 worker
		p.lock.Unlock()
		spawnWorker()
	} else {
	retry:
		p.addWaiting(1)
		p.cond.Wait() // 阻塞并等待可用的 worker
		p.addWaiting(-1)

		if w = p.workers.detach(); w == nil {
			if p.Free() > 0 {
				p.lock.Unlock()
				spawnWorker()
				return
			}
			goto retry
		}
		p.lock.Unlock()
	}
	return
}

// revertWorker 用于归还 worker
func (p *Pool) revertWorker(worker *goWorker) bool {
	p.lock.Lock()
	if err := p.workers.insert(worker); err != nil {
		p.lock.Unlock()
		return false
	}
	p.cond.Signal()
	p.lock.Unlock()

	return true
}
