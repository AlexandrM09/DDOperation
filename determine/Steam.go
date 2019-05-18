package determine

import (
	"errors"
	"fmt"
	"time"
	"bufio"
    "encoding/csv"
   
    "log"
    "io"
    
    "os"
)
type SteamRND struct{}
func (St *SteamRND) Read(ScapeDataCh chan ScapeDataD, DoneCh chan struct{}) {
	fmt.Println("RND")
    return 
}

// Steam for csv example files
type SteamCsv struct{}
func (St *SteamCsv) Read(ScapeDataCh chan ScapeDataD, DoneCh chan struct{}) {
	defer func(){DoneCh <- struct{}{}}()
	//nothing
	v1 := [20]float32{0, 0, 100, 90, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	csvFile, _ := os.Open("source/0930small.csv")
	reader := csv.NewReader(bufio.NewReader(csvFile))
	for {
		n:=0
        line, error := reader.Read()
        if error == io.EOF {
			
            break
        } else if error != nil {
			
            log.Fatal(error)
		}
		if n==0 {
			n=1
			err:=parseheader(line,&scapeHeader{})
			if err!=nil{
				
            log.Fatal(err)
			}
		}
      /*  people = append(people, Person{
            Firstname: line[0],
            Lastname:  line[1],
            Address: &Address{
                City:  line[2],
                State: line[3],
            },
		})
		*/
    }
	ScapeData := ScapeDataD{Time: time.Now(), Values: v1}
	for i := 0; i < 30; i++ {
		
		fmt.Println("Sending ScapeData ", fmt.Sprint(i))
		ScapeDataCh <- ScapeData
		ScapeData.Time = ScapeData.Time.Add(time.Second)
		//<-time.After(10 * time.Millisecond)
	}
	DoneCh <- struct{}{}
	return
}
// 

// parse csv header

func parseheader(record []string,sH *scapeHeader) error{ 
h:=[20]string{"","","","","","","","","","","","","","","","","","","",""}	
for i:=2;i<20;i++{
sH[2]=findColumn(h[i],record)
if sH[2]==0 {return errors.New("not found column")}
}
return nil
}

func findColumn(sourceString string,record []string) int {
	var i int
	for i=4;i<len(record);i=i+2{
	if record[i]==sourceString{return i}	
	}
return i
}

type scapeHeader [20]int