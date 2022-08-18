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
	var nCenter = notification.New[int]()

	var w = &sync.WaitGroup{}
	nCenter.Handle("b1", func(name string, value int) {
		w.Done()
	})
	for i := 0; i < b.N; i++ {
		w.Add(1)
		nCenter.Dispatch("b1", i)
	}
	w.Wait()
	nCenter.Close()
}
