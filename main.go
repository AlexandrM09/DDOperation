package main

import (
	"fmt"

	dtm "./determine"
)

func main() {
	fmt.Println("Start program", dtm.GetList())

	sr := dtm.DrillDataType{OperationList: make([]dtm.OperationOne, 1),
		SteamCh:     make(chan dtm.OperationOne),
		ScapeDataCh: make(chan dtm.ScapeDataD),
		DoneCh:      make(chan struct{}),
	}

	tm := dtm.NewDetermine(&sr, &dtm.SteamRND{})
	_ = tm.Read()

}
