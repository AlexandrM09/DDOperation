package StoreMap

import (
	_ "errors"
	_ "fmt"
	"sync"
	_ "time"

	//	steam "github.com/AlexandrM09/DDOperation/pkg/steamd"

	"github.com/sirupsen/logrus"
)

const (
	chanelbufcount = 100
	wellcount      = 10
)

type (
	// Topic = struct {
	// 	mu       *sync.RWMutex
	// 	Value    map[string]chan interface{}
	// 	flagFull bool
	// }
	Topic   = map[string]chan interface{}
	Brocker struct {
		Log    *logrus.Logger
		count  int
		Topics Topic
		//ScapeCh
		mu    *sync.RWMutex
		state bool
	}
)

func New(log *logrus.Logger, topic []string) *Brocker {

	b := Brocker{Log: log,
		count:  chanelbufcount,
		Topics: make(Topic, len(topic)),
		mu:     &sync.RWMutex{},
		state:  true,
	}
	for _, v := range topic {
		b.Topics[v] = make(chan interface{}, chanelbufcount)
	}

	return &b
}
func (b *Brocker) Close() {
	b.mu.Lock()
	b.state = false
	b.mu.Unlock()
	for _, v := range b.Topics {
		close(v)
	}
}
func (b *Brocker) Send(topic string, val interface{}) bool {
	b.mu.Lock()
	if !b.state {
		b.mu.Unlock()
		return false
	}
	b.mu.Unlock()
	var v chan interface{}
	var ok bool
	v, ok = b.Topics[topic]
	if !ok {
		return false
	}
	if len(v) == cap(v) {
		b.Log.Debugf("topicid=%s is full \n", topic)
		return false
	}
	v <- val
	// fmt.Printf("broker Send return=%s,id =%s,data=%v\n", partition, id, v)
	return true
}
func (b *Brocker) Receive(topic string) interface{} {
	var v chan interface{}
	var ok bool
	b.mu.Lock()
	v, ok = b.Topics[topic]
	b.mu.Unlock()
	if !ok {
		return nil
	}
	var val interface{}
	select {
	case val, ok = <-v:
	default:
		return nil
	}
	if !ok {

		return nil
	}

	return val
}
