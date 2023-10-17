package notification

import (
	"github.com/smartwalle/queue/block"
	"sync"
	"unsafe"
)

var shared Center[interface{}]
var once sync.Once

func Default() Center[interface{}] {
	once.Do(func() {
		shared = New[interface{}]()
	})
	return shared
}

type Option func(opts *options)

type options struct {
	waiter Waiter
}

func WithWaiter(waiter Waiter) Option {
	return func(opts *options) {
		opts.waiter = waiter
	}
}

type Center[T any] interface {
	Handle(name string, handler Handler[T])

	Remove(name string)

	RemoveHandler(name string, handler Handler[T])

	RemoveAll()

	Dispatch(name string, value T) bool

	Close()
}

type center[T any] struct {
	opts   *options
	mu     sync.RWMutex
	queue  block.Queue[Notification[T]]
	chains map[string]HandlerChain[T]
}

func New[T any](opts ...Option) Center[T] {
	var nCenter = &center[T]{}
	nCenter.opts = &options{}
	nCenter.queue = block.New[Notification[T]]()
	nCenter.chains = make(map[string]HandlerChain[T])

	for _, opt := range opts {
		if opt != nil {
			opt(nCenter.opts)
		}
	}

	go nCenter.run()
	return nCenter
}

func (c *center[T]) Handle(name string, handler Handler[T]) {
	if len(name) == 0 || handler == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var chain = c.chains[name]
	var pointer = *(*int)(unsafe.Pointer(&handler))
	for _, current := range chain {
		if *(*int)(unsafe.Pointer(&current)) == pointer {
			return
		}
	}

	chain = append(chain, handler)
	c.chains[name] = chain
}

func (c *center[T]) Remove(name string) {
	if len(name) == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.chains, name)
}

func (c *center[T]) RemoveHandler(name string, handler Handler[T]) {
	if len(name) == 0 || handler == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var chain, ok = c.chains[name]
	if ok == false {
		return
	}

	var found = -1
	var handlerPoint = *(*int)(unsafe.Pointer(&handler))
	for i, current := range chain {
		if *(*int)(unsafe.Pointer(&current)) == handlerPoint {
			found = i
		}
	}

	if found >= 0 {
		chain = append(chain[:found], chain[found+1:]...)
	}

	if len(chain) == 0 {
		delete(c.chains, name)
	}
}

func (c *center[T]) RemoveAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for name := range c.chains {
		delete(c.chains, name)
	}
}

func (c *center[T]) Dispatch(name string, value T) bool {
	if len(name) == 0 {
		return false
	}

	var notification = Notification[T]{
		name:  name,
		value: value,
	}
	if c.opts.waiter != nil {
		c.opts.waiter.Add(1)
	}
	var ok = c.queue.Enqueue(notification)

	if !ok && c.opts.waiter != nil {
		c.opts.waiter.Done()
	}
	return ok
}

func (c *center[T]) Close() {
	c.queue.Close()
}

func (c *center[T]) run() {
	var notifications []Notification[T]

	for {
		notifications = notifications[0:0]
		var ok = c.queue.Dequeue(&notifications)

		for _, notification := range notifications {
			c.mu.RLock()
			var chain = c.chains[notification.name]
			c.mu.RUnlock()

			for _, handler := range chain {
				handler(notification.name, notification.value)
			}

			if c.opts.waiter != nil {
				c.opts.waiter.Done()
			}
		}

		if !ok {
			return
		}
	}
}
