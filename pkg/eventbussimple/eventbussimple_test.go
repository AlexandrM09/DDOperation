package eventbussimple

import (
	"fmt"
	"sync"
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

	evnt := Neweventbussimple(3, logrus.New())
	evnt.AddWell("1", 4)
	evnt.AddWell("well2", 4)

	send := func(id string) {
		for i := 0; i < 3; i++ {
			evnt.Send("ScapeData", id, &artest[i])
			evnt.Send("Save", id, &artest[i])
			evnt.Send("Determine", id, &artest[i])
			evnt.Send("Summary", id, &artest[i])
		}
	}
	send("1")
	send("well2")

	//var val interface{}
	//var w Example
	read := func(sub, id string) {
		for i := 0; i < 3; i++ {
			val, ok := evnt.Receive(sub, id)
			if !ok {
				t.Errorf("not reading sendig date ")
			}
			//var ex example

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
	read("ScapeData", "1")
	read("Save", "1")
	read("Determine", "1")
	read("Summary", "1")
	read("ScapeData", "well2")
	read("Save", "well2")
	read("Determine", "well2")
	read("Summary", "well2")
}

//Test data sending*2=reading
func TestFullFlow(t *testing.T) {

	evnt := Neweventbussimple(10, logrus.New())
	countBuf := 40
	evnt.AddWell("well1", countBuf)

	wg := &sync.WaitGroup{}
	send := func(id string, countbuf int, countSend *int, Wg *sync.WaitGroup) {
		wg.Add(1)
		for j := 0; j < 70; j++ { //countBuf*2
			s := id
			vs := fmt.Sprintf("color:%d", j+1)
			//	for i := 0; i < 3; i++ {
			t := Example{id: s,
				val: vs}
			//fmt.Printf("in send j=%d,t=%v\n", j, t)
			//artest[i].val =
			for !evnt.Send("ScapeData", id, &t) {

			}
			*countSend++
			fmt.Printf("Send Id:%s,count:%d,val:%v\n", id, *countSend, t)
			//evnt.Send("Save", id, &t)
			//evnt.Send("Determine", id, &t)
			//evnt.Send("Summary", id, &t)
			//}
		}
		time.Sleep(500 * time.Millisecond)
		wg.Done()
	}
	CountSend := 0
	fmt.Printf("Start send")
	go send("well1", countBuf, &CountSend, wg)

	//var val interface{}
	//var w Example
	read1 := func(sub, id string, count *int) bool {
		//for i := 0; i < 3; i++ {
		val, ok := evnt.Receive(sub, id)
		if !ok {
			//fmt.Printf("nil data read\n")
			return false
			//t.Errorf("not reading sendig date ")
		}
		//var ex example

		w, ok1 := val.(*Example)
		//p:=val.(type).k
		if !ok1 {
			t.Errorf("not unmarshal date ")
		}
		//fmt.Printf("ScapeData %s",w.val)
		if ok1 {
			fmt.Printf("Read:%s,Id:%s,value:%v\n", sub, id, w)
			*count++
			//if !(w.val == artest[i].val) {
			//	t.Errorf("not equivalent string nead:%s,reading:%s", artest[i].val, w.val)
			//}
		}
		//}
		return true
	}

	var countReding int
	j := 0
	fmt.Printf("Start read")
	go func() {
		for {
			if read1("ScapeData", "well1", &countReding) {

				j++
				fmt.Printf("read j= %d\n", j)
			}
			//read("Save", "well1", &countReding)
			//read("Determine", "well1", &countReding)
			//read("Summary", "well1", &countReding)

		}
	}()
	wg.Wait()
	//countReding = countReding
	if CountSend != countReding {
		t.Errorf("not equivalent count data need:%d,real:%d", CountSend, countReding)
	}

	fmt.Printf("Count sending data:%d,reading data:%d\n ", CountSend, countReding)

}
