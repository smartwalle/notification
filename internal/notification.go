package internal

type Notification[T any] struct {
	name  string
	value T
}

func NewNotification[T any](name string, value T) *Notification[T] {
	var notification = &Notification[T]{}
	notification.name = name
	notification.value = value
	return notification
}
