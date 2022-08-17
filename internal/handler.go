package internal

type Handler[T any] func(name string, value T)

type HandlerChain[T any] struct {
	Notification chan *Notification[T]
	HandlerList  []Handler[T]
}

func NewHandlerChain[T any]() *HandlerChain[T] {
	var chain = &HandlerChain[T]{}
	chain.Notification = make(chan *Notification[T], 32)
	go chain.run()
	return chain
}

func (this *HandlerChain[T]) run() {
	for {
		select {
		case notification, ok := <-this.Notification:
			if ok == false {
				return
			}

			for _, handler := range this.HandlerList {
				handler(notification.name, notification.value)
			}
		}
	}
}
