package notification_test

import (
	"github.com/smartwalle/notification"
	"sync"
	"testing"
)

func TestCenter_Dispatch(t *testing.T) {
	var w = &sync.WaitGroup{}
	notification.Default().Handle("n1", func(name string, value interface{}) {
		t.Log("new message", name, value)
		w.Done()
	})
	w.Add(1)
	notification.Default().Dispatch("n1", "haha")
	w.Wait()
}

func BenchmarkCenter_Dispatch(b *testing.B) {
	var w = &sync.WaitGroup{}

	var nCenter = notification.New[int](notification.WithWaiter(w))
	nCenter.Handle("b1", func(name string, value int) {
	})
	for i := 0; i < b.N; i++ {
		nCenter.Dispatch("b1", i)
	}
	w.Wait()

	nCenter.Close()
}
