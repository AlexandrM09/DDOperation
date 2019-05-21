package main

import (
	"fmt"
	//"io/ioutil"
	//"time"
	//"encoding/json"
	//"log"
	"os"
  _ "gopkg.in/yaml.v2"
    dtm "./determine"
	"github.com/sirupsen/logrus"
	//	"sync"
)

func main() {
	fmt.Println("Start program", dtm.GetList())
	cfg := dtm.ConfigDt{}
	err:=dtm.LoadConfig("config.json",&cfg)
	if err == nil {
		fmt.Println(cfg)
	} else {
		fmt.Println("not parse config.json", cfg)
	}


	
	/*	sr := dtm.DrillDataType{OperationList: make([]dtm.OperationOne, 0),
			SteamCh:         make(chan dtm.OperationOne),
			ScapeDataCh:     make(chan dtm.ScapeDataD),
			ErrCh:           make(chan error, 2),
			DoneCh:          make(chan struct{}),
			DoneScapeCh:     make(chan struct{}),
			ActiveOperation: -1,
			Operationtype:dtm.DrillOperationConst,
			Log:CreateLog(),

		}


		tm := dtm.NewDetermine(&sr, &dtm.SteamRND{})
		_ = tm.Start(5)
		_=tm.Wait()
	*/

	/*decoder = json.NewDecoder(file2)
	  config := new(Config)
	  err = decoder.Decode(&config)
	  if err != nil {
	  	// handle it
	  	fmt.Printf("%+v",config)
	  }
	*/

}
func CreateLog() *logrus.Logger {
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
