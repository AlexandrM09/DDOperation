package main

import (
	"fmt"
	"time"

	dtm "./determine"
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
	}

	tm := dtm.NewDetermine(&sr, &dtm.SteamRND{})
	_ = tm.Start()
	<-time.After(200 * time.Millisecond)
}
