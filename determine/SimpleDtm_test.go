package determine

import (
	"io/ioutil"
	_"strconv"
	"strings"

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
	Cfg := ConfigDt{}
	errf = LoadConfigYaml("../config.yaml", &Cfg)
	if errf != nil {
		t.Fatal("not load config file")
	}
	sr := DrillDataType{
		Log: CLog(),
		Cfg: &Cfg,
	}

	tm := NewDetermine(&sr, &SteamCsv{FilePath: "../source/source1.zip"})
	_ = tm.Start(29)
	err := tm.Wait()
	if err != nil {
		t.Errorf("error:time limit exceeded")
	}
	data := tm.GetOperationList()
	for i := 0; i < len(data); i++ {
		fmt.Fprintf(file, "%s | %s |%s \r\n", data[i].StartData.Time.Format("2006-01-02 15:04:05"),
			data[i].StopData.Time.Format("15:04:05"),
			data[i].Operaton)
	}
	FileRes, err2 := ioutil.ReadFile("./result1.txt")
	if err2 != nil {
		t.Errorf("file not find result1.txt ")
	}
	resLines := strings.Split(string(FileRes), "\r\n")
	if len(resLines) == 0 {
		t.Errorf("result1.txt is empty")
	}
	sres := ""
	var n int
	data2 := tm.GetSummarysheet()
	fmt.Printf("Start print Summarysheet len=%v \n", len(data))
	for i := 0; i < len(data2); i++ {
		//fmt.Printf("%s | %s |%s \r\n", data2[i].Sheet.StartData.Time.Format("2006-01-02 15:04:05"),
		//	data2[i].Sheet.StopData.Time.Format("15:04:05"),
		//	data2[i].Sheet.Operaton)
		sres = fmt.Sprintf("%s | %s |%s %s ", data2[i].Sheet.StartData.Time.Format("2006-01-02 15:04:05"),
			data2[i].Sheet.StopData.Time.Format("15:04:05"),
			data2[i].Sheet.Operaton,data2[i].Sheet.Params)
			fmt.Println(sres)
		if (!(sres == resLines[n])) && (n > 0) {
			//t.Errorf("string not equale result1 ")
			//fmt.Println("n=", strconv.Itoa(int(n)))
			//fmt.Println("str=", sres)
			//fmt.Println("r=", resLines[n])
		}
		n = n + 1
		d3 := data2[i].Details
		//fmt.Printf("len Details =%v \n",len(d3))
		for j := 0; j < len(d3); j++ {
			//fmt.Printf("____ %s | %s |%s \r\n", d3[j].StartData.Time.Format("15:04:05"),
			//	d3[j].StopData.Time.Format("15:04:05"),
			//	d3[j].Operaton)
			sres = fmt.Sprintf("____ %s | %s |%s ", d3[j].StartData.Time.Format("15:04:05"),
				d3[j].StopData.Time.Format("15:04:05"),
				d3[j].Operaton)
				fmt.Println(sres)
			if !(sres == resLines[n]) {
				t.Errorf("string not equale result1 ")
				//fmt.Println("n=", strconv.Itoa(int(n)))
				//fmt.Println("str=", sres)
				//fmt.Println("r=", resLines[n])
			}
			n = n + 1
		}
	}
}

//very simple determine test
func TestSimpleDtm(t *testing.T) {
	fmt.Println("Start test")
	Cfg := ConfigDt{}
	errf := LoadConfig("../config.json", &Cfg)
	if errf != nil {
		t.Fatal("not load config file")
	}

	sr := DrillDataType{
		Log: CLog(),
		Cfg: &Cfg,
	}

	tm := NewDetermine(&sr, &SteamSmpl{})
	_ = tm.Start(29)
	err := tm.Wait()
	if err != nil {
		t.Errorf("error:time limit exceeded")
	}
	data := tm.GetOperationList()
	fmt.Println("count operation ", len(data))
	fmt.Println("Start printing OperationList")

	for i := 0; i < len(data); i++ {
		fmt.Printf("%s | %s |%s \n", data[i].StartData.Time.Format("2006-01-02 15:04:05"),
			data[i].StopData.Time.Format("15:04:05"),
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
		n = int64(dd.StopData.Time.Sub(dd.StartData.Time) / time.Second)
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
return log

	
}
