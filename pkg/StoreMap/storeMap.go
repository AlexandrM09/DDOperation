package StoreMap

import (
	_ "errors"
	"fmt"
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
	Topic = struct {
		mu    *sync.RWMutex
		Value map[string]chan interface{}
	}
	Partition = map[string]Topic
	Brocker   struct {
		Log        *logrus.Logger
		count      int
		Partitions Partition
		//ScapeCh
		mu *sync.RWMutex
	}
)

func New(log *logrus.Logger) *Brocker {

	b := Brocker{Log: log,
		count:      chanelbufcount,
		Partitions: make(Partition, 6),
		mu:         &sync.RWMutex{}}
	b.Partitions["Sensors data save"] = Topic{
		Value: make(map[string]chan interface{}, wellcount),
		mu:    &sync.RWMutex{}}
	b.Partitions["Sensors data"] = Topic{
		Value: make(map[string]chan interface{}, wellcount),
		mu:    &sync.RWMutex{}}
	b.Partitions["Determine save"] = Topic{
		Value: make(map[string]chan interface{}, wellcount),
		mu:    &sync.RWMutex{}}
	b.Partitions["Determine"] = Topic{
		Value: make(map[string]chan interface{}, wellcount),
		mu:    &sync.RWMutex{}}
	b.Partitions["Summary"] = Topic{
		Value: make(map[string]chan interface{}, wellcount),
		mu:    &sync.RWMutex{}}

	return &b
}

func (b *Brocker) Send(partition string, id string, val interface{}) bool {
	var t Topic
	var ok bool
	if t, ok = b.Partitions[partition]; !ok {
		return false
	}

	v, ok := t.Value[id]
	if !ok {
		t.mu.Lock()
		t.Value[id] = make(chan interface{}, chanelbufcount)
		v, ok = t.Value[id]
		t.mu.Unlock()
		fmt.Printf("broker Send partition=%s,id =%s,ok=%v\n", partition, id, ok)
	}
	v <- val
	fmt.Printf("broker Send return=%s,id =%s,data=%v\n", partition, id, v)
	return true
}
func (b *Brocker) Receive(partition string, id string) interface{} {
	var t Topic
	var ok bool
	if t, ok = b.Partitions[partition]; !ok {
		return nil
	}

	t.mu.Lock()
	v, ok := t.Value[id]
	t.mu.Unlock()
	if !ok {
		return nil
	}
	val, ok := <-v
	if !ok {
		fmt.Printf("broker Read partition=%s,id =%s is nil\n", partition, id)
		return nil
	}
	fmt.Printf("broker Read partition=%s,id =%s,data = %v\n", partition, id, val)
	return val
}
func (b *Brocker) CloseBrockerChanel() {
	for _, t := range b.Partitions {
		for _, v := range t.Value {
			close(v)
		}
	}
}
