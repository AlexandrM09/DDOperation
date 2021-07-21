package eventbussimple

import (
	_ "fmt"
	"sync"
	_ "time"

	//	steam "github.com/AlexandrM09/DDOperation/pkg/steamd"
	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	"github.com/sirupsen/logrus"
)

type (
	event = struct {
		value chan interface{}
		mu    *sync.RWMutex
	}
	Eventbus struct {
		Log       *logrus.Logger
		count     int
		ScapeData map[string]event
		Save      map[string]event
		Determine map[string]event
		Summary   map[string]event
		//ScapeCh
		mu *sync.RWMutex
	}
)

//AddWell adding 4 chanels
func (Ev *Eventbus) AddWell(id string, chanelbufcount int) {
	Ev.count = chanelbufcount
	if _, ok := Ev.ScapeData[id]; ok {
		return
	}
	Ev.ScapeData[id] = event{
		value: make(chan interface{}, chanelbufcount),
		mu:    &sync.RWMutex{},
	}
	Ev.Save[id] = event{
		value: make(chan interface{}, chanelbufcount),
		mu:    &sync.RWMutex{},
	}
	Ev.Determine[id] = event{
		value: make(chan interface{}, chanelbufcount),
		mu:    &sync.RWMutex{},
	}
	Ev.Summary[id] = event{
		value: make(chan interface{}, chanelbufcount),
		mu:    &sync.RWMutex{},
	}
}

//Send
func (Ev *Eventbus) Send(evt string, id string, val interface{}) bool {
	//defer Ev.mu.Unlock()
	//Ev.mu.Lock()
	switch evt {
	case "ScapeData":
		{
			ch, ok := Ev.ScapeData[id]
			if ok {
				//defer ch.mu.Unlock()
				//ch.mu.Lock()
				//if len(ch.value) < Ev.count {

				ch.value <- val
				d, ok3 := val.(*nt.ScapeDataD)
				if ok3 {
					Ev.Log.Debugf("Ev:Send event ScapeData id=%s,countbuf=%d,count=%d,t=%s,v=%.3f ", d.Id, len(ch.value), d.Count, d.Time.Format("2006-01-02 15:04:05"), d.Values[3])
					//	}

				}
				return true
				//ch.value <- val
			}
		}
	case "Save":
		{
			ch, ok := Ev.Save[id]
			if ok {
				defer ch.mu.Unlock()
				ch.mu.Lock()
				//vlocal := val.(iBus).Get()
				//fmt.Printf("Save cal:%v, len(chan)=%d\n", ch, len(ch.value))
				if len(ch.value) < Ev.count {
					ch.value <- val
					return true
				}

			}

		}
	case "Determine":
		{
			ch, ok := Ev.Determine[id]
			if ok {
				defer ch.mu.Unlock()
				ch.mu.Lock()
				if len(ch.value) < Ev.count {
					ch.value <- val
					return true
				}
			}
		}
	case "Summary":
		{
			ch, ok := Ev.Summary[id]
			if ok {
				defer ch.mu.Unlock()
				ch.mu.Lock()
				if len(ch.value) < Ev.count {
					ch.value <- val
					return true
				}
			}
		}
	}
	return false
}

//Receive
func (Ev *Eventbus) Receive(evt string, id string) (interface{}, bool) {

	var ch event
	//defer Ev.mu.Unlock()
	//Ev.mu.Lock()
	ok := false

	switch evt {
	case "ScapeData":
		{
			ch, ok = Ev.ScapeData[id]
			Ev.Log.Debugf("Ev:Scapedate")
		}
	case "Save":
		{
			ch, ok = Ev.Save[id]
			Ev.Log.Debugf("Ev:Save")
		}
	case "Determine":
		{
			ch, ok = Ev.Determine[id]
			Ev.Log.Debugf("Ev:Determine")
		}
	case "Summary":
		{
			ch, ok = Ev.Summary[id]
			Ev.Log.Debugf("Ev:Summary")
		}
	}
	//Ev.mu.Unlock()
	if ok {
		//defer ch.mu.Unlock()
		//ch.mu.Lock()
		select {

		case val := <-ch.value:
			{

				//w := val.(iBus).Get()
				//println("iBus get:", w.(*Example).val)
				//Ev.Log.Debugf("Ev:<-ch.value,val=%v", val)
				d, ok4 := val.(*nt.ScapeDataD)
				if ok4 {
					Ev.Log.Debugf("Ev:Receive event ScapeData id=%s,count buf=%d,count=%d,t=%s,v=%.3f", d.Id, len(ch.value), d.Count, d.Time.Format("2006-01-02 15:04:05"), d.Values[3])
				}
				return val, true
			}

		default:
			return nil, false
		}
	}

	return nil, false
}

//Neweventbussimple constructor
func Neweventbussimple(count int, l *logrus.Logger) *Eventbus {
	return &Eventbus{
		ScapeData: make(map[string]event, count),
		Save:      make(map[string]event, count),
		Determine: make(map[string]event, count),
		Summary:   make(map[string]event, count),
		mu:        &sync.RWMutex{},
		Log:       l,
	}
}

type iBus interface {
	Get() interface{}
	Set(d interface{})
}
