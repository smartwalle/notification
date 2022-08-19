package notification

import (
	"github.com/smartwalle/queue/block"
	"sync"
	"unsafe"
)

var shared *Center[interface{}]
var once sync.Once

func Default() *Center[interface{}] {
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

type Center[T any] struct {
	opts   *options
	mu     *sync.Mutex
	queue  block.Queue[Notification[T]]
	chains map[string]HandlerChain[T]
}

func New[T any](opts ...Option) *Center[T] {
	var center = &Center[T]{}
	center.opts = &options{}
	center.mu = &sync.Mutex{}
	center.queue = block.New[Notification[T]]()
	center.chains = make(map[string]HandlerChain[T])

	for _, opt := range opts {
		if opt != nil {
			opt(center.opts)
		}
	}

	go center.run()
	return center
}

func (this *Center[T]) Handle(name string, handler Handler[T]) {
	if len(name) == 0 || handler == nil {
		return
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	var chain = this.chains[name]
	var handlerPoint = *(*int)(unsafe.Pointer(&handler))
	for _, current := range chain {
		if *(*int)(unsafe.Pointer(&current)) == handlerPoint {
			return
		}
	}

	chain = append(chain, handler)
	this.chains[name] = chain
}

func (this *Center[T]) Remove(name string) {
	if len(name) == 0 {
		return
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	delete(this.chains, name)
}

func (this *Center[T]) RemoveHandler(name string, handler Handler[T]) {
	if len(name) == 0 || handler == nil {
		return
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	var chain, ok = this.chains[name]
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
		delete(this.chains, name)
	}
}

func (this *Center[T]) RemoveAll() {
	this.mu.Lock()
	defer this.mu.Unlock()

	for name := range this.chains {
		delete(this.chains, name)
	}
}

func (this *Center[T]) Dispatch(name string, value T) bool {
	if len(name) == 0 {
		return false
	}

	var notification = Notification[T]{
		name:  name,
		value: value,
	}
	if this.opts.waiter != nil {
		this.opts.waiter.Add(1)
	}
	var ok = this.queue.Enqueue(notification)

	if ok == false && this.opts.waiter != nil {
		this.opts.waiter.Done()
	}
	return ok
}

func (this *Center[T]) Close() {
	this.queue.Close()
}

func (this *Center[T]) run() {
	var notifications []Notification[T]

	for {
		notifications = notifications[0:0]
		var ok = this.queue.Dequeue(&notifications)

		for _, notification := range notifications {
			this.mu.Lock()
			var chain = this.chains[notification.name]
			this.mu.Unlock()

			for _, handler := range chain {
				handler(notification.name, notification.value)
			}

			if this.opts.waiter != nil {
				this.opts.waiter.Done()
			}
		}

		if ok == false {
			return
		}
	}
}
