package determine

import (
	"fmt"
	"io/ioutil"
	"os"
	_ "strconv"
	"strings"
	"testing"
	"time"

	nt "github.com/AlexandrM09/DDOperation/pkg/Sharetype"
	"github.com/sirupsen/logrus"
)

// simle steam
type SteamSmpl struct{}

//function return simple fake ScapeDate for TestSimpleDtm
func (St *SteamSmpl) Read(ScapeDataCh chan nt.ScapeDataD, DoneCh chan struct{}) {
	//nothing
	v1 := [20]float32{0, 0, 100, 90, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	//flow data
	v2 := v1
	v2[4] = 50
	//drill data
	v3 := v2
	v3[3] = 100
	ScapeData := nt.ScapeDataD{Time: time.Now(), Values: v1}
	for i := 0; i < 30; i++ {
		if i > 9 {
			ScapeData.Values = v2
		}
		if i > 19 {
			ScapeData.Values = v3
		}
		//fmt.Println("Sending ScapeData ", fmt.Sprint(i))
		ScapeDataCh <- ScapeData
		ScapeData.Time = ScapeData.Time.Add(time.Second)
		//<-time.After(10 * time.Millisecond)
	}
	DoneCh <- struct{}{}
	return
}

//test steam for csv files
func TestSteamCsv(t *testing.T) {
	var Scd nt.ScapeDataD
	ScapeDataCh := make(chan nt.ScapeDataD)
	DoneCh := make(chan struct{})
	SteamCsv := &SteamCsv{FilePath: "../../source/source.zip"}
	go SteamCsv.Read(ScapeDataCh, DoneCh)
	fmt.Println("start test TestSteamCsv")
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
				fmt.Println("test TestSteamCsv completed successfully")
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
	fmt.Println("Start test TestElementaryDtm")
	file, errf := os.Create("operation.txt")
	if errf != nil {
		t.Errorf("Unable to create file")

	}
	defer file.Close()

	Cfg := nt.ConfigDt{}
	errf = LoadConfigYaml("../../config.yaml", &Cfg)
	if errf != nil {
		t.Fatal("not load config file")
	}
	sr := nt.DrillDataType{
		Log: CLog(),
		Cfg: &Cfg,
	}

	tm := NewDetermine(&sr, &SteamCsv{FilePath: "../../source/source1.zip"})
	//err := tm.Start(120)
	dur, err := tm.Start(60)
	tempt, _ := time.Parse("15:04:01", "00:00:00")
	fmt.Printf("duration:%s,result err:%v\n", tempt.Add(dur).Format("15:04:00.000"), err)
	//err := tm.Wait()
	if err != nil {
		t.Errorf("error:time limit exceeded")
	}
	//data := tm.GetOperationList()
	//for i := 0; i < len(data); i++ {
	//	fmt.Fprintf(file, "%s | %s |%s \r\n", data[i].StartData.Time.Format("2006-01-02 15:04:05"),
	//		data[i].StopData.Time.Format("15:04:05"),
	//		data[i].Operaton)
	//}
	FileRes, err2 := ioutil.ReadFile("./result1.txt")
	if err2 != nil {
		t.Errorf("file not find result1.txt ")
	}
	resLines := strings.Split(string(FileRes), "\r\n")
	if len(resLines) == 0 {
		t.Errorf("result1.txt is empty")
	}

	var n int
	n = 0
	data2 := tm.GetSummarysheet()

	for i := 0; i < len(data2); i++ {
		//fmt.Printf("%s | %s |%s \r\n", data2[i].Sheet.StartData.Time.Format("2006-01-02 15:04:05"),
		//	data2[i].Sheet.StopData.Time.Format("15:04:05"),
		//	data2[i].Sheet.Operaton)
		sres1 := FormatSheet(data2[i])
		//fmt.Println(sres)
		sf := resLines[n]
		if !(sres1 == sf) { // && (n > 0)
			t.Errorf("string not equale result1 ")
			t.Errorf("progrm:%s,i=%d,len=%d", sres1, i, len(sres1))
			t.Errorf("result:%s,n=%d,len=%d", sf, n, len(sf))

		}
		n = n + 1
		d3 := data2[i].Details

		for j := 0; j < len(d3); j++ {

			sres2 := FormatSheetDetails(data2[i].Details[j])
			//fmt.Println(sres)
			if !(sres2 == resLines[n]) {
				t.Errorf("string not equale result1.txt ")
				t.Errorf("programm:%s,i=%d,j=%d,len=%d", sres2, i, j, len(sres2))
				t.Errorf(" result1:%s,n=%d,len=%d", resLines[n], n, len(resLines[n]))

			}
			n = n + 1
			continue
		}
	}

	fmt.Println("test TestElementaryDtm completed successfully")
}

//very simple determine test
func TestSimpleDtm(t *testing.T) {
	fmt.Println("Start test TestSimpleDtm")

	Cfg := nt.ConfigDt{}
	errf := LoadConfigYaml("../../config.yaml", &Cfg)
	if errf != nil {
		t.Fatal("not load config file")
	}

	sr := nt.DrillDataType{
		Log: CLog(),
		Cfg: &Cfg,
	}

	tm := NewDetermine(&sr, &SteamSmpl{})
	//err := tm.Start(120)
	dur, err := tm.Start(60)
	tempt, _ := time.Parse("15:04:01", "00:00:00")
	fmt.Printf("duration:%s,result err:%v\n", tempt.Add(dur).Format("15:04:00.000"), err)
	//err := tm.Wait()
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
	neadres := [3]string{"Наращивание", "Промывка", "Бурение"}
	var dd nt.OperationOne
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
	fmt.Println("test TestSimpleDtm completed successfully")
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
