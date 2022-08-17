package notification

import (
	"errors"
	"github.com/smartwalle/notification/internal"
	"sync"
	"time"
	"unsafe"
)

type Handler[T any] func(name string, value T)

var (
	ErrTimeout = errors.New("dispatch notification timeout")
)

var shared *Center[interface{}]
var once sync.Once

func Default() *Center[interface{}] {
	once.Do(func() {
		shared = New[interface{}]()
	})
	return shared
}

type Center[T any] struct {
	mu     *sync.Mutex
	chains map[string]*internal.HandlerChain[T]
}

func New[T any]() *Center[T] {
	var center = &Center[T]{}
	center.mu = &sync.Mutex{}
	center.chains = make(map[string]*internal.HandlerChain[T])
	return center
}

func (this *Center[T]) Handle(name string, handler Handler[T]) {
	if len(name) == 0 || handler == nil {
		return
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	var chain, ok = this.chains[name]
	if ok == false {
		chain = internal.NewHandlerChain[T]()
		this.chains[name] = chain
	}

	var handlerPoint = *(*int)(unsafe.Pointer(&handler))
	for _, current := range chain.HandlerList {
		if *(*int)(unsafe.Pointer(&current)) == handlerPoint {
			return
		}
	}

	chain.HandlerList = append(chain.HandlerList, internal.Handler[T](handler))
}

func (this *Center[T]) Remove(name string) {
	if len(name) == 0 {
		return
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	var chain, ok = this.chains[name]
	if ok == false {
		return
	}

	close(chain.Notification)
	chain.HandlerList = nil
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
	for i, current := range chain.HandlerList {
		if *(*int)(unsafe.Pointer(&current)) == handlerPoint {
			found = i
		}
	}

	if found >= 0 {
		chain.HandlerList = append(chain.HandlerList[:found], chain.HandlerList[found+1:]...)
	}

	if len(chain.HandlerList) == 0 {
		close(chain.Notification)
		chain.HandlerList = nil
		delete(this.chains, name)
	}
}

func (this *Center[T]) RemoveAll() {
	this.mu.Lock()
	defer this.mu.Unlock()

	for key, chain := range this.chains {
		close(chain.Notification)
		chain.HandlerList = nil
		delete(this.chains, key)
	}
}

func (this *Center[T]) Dispatch(name string, value T) error {
	if len(name) == 0 {
		return nil
	}
	var notification = internal.NewNotification[T](name, value)

	this.mu.Lock()
	var chain = this.chains[name]
	this.mu.Unlock()

	if chain != nil {
		select {
		case chain.Notification <- notification:
			return nil
		case <-time.After(time.Second * 1):
			return ErrTimeout
		}
	}
	return nil
}
