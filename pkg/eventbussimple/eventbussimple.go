package eventbussimple

import (
	_ "fmt"
	"sync"
	_ "time"
	//	steam "github.com/AlexandrM09/DDOperation/pkg/steamd"
)

type (
	event = struct {
		value chan interface{}
	}
	Eventbus struct {
		count     int
		ScapeData map[string]event
		Save      map[string]event
		Determine map[string]event
		Summary   map[string]event
		mu        sync.RWMutex
	}
)

//AddWell adding 4 chanels
func (Ev *Eventbus) AddWell(id string, chanelbufcount int) {
	Ev.count = chanelbufcount
	Ev.ScapeData[id] = event{
		value: make(chan interface{}, chanelbufcount),
	}
	Ev.Save[id] = event{
		value: make(chan interface{}, chanelbufcount),
	}
	Ev.Determine[id] = event{
		value: make(chan interface{}, chanelbufcount),
	}
	Ev.Summary[id] = event{
		value: make(chan interface{}, chanelbufcount),
	}
}

//Send
func (Ev *Eventbus) Send(evt string, id string, val interface{}) {
	defer Ev.mu.Unlock()
	Ev.mu.Lock()
	switch evt {
	case "ScapeData":
		{
			ch, ok := Ev.ScapeData[id]
			if ok {
				if len(ch.value) < Ev.count {
					ch.value <- val
				}
			}
		}
	case "Save":
		{
			ch, ok := Ev.Save[id]
			if ok {
				//vlocal := val.(iBus).Get()
				//fmt.Printf("Save cal:%v, len(chan)=%d\n", ch, len(ch.value))
				if len(ch.value) < Ev.count {
					ch.value <- val
				}

			}
		}
	case "Determine":
		{
			ch, ok := Ev.Determine[id]
			if ok {
				if len(ch.value) < Ev.count {
					ch.value <- val
				}
			}
		}
	case "Summary":
		{
			ch, ok := Ev.Summary[id]
			if ok {
				if len(ch.value) < Ev.count {
					ch.value <- val
				}
			}
		}
	}

}

//Receive
func (Ev *Eventbus) Receive(evt string, id string) interface{} {
	defer Ev.mu.Unlock()
	Ev.mu.Lock()
	var ch event
	var ok bool
	switch evt {
	case "ScapeData":
		{
			ch, ok = Ev.ScapeData[id]

		}
	case "Save":
		{
			ch, ok = Ev.Save[id]

		}
	case "Determine":
		{
			ch, ok = Ev.Determine[id]

		}
	case "Summary":
		{
			ch, ok = Ev.Summary[id]

		}
	}
	if ok {
		select {

		case val := <-ch.value:
			{

				//w := val.(iBus).Get()
				//println("iBus get:", w.(*Example).val)
				return val
			}

		default:

		}
	}

	return nil
}

//Neweventbussimple constructor
func Neweventbussimple(countwell int) *Eventbus {
	return &Eventbus{
		ScapeData: make(map[string]event, countwell),
		Save:      make(map[string]event, countwell),
		Determine: make(map[string]event, countwell),
		Summary:   make(map[string]event, countwell),
		mu:        sync.RWMutex{},
	}
}

type iBus interface {
	Get() interface{}
	Set(d interface{})
}
