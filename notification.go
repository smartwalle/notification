package notification

type Notification[T any] struct {
	value T
	name  string
}
