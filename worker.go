package fucker

import (
	"fmt"
	"runtime/debug"
)

type goWorker struct {
	pool *Pool
	task chan func()
}

func (w *goWorker) run() {
	w.pool.addRunning(1)
	go func() {
		defer func() {
			w.pool.addRunning(-1)
			w.pool.workerCache.Put(w)
			if p := recover(); p != nil {
				fmt.Printf("worker exit from panic: %v\n%s\n", p, debug.Stack())
			}
			w.pool.cond.Signal()
		}()

		for f := range w.task {
			if f == nil {
				return
			}
			f()
			if ok := w.pool.revertWorker(w); !ok {
				return
			}
		}
	}()
}

func (w *goWorker) finish() {
	w.task <- nil
}

func (w *goWorker) inputFunc(fn func()) {
	w.task <- fn
}
