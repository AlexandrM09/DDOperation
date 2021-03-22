package algoritmdetermine

import (
	"sync"
	"time"

	bus "github.com/AlexandrM09/DDOperation/pkg/eventbussimple"
	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	logrus "github.com/sirupsen/logrus"
)

type (
	//determineOne
	determineOne interface {
		Check(d *DetermineElementary) (int, bool)
	}
	//SaveDeteElementary
	SaveDetElementary = struct {
		OperationList        []nt.OperationOne
		ScapeFullData        bool
		LastScapeData        nt.ScapeDataD
		ScapeData            nt.ScapeDataD
		ActiveOperation      int
		StartActiveOperation time.Time
		Temp                 struct {
			LastToolDepht     float32
			LastTimeToolDepht time.Time
			StartDepht        float32
			LastStartData     nt.ScapeDataD
			LastTripData      nt.ScapeDataD
			FlagChangeTrip    int
		}
	}
	//DrillDataType drill basic data struct
	DetermineElementary struct {
		data      SaveDetElementary
		ListCheck []determineOne
		Log       *logrus.Logger
		//Mu                   *sync.RWMutex
		Wg    *sync.WaitGroup
		Cfg   *nt.ConfigDt
		evnt  *bus.Eventbus
		IdIn  string //pub name in busevent
		IdOut string //pub name in busevent
		Id    string
	}
	//DetermineElementary well
	DetermineElementaryI2 interface {
		Run(Done chan struct{}, ErrCh chan error)
		//ReadTime(Done chan struct{}, ErrCh chan error)
		Stop(Done chan struct{}) SaveDetElementary
	}
)

func (d *DetermineElementary) Stop(Done chan struct{}) SaveDetElementary {
	Done <- struct{}{}
	d.Wg.Wait()
	return d.data
}
func (d *DetermineElementary) Run(Done chan struct{}, ErrCh chan error) {

	DoneInside := make(chan struct{})
	go func() {
		defer func() {
			close(DoneInside)
		}()

		go d.Read(DoneInside, Done, ErrCh)
		d.Wg.Add(1)
		for {
			select {
			//case <-ErrCh:
			//	return
			case <-DoneInside:
				{
					d.Wg.Done()
					return
				}
			default:
			}
		}
	}()
	return
}
func (d *DetermineElementary) addDatatooperation(flag int) {
	//dt.Data.Mu.Lock()
	//defer dt.Data.Mu.Unlock()
	len := len(d.data.OperationList)
	if len == 0 {
		return
	}
	if !(d.data.OperationList[len-1].Operaton == d.Cfg.Operationtype[d.data.ActiveOperation]) {
		d.data.OperationList[len-1].Lastchangeoperation = d.data.OperationList[len-1].Operaton
		d.data.OperationList[len-1].Operaton = d.Cfg.Operationtype[d.data.ActiveOperation]
	}
	Op := &d.data.OperationList[len-1]
	data := &d.data.ScapeData
	Op.Count++
	for i := 4; i < 12; i++ {
		if data.Values[i] < Op.MinData.Values[i] {
			Op.MinData.Values[i] = data.Values[i]
		}
		if data.Values[i] > Op.MaxData.Values[i] {
			Op.MaxData.Values[i] = data.Values[i]
		}
		Op.Agv.Values[i] = Op.Agv.Values[i] + data.Values[i]
		if flag == 1 {
			Op.Agv.Values[i] = Op.Agv.Values[i] / float32(Op.Count)
		}
	}
}
func (d *DetermineElementary) startnewoperation() {

	//dt.Data.Mu.Lock()
	//defer dt.Data.Mu.Unlok()

	tempData := d.data.ScapeData
	if d.data.Temp.FlagChangeTrip == 1 {
		d.data.Temp.FlagChangeTrip = 0
		tempData = d.data.Temp.LastTripData
	}
	d.data.OperationList = append(d.data.OperationList,
		nt.OperationOne{Operaton: d.Cfg.Operationtype[d.data.ActiveOperation], StartData: tempData, Status: "start"})
	d.data.Temp.LastStartData = tempData
	d.data.StartActiveOperation = tempData.Time //
	dtmp := d.data.OperationList[len(d.data.OperationList)-1]
	d.evnt.Send("Determine", d.Id, &dtmp)
	//	d.data.SteamCh <- dt.Data.OperationList[len(dt.Data.OperationList)-1]
	d.Log.Debug("Start operation")
}
func (d *DetermineElementary) saveoperation() {
	//
	//dt.Data.Mu.Lock()
	//defer dt.Data.Mu.Unlock()
	len := len(d.data.OperationList)
	if len == 0 {
		return
	}
	if d.data.Temp.FlagChangeTrip == 1 {
		//dt.Data.temp.FlagChangeTrip=0
		d.data.OperationList[len-1].StopData = d.data.Temp.LastTripData
		//l.Printf("FlagChangeTrip == 1")
		d.Log.Debug("FlagChangeTrip == 1")
	} else {
		d.data.OperationList[len-1].StopData = d.data.LastScapeData
	}
	d.data.OperationList[len-1].Status = "save"
	dtmp := d.data.OperationList[len-1]
	d.evnt.Send("Determine", d.Id, &dtmp)
	d.Log.Debug("Stop and save  operation ")
}
func (d *DetermineElementary) Read(DoneCh chan struct{}, Done chan struct{}, ErrCh chan error) {
	defer func() {

		DoneCh <- struct{}{}
		d.Log.Info("Exit read DetermineElementary")
	}()
	//init
	var res int
	var changeOp bool
	var ok bool
	d.data.ActiveOperation = -1
	var temp1 *nt.ScapeDataD
	// num interface check
	checkInt := [11]int{0, 1, 2, 3, 4, 5, 0, 6, 7, 8, 9}
	d.ListCheck = []determineOne{&Check0{}, &Check1{},
		&Check2{}, &Check3{}, &Check4{}, &Check5{}, &Check7{}, &Check8{}, &Check9{}, &Check10{}}
	for {

		select {
		case <-Done:
			{
				d.Log.Info("on-demand output DetermineElementary")
				return
			}
		default:
			{
			}
		}
		//read data
		ii2 := d.evnt.Receive("ScapeData", d.Id)
		if ii2 == nil {
			continue
		}

		temp1, ok = ii2.(*nt.ScapeDataD)
		d.data.ScapeData = *temp1
		if !ok {
			continue
		}
		if d.data.ActiveOperation >= 0 {
			res, changeOp = d.ListCheck[checkInt[d.data.ActiveOperation]].Check(d)
		} else {
			res = -1
		}
		if res == -1 {
			for i := 0; i < len(d.ListCheck) && (res == -1); i++ {
				res, changeOp = d.ListCheck[i].Check(d)
			}
		} // select operation
		if res == -1 {
			res = len(d.ListCheck) - 1
			changeOp = false
		}
		switch {
		case res == d.data.ActiveOperation:
			{ //addDatatooperation
				d.addDatatooperation(0)
			}
		default:
			{

				if !changeOp {
					d.addDatatooperation(1)
					d.saveoperation()

				}
				d.Log.Debug(" after d.AtiveOperation = res")
				d.data.ActiveOperation = res
				if changeOp {
					d.addDatatooperation(0)
				}
				if !changeOp {
					d.startnewoperation()
					d.addDatatooperation(0)
				}

				if changeOp {
					changeOp = false
				}
				d.startnewoperation()
			}

		}

		d.data.LastScapeData = d.data.ScapeData
	}
}
