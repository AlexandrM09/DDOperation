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

//test steam for csv files
func TestSteamCsv(t *testing.T) {
	var Scd ScapeDataD
	ScapeDataCh := make(chan ScapeDataD)
	DoneCh := make(chan struct{})
	SteamCsv := &SteamCsv{FilePath: "../source/source.zip"}
	go SteamCsv.Read(ScapeDataCh, DoneCh)
	fmt.Printf("start")
	//data := []byte("Hello Bold!")
	file, err := os.Create("operation.txt")
	if err != nil {
		t.Errorf("Unable to create file")

	}
	defer file.Close()
	s := ""
	for {
		select {
		case <-DoneCh:
			{
				fmt.Printf("finish")
				return
			}
		case Scd = <-ScapeDataCh:
			{
				s = fmt.Sprintf(" %s | %+v \r\n",
					Scd.Time.Format("2006-01-02 15:04:05"),
					Scd.Values)
				_, _ = file.WriteString(s)

			}
		}
	}

}

func TestElementaryDtm(t *testing.T) {
	fmt.Println("Start test")
	file, errf := os.Create("operation.txt")
	if errf != nil {
		t.Errorf("Unable to create file")

	}
	defer file.Close()
	fmt.Println("Load config")
	cfg := ConfigDt{}
	errf = LoadConfig("../config.json", &cfg)
	if errf != nil {
		t.Fatal("not load config file")
	}
	sr := DrillDataType{
		Log: CLog(),
		cfg: &cfg,
	}

	tm := NewDetermine(&sr, &SteamCsv{FilePath: "../source/source.zip"})
	_ = tm.Start(30)
	err := tm.Wait()
	if err != nil {
		t.Errorf("error:time limit exceeded")
	}
	data:=tm.GetOperationList()
	for i := 0; i < len(data); i++ {
		fmt.Fprintf(file, "%s | %s |%s \r\n", data[i].startData.Time.Format("2006-01-02 15:04:05"),
		data[i].stopData.Time.Format("15:04:05"),
		data[i].Operaton)
	}
	
	data=tm.GetSummarysheet()
	for i := 0; i < len(data); i++ {
		fmt.Printf("%s | %s |%s \r\n", data[i].startData.Time.Format("2006-01-02 15:04:05"),
		data[i].stopData.Time.Format("15:04:05"),
		data[i].Operaton)
	}
}

//very simple determine test
func TestSimpleDtm(t *testing.T) {
	fmt.Println("Start test")
	cfg := ConfigDt{}
	errf := LoadConfig("../config.json", &cfg)
	if errf != nil {
		t.Fatal("not load config file")
	}

	sr := DrillDataType{
		Log: CLog(),
		cfg: &cfg,
	}

	tm := NewDetermine(&sr, &SteamSmpl{})
	_ = tm.Start(25)
	err := tm.Wait()
	if err != nil {
		t.Errorf("error:time limit exceeded")
	}
	data:=tm.GetOperationList()
	fmt.Println("count operation ", len(data))
	fmt.Println("Start printing OperationList")

	for i := 0; i < len(data); i++ {
		fmt.Printf("%s | %s |%s \n", data[i].startData.Time.Format("2006-01-02 15:04:05"),
			data[i].stopData.Time.Format("15:04:05"),
			data[i].Operaton)
	}
	if !(len(data) == 3) {
		t.Errorf("the number of operations does not match")
	}
	neadres := [3]string{"Наращивание", "Промывка", "Бурение (слайд)"}
	var dd OperationOne
	var n int64

	for i := 0; i < len(data); i++ {
		dd = data[i]
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
