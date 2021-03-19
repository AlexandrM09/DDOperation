package eventbussimple

import (
	"fmt"
	"testing"
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

	evnt := Neweventbussimple(3)
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
			val := evnt.Receive(sub, id)
			if val == nil {
				t.Errorf("not reading sendig date ")
			}
			//var ex example

			w, ok1 := val.(*Example)
			//p:=val.(type).k
			if !ok1 {
				t.Errorf("not unmarshal date ")
			}
			if ok1 {
				fmt.Printf("Sub:%s,Id:%s,val:%s\n", sub, id, w.val)

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
func TestSteamoverflow(t *testing.T) {
	/**	artest := [3]Example{
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
	**/
	evnt := Neweventbussimple(10)
	countBuf := 40
	evnt.AddWell("well1", countBuf)
	send := func(id string, countbuf int) {
		for j := 0; j < countBuf*2; j++ {
			s := id
			for i := 0; i < 3; i++ {
				t := Example{id: s,
					val: fmt.Sprintf("color:%d,col:%d", j, i)}
				//artest[i].val =
				evnt.Send("ScapeData", id, &t)
				evnt.Send("Save", id, &t)
				evnt.Send("Determine", id, &t)
				evnt.Send("Summary", id, &t)
			}
		}
	}

	send("well1", countBuf)

	//var val interface{}
	//var w Example
	read := func(sub, id string, count *int) {
		for i := 0; i < 3; i++ {
			val := evnt.Receive(sub, id)
			if val == nil {
				fmt.Printf("nil data read\n")
				return
				//t.Errorf("not reading sendig date ")
			}
			//var ex example

			w, ok1 := val.(*Example)
			//p:=val.(type).k
			if !ok1 {
				t.Errorf("not unmarshal date ")
			}
			if ok1 {
				fmt.Printf("Read:%s,Id:%s,val:%s\n", sub, id, w.val)
				*count++
				//if !(w.val == artest[i].val) {
				//	t.Errorf("not equivalent string nead:%s,reading:%s", artest[i].val, w.val)
				//}
			}
		}
	}
	var countReding int
	for j := 0; j < countBuf; j++ {

		read("ScapeData", "well1", &countReding)
		read("Save", "well1", &countReding)
		read("Determine", "well1", &countReding)
		read("Summary", "well1", &countReding)

	}
	countReding = countReding / 4
	if countBuf != countReding {
		t.Errorf("not equivalent count data need:%d,real:%d", countBuf, countReding)
	}

	fmt.Printf("Count sending data:%d,reading data:%d\n ", countBuf*2, countReding)

}
