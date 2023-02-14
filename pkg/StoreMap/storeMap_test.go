package StoreMap

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

type (
	Example struct {
		id  string
		val string
	}
)

const countdata = 10

func TestSteam(t *testing.T) {
	artest := [3]Example{
		{
			id:  "1",
			val: "red"},
		{
			id:  "1",
			val: "green"},
		{
			id:  "1",
			val: "black"},
	}
	topic := []string{"Sensors data",
		"Sensors data save", "Determine save", "Determine", "Summary"}
	store := New(logrus.New(), topic)
	// evnt.AddWell("well2", 4)

	send := func(id string) {
		for j := 0; j < countdata; j++ {
			for i := 0; i < 3; i++ {
				artest[i].id = id

				fmt.Printf("test Send j=%v\n", j)

				for _, v := range topic {
					for !store.Send(v, &artest[i]) {
					}

				}
			}
		}
	}
	go send("well1")
	// go send("well2")
	time.Sleep(500 * time.Millisecond)
	readTopic := func(sub string, count *int) {

		for i := 0; i < 3; i++ {
			val := store.Receive(sub)
			*count++
			fmt.Printf("Count read read=%v\n", *count)
			if val == nil {
				t.Errorf("not reading sendig date ")
			}
			w, ok1 := val.(*Example)
			//p:=val.(type).k
			if !ok1 {
				t.Errorf("not unmarshal date ")
			}
			if ok1 {
				fmt.Printf("Sub1:%s,val:%s\n", sub, w.val)
				if !(w.val == artest[i].val) {
					t.Errorf("not equivalent string nead:%s,reading:%s", artest[i].val, w.val)
				}
			}
		}
	}
	read := func(count *int) {
		for j2 := 0; j2 < countdata; j2++ {
			fmt.Printf("test read j2=%v\n", j2)
			for _, v := range topic {
				readTopic(v, count)
			}

		}
	}
	var count int
	// read(&count)
	read(&count)
	sendingCount := 3 * countdata * len(topic)
	if !(count == sendingCount) {
		t.Errorf("not equivalent count sending and reading data:sending %d,reading:%d", sendingCount, count)
	}
}
