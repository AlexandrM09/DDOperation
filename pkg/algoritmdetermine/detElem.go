package algoritmdetermine

import (
	"sync"
	"time"

	StoreMap "github.com/AlexandrM09/DDOperation/pkg/StoreMap"
	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	logrus "github.com/sirupsen/logrus"
)

type (
	//determineOne
	determineOne interface {
		Check(d *DetermineElementary, data *SaveDetElementary) (int, bool)
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
		DataMapId map[string]SaveDetElementary
		ListCheck []determineOne
		Log       *logrus.Logger
		//Mu                   *sync.RWMutex
		Wg    *sync.WaitGroup
		Cfg   *nt.ConfigDt
		Store *StoreMap.Brocker
		In    string   //pub name in busevent
		Out   []string //pub name in busevent
		//id скважин должны быть добавлены до старта,
		// в процессе работы скважины с текущей архитектурой добавлять нeльзя
		Id   map[string]int
		done chan struct{}
	}
)

func (d *DetermineElementary) AddWell(id string) {
	_, ok := d.Id[id]
	if !ok {
		d.Id[id] = 1
		d.DataMapId[id] = SaveDetElementary{ActiveOperation: -1}
	}
}

func (d *DetermineElementary) Stop() map[string]SaveDetElementary {
	defer close(d.done)
	d.done <- struct{}{}
	d.Wg.Wait()
	return d.DataMapId
}
func (d *DetermineElementary) Run(ErrCh chan error) {
	d.done = make(chan struct{})
	DoneInside := make(chan struct{})
	go func() {
		defer close(DoneInside)

		d.Wg.Add(1)
		go d.Read(DoneInside, d.done, ErrCh)
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
func (d *DetermineElementary) addDatatooperation(id string, flag int) {
	//dt.Data.Mu.Lock()
	//defer dt.Data.Mu.Unlock()
	data, ok := d.DataMapId[id]
	if !ok {
		d.Log.Debugf("DetermineElementary:something went very wrong")
		return
	}

	len := len(data.OperationList)
	if len == 0 {
		return
	}
	if !(data.OperationList[len-1].Operaton == d.Cfg.Operationtype[data.ActiveOperation]) {
		data.OperationList[len-1].Lastchangeoperation = data.OperationList[len-1].Operaton
		data.OperationList[len-1].Operaton = d.Cfg.Operationtype[data.ActiveOperation]
	}
	Op := &data.OperationList[len-1]
	g := &data.ScapeData
	Op.Count++
	for i := 4; i < 12; i++ {
		if g.Values[i] < Op.MinData.Values[i] {
			Op.MinData.Values[i] = g.Values[i]
		}
		if g.Values[i] > Op.MaxData.Values[i] {
			Op.MaxData.Values[i] = g.Values[i]
		}
		Op.Agv.Values[i] = Op.Agv.Values[i] + g.Values[i]
		if flag == 1 {
			Op.Agv.Values[i] = Op.Agv.Values[i] / float32(Op.Count)
		}
	}
	d.DataMapId[id] = data
}
func (d *DetermineElementary) startnewoperation(id string) {
	data, ok := d.DataMapId[id]
	if !ok {
		d.Log.Debugf("DetermineElementary:something went very wrong")
		return
	}
	g := data.ScapeData
	if data.Temp.FlagChangeTrip == 1 {
		data.Temp.FlagChangeTrip = 0
		g = data.Temp.LastTripData
	}
	data.OperationList = append(data.OperationList,
		nt.OperationOne{Operaton: d.Cfg.Operationtype[data.ActiveOperation], StartData: g, Status: "start"})
	data.Temp.LastStartData = g
	data.StartActiveOperation = g.Time //
	dtmp := data.OperationList[len(data.OperationList)-1]
	d.Log.Debugf("Send Determine id=%s,time=%s,Op=%s", id, dtmp.StartData.Time.Format("15:04"), dtmp.Operaton)
	d.DataMapId[id] = data
	d.Store.Send("Determine", id, &dtmp)
	d.Log.Debug("Start operation")
}
func (d *DetermineElementary) saveoperation(id string) {
	data, ok := d.DataMapId[id]
	if !ok {
		d.Log.Debugf("DetermineElementary:something went very wrong")
		return
	}
	len := len(data.OperationList)
	if len == 0 {
		return
	}
	if data.Temp.FlagChangeTrip == 1 {
		//dt.Data.temp.FlagChangeTrip=0
		data.OperationList[len-1].StopData = data.Temp.LastTripData
		//l.Printf("FlagChangeTrip == 1")
		d.Log.Debug("FlagChangeTrip == 1")
	} else {
		data.OperationList[len-1].StopData = data.LastScapeData
	}
	data.OperationList[len-1].Status = "save"
	d.DataMapId[id] = data
	dtmp := data.OperationList[len-1]
	d.Log.Debugf("Send Determine id=%s,time=%s,Op=%s", id, dtmp.StartData.Time.Format("15:04"), dtmp.Operaton)
	d.Store.Send("Determine", id, &dtmp)
	d.Log.Debug("Stop and save  operation ")
}
func (d *DetermineElementary) Read(DoneInside chan struct{}, Done chan struct{}, ErrCh chan error) {
	defer func() {
		DoneInside <- struct{}{}
		d.Log.Infof("Exit read DetermineElementary ")
	}()
	d.Log.Infof("Start Run DetEl ")
	//init
	var res int
	var changeOp bool
	//var ok bool
	for key, v := range d.DataMapId {
		v.ActiveOperation = -1
		d.DataMapId[key] = v
	}
	// num interface check
	checkInt := [11]int{0, 1, 2, 3, 4, 5, 0, 6, 7, 8, 9}
	d.ListCheck = []determineOne{&Check0{}, &Check1{},
		&Check2{}, &Check3{}, &Check4{}, &Check5{}, &Check7{}, &Check8{}, &Check9{}, &Check10{}}
	for {
		for id, _ := range d.Id {
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
			g := d.Store.Receive(d.In, id) //"ScapeData"
			//	d.Log.Infof("Read ScapeData id=%s,ii2=%v", d.Id, ii2)
			if g == nil {
				continue
			}
			var temp1 *nt.ScapeDataD
			temp1, ok := g.(*nt.ScapeDataD)
			data, ok := d.DataMapId[id]
			if !ok {
				d.Log.Debugf("DetermineElementary:something went very wrong")
				return
			}
			data.ScapeData = *temp1
			if !ok {
				continue
			}
			d.Log.Infof("after parse ScapeData id=%s,", d.Id)
			if data.ActiveOperation >= 0 {
				res, changeOp = d.ListCheck[checkInt[data.ActiveOperation]].Check(d, &data)
			} else {
				res = -1
			}
			if res == -1 {
				for i := 0; i < len(d.ListCheck) && (res == -1); i++ {
					res, changeOp = d.ListCheck[i].Check(d, &data)
				}
			} // select operation
			if res == -1 {
				res = len(d.ListCheck) - 1
				changeOp = false
			}
			switch {
			case res == data.ActiveOperation:
				{ //addDatatooperation
					d.addDatatooperation(id, 0)
				}
			default:
				{
					if !changeOp {
						d.addDatatooperation(id, 1)
						d.saveoperation(id)
					}
					d.Log.Debug(" after d.AtiveOperation = res")
					data.ActiveOperation = res
					if changeOp {
						d.addDatatooperation(id, 0)
					}
					if !changeOp {
						d.startnewoperation(id)
						d.addDatatooperation(id, 0)
					}

					if changeOp {
						changeOp = false
					}
					//d.startnewoperation()
				}

			}
			data.LastScapeData = data.ScapeData
		}
	}
}
