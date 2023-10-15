package fucker

const (
	DefaultFuckerSize = 1e9
)

var (
	defaultFuckerPool, _ = NewPool(DefaultFuckerSize)
)

func Submit(task func()) error {
	return defaultFuckerPool.Submit(task)
}

func Running() int {
	return defaultFuckerPool.Running()
}

func Free() int {
	return defaultFuckerPool.Free()
}

func Cap() int {
	return defaultFuckerPool.Cap()
}
