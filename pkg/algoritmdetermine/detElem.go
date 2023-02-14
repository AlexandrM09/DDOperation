package algoritmdetermine

import (
	"context"
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
		IdWell               string
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
		Out   []string //pub name in busev ent
		//id скважин должны быть добавлены до старта,
		// в процессе работы скважины с текущей архитектурой добавлять нeльзя
		doneWait chan struct{}
		ctx      context.Context
		flagExit bool
	}
)

func NewDetElementary(ctx context.Context, in string, out []string, l *logrus.Logger, cfg *nt.ConfigDt, store *StoreMap.Brocker, wellid []string) *DetermineElementary {
	dt := DetermineElementary{
		Log:       l,
		Wg:        &sync.WaitGroup{},
		Cfg:       cfg,
		In:        in,
		Out:       out,
		DataMapId: make(map[string]SaveDetElementary, len(wellid)),
		Store:     store,
		ctx:       ctx,
		doneWait:  make(chan struct{}),
	}
	for ind, v := range wellid {
		dt.Log.Debugf("before e.AddWell(v.Id),id:%s\n", v)
		dt.addWell(ind, v)
	}
	return &dt
}
func (d *DetermineElementary) addWell(index int, id string) {
	d.DataMapId[id] = SaveDetElementary{IdWell: id, ActiveOperation: -1}
}

// Ожидание пока хотябы в одной скважине поступают данные
func (d *DetermineElementary) WaitandGetReault() map[string]SaveDetElementary {
	d.flagExit = true
	// d.Wg.Wait()
	<-d.doneWait
	d.Log.Info("DetermineElementary WaitandGetReault done")
	return d.DataMapId
}
func (d *DetermineElementary) Run(ErrCh chan error) {

	DoneInside := make(chan struct{})
	d.Wg.Add(1)
	go func() {
		defer close(DoneInside)
		defer close(d.doneWait)

		go d.Read(d.ctx, DoneInside, ErrCh)
		for {
			select {

			case <-DoneInside:
				{
					d.Log.Info("DetermineElementary Run case <-DoneInside done")
					d.doneWait <- struct{}{}
					d.Wg.Done()
					return
				}
			}
		}
	}()

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
	d.Log.Debugf("DetermineElementary startnewoperation data.ActiveOperation =%d", data.ActiveOperation)

	data.OperationList = append(data.OperationList,
		nt.OperationOne{Operaton: d.Cfg.Operationtype[data.ActiveOperation], StartData: g, Status: "start"})
	data.Temp.LastStartData = g
	data.StartActiveOperation = g.Time //
	dtmp := nt.SendingTopicDeterm{
		IdWell:    data.IdWell,
		Operation: data.OperationList[len(data.OperationList)-1],
	}
	d.Log.Debugf("Send Determine id=%s,time=%s,Op=%s", id, dtmp.Operation.StartData.Time.Format("15:04"), dtmp.Operation.Operaton)
	d.DataMapId[id] = data
	for ind := range d.Out {
		d.Store.Send(d.Out[ind], &dtmp)
	}

	d.Log.Debug("Start operation")
}
func (d *DetermineElementary) saveoperation(id string) {
	data, ok := d.DataMapId[id]
	if !ok {
		d.Log.Debugf("DetermineElementary:something went very wrong")
		return
	}
	lenOpl := len(data.OperationList)
	if lenOpl == 0 {
		return
	}
	if data.Temp.FlagChangeTrip == 1 {
		//dt.Data.temp.FlagChangeTrip=0
		data.OperationList[lenOpl-1].StopData = data.Temp.LastTripData
		//l.Printf("FlagChangeTrip == 1")
		d.Log.Debug("FlagChangeTrip == 1")
	} else {
		data.OperationList[lenOpl-1].StopData = data.LastScapeData
	}
	data.OperationList[lenOpl-1].Status = "save"
	d.DataMapId[id] = data
	dtmp := nt.SendingTopicDeterm{
		IdWell:    data.IdWell,
		Operation: data.OperationList[lenOpl-1],
	}

	d.Log.Debugf("Send Determine id=%s,time=%s,Op=%s", id, dtmp.Operation.StartData.Time.Format("15:04"), dtmp.Operation.Operaton)
	for ind := range d.Out {
		d.Store.Send(d.Out[ind], &dtmp)
	}
	d.Log.Debug("Stop and save  operation ")
}
func (d *DetermineElementary) Read(ctx context.Context, DoneInside chan struct{}, ErrCh chan error) {
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
	countEmpty := 0
	for {
		countEmpty += 1
		if countEmpty > 500 && d.flagExit {
			d.Log.Error("DetermineElementary AllSteamsEmpty")
			return
		}
		d.Log.Infof("countEmpty=%d,flagExit=%v", countEmpty, d.flagExit)

		select {
		case <-ctx.Done():
			{
				d.Log.Info("on-demand output DetermineElementary")
				return
			}
		default:
			{
			}
		}
		//read data

		g := d.Store.Receive(d.In) //"ScapeData"

		if g == nil {
			d.Log.Infof("Read ScapeData nil in=%s,value=%v,countEmpty=%d", d.In, g, countEmpty)
			continue
		}

		var temp1 *nt.ScapeDataD
		temp1, ok := g.(*nt.ScapeDataD)
		if !ok {
			d.Log.Debugf("unknown data well")
			continue
		}

		data, ok := d.DataMapId[temp1.Id]
		if !ok {
			d.Log.Debugf("unknown well")
			continue
		}

		data.ScapeData = *temp1
		if !ok {
			continue
		}
		countEmpty = 0
		d.Log.Infof("after parse ScapeData id=%s, data=%v", data.IdWell, data.ScapeData)
		if data.ActiveOperation >= 0 {
			res, changeOp = d.ListCheck[checkInt[data.ActiveOperation]].Check(d, &data)
			// d.Log.Infof("after d.ListCheck[] res=%v,changeOp=%v", res, changeOp)
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
				// d.Log.Infof("before d.addDatatooperation id=%s", id)
				d.addDatatooperation(data.IdWell, 0)
				// d.Log.Infof("after d.addDatatooperation id=%s", id)
			}
		default:
			{
				// d.Log.Infof("before switch default id=%s ,changeOp=%v", data.IdWell, changeOp)
				if !changeOp {
					d.addDatatooperation(data.IdWell, 1)
					d.saveoperation(data.IdWell)
				}
				// d.Log.Debugf(" after d.AtiveOperation = res,data.ActiveOperation =%d,res=%d", data.ActiveOperation, res)
				data.ActiveOperation = res
				d.DataMapId[data.IdWell] = data
				if changeOp {
					d.addDatatooperation(data.IdWell, 0)
				}
				if !changeOp {
					// d.Log.Debugf(" after d.startnewoperation(id) ,data.ActiveOperation =%d,res=%d", data.ActiveOperation, res)

					d.startnewoperation(data.IdWell)
					d.addDatatooperation(data.IdWell, 0)
				}

				if changeOp {
					changeOp = false
				}
				//d.startnewoperation()
				// d.Log.Infof("after switch default id=%s ,changeOp=%v",data.IdWell, changeOp)
			}

		}
		data.LastScapeData = data.ScapeData
		d.DataMapId[data.IdWell] = data
		// d.Log.Infof("after switch id=%s ,changeOp=%v", data.IdWell, changeOp)

	}
}
