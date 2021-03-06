package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"

	dtm "./determine"
	logrus "github.com/sirupsen/logrus"
	_ "gopkg.in/yaml.v2"
)

func main() {
	Cfg := dtm.ConfigDt{}
	errf := dtm.LoadConfigYaml("./config.yaml", &Cfg)
	if errf != nil {
		log.Fatal("not load config file")
	}
	sr := dtm.DrillDataType{
		Log: createLog(logrus.DebugLevel),
		Cfg: &Cfg,
	}
	tm := dtm.NewDetermine(&sr, &dtm.SteamCsv{
		FilePath:   "./source/source1.zip",
		SatartTime: "___2019-05-25 17:52:43",
	})
	err := tm.Start(60)
	if err != nil {
		log.Fatal("error:time limit exceeded")
	}
	data2 := tm.GetSummarysheet()
	fmt.Printf("Start print Summarysheet len=%v \n", len(data2))
	for i := 0; i < len(data2); i++ {
		fmt.Println(dtm.FormatSheet(data2[i]))

		//d3 := data2[i].Details
		for j := 0; j < len(data2[i].Details); j++ {
			//d4 := d3[j]
			fmt.Println(dtm.FormatSheetDetails(data2[i].Details[j]))

		}
	}
	fmt.Printf("Start print Summarysheet short form len=%v \n", len(data2))
	fmt.Println(data2[0].Sheet.StartData.Time.Format("2006-01-02"))
	//tempt, _ := time.Parse("15:04:01", "00:00:00")
	for i := 0; i < len(data2); i++ {
		fmt.Println(dtm.FormatSheet2(data2[i]))
		//	dur := data2[i].Sheet.StopData.Time.Sub(data2[i].Sheet.StartData.Time)
		//	fmt.Printf("%s | %s |%s |%s %s \r\n",
		//		data2[i].Sheet.StartData.Time.Format("15:04"),
		//		data2[i].Sheet.StopData.Time.Format("15:04"),
		//		tempt.Add(dur).Format("15:04"),
		//		data2[i].Sheet.Operaton,
		//		data2[i].Sheet.Params) //,data2[i].Sheet.Agv.Values)
	}
}
func createLog(ll logrus.Level) *logrus.Logger {
	var log = logrus.New()
	//	log.WithFields(logrus.Fields{
	//		//"mode":   "[access_log]",
	//		"logger": "LOGRUS",
	//	})
	//	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetReportCaller(true)

	log.SetFormatter(&logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			_, filename := path.Split(f.File)
			filename = fmt.Sprintf("%s:%d", filename, f.Line)
			return "", filename
		},
		DisableColors: false,
		//	FullTimestamp: true,
	},
	)

	log.SetLevel(ll)
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
