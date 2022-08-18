package notification

type Notification[T any] struct {
	name  string
	value T
}
