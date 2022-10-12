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

const countdata = 10000

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

	store := New(logrus.New())
	// evnt.AddWell("well2", 4)

	send := func(id string) {
		for j := 0; j < countdata; j++ {
			fmt.Printf("test Send j=%v\n", j)
			for i := 0; i < 3; i++ {
				store.Send("Sensors data", id, &artest[i])
				store.Send("Sensors data save", id, &artest[i])
				store.Send("Determine save", id, &artest[i])
				store.Send("Determine", id, &artest[i])
				store.Send("Summary", id, &artest[i])
			}
		}
	}
	go send("well1")
	go send("well2")
	time.Sleep(500 * time.Millisecond)
	readTopic := func(sub, id string, count *int) {

		for i := 0; i < 3; i++ {
			val := store.Receive(sub, id)
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
				fmt.Printf("Sub1:%s,Id:%s,val:%s\n", sub, id, w.val)
				if !(w.val == artest[i].val) {
					t.Errorf("not equivalent string nead:%s,reading:%s", artest[i].val, w.val)
				}
			}
		}
	}
	read := func(id string, count *int) {
		for j2 := 0; j2 < countdata; j2++ {
			fmt.Printf("test read j2=%v\n", j2)
			readTopic("Sensors data", id, count)
			readTopic("Sensors data save", id, count)
			readTopic("Determine save", id, count)
			readTopic("Determine", id, count)
			readTopic("Summary", id, count)
		}
	}
	var count int
	read("well1", &count)
	read("well2", &count)

}
