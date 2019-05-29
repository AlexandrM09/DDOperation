package main

import (
	"fmt"
	//"io/ioutil"
	//"time"
	//"encoding/json"
	//"log"
	"log"
	"os"

	dtm "./determine"
	"github.com/sirupsen/logrus"
	_ "gopkg.in/yaml.v2"
	//	"sync"
)

func main() {
	Cfg := dtm.ConfigDt{}
	errf := dtm.LoadConfigYaml("./config.yaml", &Cfg)
	if errf != nil {
		log.Fatal("not load config file")
	}
	sr := dtm.DrillDataType{
		Log: createLog(),
		Cfg: &Cfg,
	}

	tm := dtm.NewDetermine(&sr, &dtm.SteamCsv{FilePath: "./source/source2.zip",
	  SatartTime:"___2019-05-25 17:52:43"})
	_ = tm.Start(60)
	err := tm.Wait()
	if err != nil {
		log.Fatal("error:time limit exceeded")
	}


	data2 := tm.GetSummarysheet()
	fmt.Printf("Start print Summarysheet len=%v \n", len(data2))
	for i := 0; i < len(data2); i++ {
		fmt.Printf("%s | %s |%s \r\n", data2[i].Sheet.StartData.Time.Format("2006-01-02 15:04:05"),
			data2[i].Sheet.StopData.Time.Format("15:04:05"),
			data2[i].Sheet.Operaton)
		d3 := data2[i].Details
		for j := 0; j < len(d3); j++ {
			fmt.Printf("____ %s | %s |%s \r\n", d3[j].StartData.Time.Format("15:04:05"),
				d3[j].StopData.Time.Format("15:04:05"),
				d3[j].Operaton)

		}
	}

}
func createLog() *logrus.Logger {
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
