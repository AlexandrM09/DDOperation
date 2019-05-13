package main

import (
	"fmt"
	"time"
	"os"
	dtm "./determine"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("Start program", dtm.GetList())

	sr := dtm.DrillDataType{OperationList: make([]dtm.OperationOne, 0),
		SteamCh:         make(chan dtm.OperationOne),
		ScapeDataCh:     make(chan dtm.ScapeDataD),
		ErrCh:           make(chan error, 2),
		DoneCh:          make(chan struct{}),
		DoneScapeCh:     make(chan struct{}),
		ActiveOperation: -1,
		Operationtype:dtm.OperationtypeD{"Бурение",
				"Наращивание",
				"Промывка",
				"Проработка",
				"Подъем",
				"Спуск",
				"Работа т/с",
				"Бурение (ротор)","Бурение (слайд)","ПЗР","","","","",""},
		Log:CreateLog(),		
	}
	 

	tm := dtm.NewDetermine(&sr, &dtm.SteamRND{})
	_ = tm.Start()
	<-time.After(200 * time.Millisecond)
	tm.Stop()
	<-time.After(200 * time.Millisecond)
}

func CreateLog() *logrus.Logger{
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
	log.Info("Failed to log to file, using default stderr")}
	/**/
	return log
   
}
