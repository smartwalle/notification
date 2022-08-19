package notification

type Waiter interface {
	Add(delta int)

	Done()

	Wait()
}
