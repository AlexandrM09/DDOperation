package determine

import (
	//"sync"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// simle steam
type SteamSmpl struct{}

func (St *SteamSmpl) Read(ScapeDataCh chan ScapeDataD, DoneCh chan struct{}) {
	//nothing
	v1 := [20]float32{0, 0, 100, 90, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	//flow data
	v2 := v1
	v2[4] = 50
	//drill data
	v3 := v2
	v3[3] = 100
	ScapeData := ScapeDataD{Time: time.Now(), Values: v1}
	for i := 0; i < 30; i++ {
		if i > 9 {
			ScapeData.Values = v2
		}
		if i > 19 {
			ScapeData.Values = v3
		}
		fmt.Println("Sending ScapeData ", fmt.Sprint(i))
		ScapeDataCh <- ScapeData
		ScapeData.Time = ScapeData.Time.Add(time.Second)
		//<-time.After(10 * time.Millisecond)
	}
	DoneCh <- struct{}{}
	return
}

//very simple determine test
func TestSimpleDtm(t *testing.T) {
	fmt.Println("Start test")
	sr := DrillDataType{OperationList: make([]OperationOne, 0),
		SteamCh:         make(chan OperationOne),
		ScapeDataCh:     make(chan ScapeDataD),
		ErrCh:           make(chan error, 2),
		DoneCh:          make(chan struct{}),
		DoneScapeCh:     make(chan struct{}),
		ActiveOperation: -1,
		Operationtype:   DrillOperationConst,
		Log:             CLog(),
		//mu:&sync.RWMutex{},
	}

	tm := NewDetermine(&sr, &SteamSmpl{})
	_ = tm.Start(15)
	err:=tm.Wait()
if err != nil{
	t.Errorf("error:time limit exceeded")
} 
	//<-time.After(2000 * time.Millisecond)
	fmt.Println("count operation ", len(tm.Data.OperationList))
	fmt.Println("Start printing OperationList")

	for i := 0; i < len(tm.Data.OperationList); i++ {
		fmt.Printf("%s | %s |%s \n", tm.Data.OperationList[i].startData.Time.Format("2006-01-02 15:04:05"),
			tm.Data.OperationList[i].stopData.Time.Format("15:04:05"),
			tm.Data.OperationList[i].Operaton)
	}
	if !(len(tm.Data.OperationList) == 3) {
		t.Errorf("the number of operations does not match")
	}
	neadres := [3]string{"ПЗР", "Промывка", "Бурение"}
	var dd OperationOne
	var n int64
	 
	for i := 0; i <  len(tm.Data.OperationList); i++ {
		dd = tm.Data.OperationList[i]
		if !(neadres[i] == dd.Operaton) {
			t.Errorf("incorrect operation definition")
		}
		n = int64(dd.stopData.Time.Sub(dd.startData.Time) / time.Second)
		if !(n == 9) {
			t.Errorf("incorrect time duration %v", n)
		}
	}

}

func CLog() *logrus.Logger {
	var log = logrus.New()
	log.WithFields(logrus.Fields{
		//"mode":   "[access_log]",
		"logger": "LOGRUS",
	})
	log.SetFormatter(&logrus.JSONFormatter{})
	log.Out = os.Stdout
	file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
	/**/
	return log

}
