package notification

type Handler[T any] func(name string, value T)

type HandlerChain[T any] []Handler[T]
