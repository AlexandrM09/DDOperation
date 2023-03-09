package algoritmdetermine

import (
	"context"
	"sync"

	// StoreMap "github.com/AlexandrM09/DDOperation/pkg/StoreMap"
	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	logrus "github.com/sirupsen/logrus"
)

type (
	//determineOne
	determineOne interface {
		Check(d *DetermineElementary, data *nt.SaveDetElementary) (int, bool)
	}
	//SaveDeteElementary
	store interface {
		DElementaryGet(id string) (*nt.SaveDetElementary,bool)
		DElementarySet(id string, v *nt.SaveDetElementary)
		DElementaryGetAll() map[string]*nt.SaveDetElementary
	}

	//DrillDataType drill basic data struct
	DetermineElementary struct {
		// DataMapId map[string]nt.SaveDetElementary
		ListCheck []determineOne
		Log       *logrus.Logger
		//Mu                   *sync.RWMutex
		Wg  *sync.WaitGroup
		Cfg *nt.ConfigDt
		// Store *StoreMap.Brocker
		//In    string   //pub name in busevent
		In  chan interface{} //*nt.ScapeDataD
		Out chan interface{} //*nt.SendingTopicDeterm
		//Out   []string //pub name in busev ent
		//id скважин должны быть добавлены до старта,
		// в процессе работы скважины с текущей архитектурой добавлять нeльзя
		doneWait chan struct{}
		ctx      context.Context
		flagExit bool
		store    store
	}
)

func NewDetElementary(ctx context.Context, in chan interface{}, out chan interface{}, l *logrus.Logger, cfg *nt.ConfigDt, wellid []string, so store) *DetermineElementary {
	dt := DetermineElementary{
		Log: l,
		Wg:  &sync.WaitGroup{},
		Cfg: cfg,
		In:  in,
		Out: out,
		// DataMapId: make(map[string]nt.SaveDetElementary, len(wellid)),
		// Store:     store,
		ctx:      ctx,
		doneWait: make(chan struct{}),
		store:    so,
	}
	for ind, v := range wellid {
		dt.Log.Debugf("before e.AddWell(v.Id),id:%s\n", v)
		dt.addWell(ind, v)
	}
	return &dt
}
func (d *DetermineElementary) addWell(index int, id string) {
	// d.DataMapId[id] = nt.SaveDetElementary{IdWell: id, ActiveOperation: -1}
	d.store.DElementarySet(id, &nt.SaveDetElementary{IdWell: id, ActiveOperation: -1})
}

// Ожидание пока хотябы в одной скважине поступают данные
func (d *DetermineElementary) WaitandGetReault() map[string]*nt.SaveDetElementary {
	d.flagExit = true
	// d.Wg.Wait()
	<-d.doneWait
	d.Log.Info("DetermineElementary WaitandGetReault done")
	return d.store.DElementaryGetAll()
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
func (d *DetermineElementary) addDatatooperation(data *nt.SaveDetElementary, id string, flag int) {
	len := len(data.OperationList)
	if len == 0 {
		d.Log.Debugf("DetermineElementary addDatatooperation len==0")
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

}
func (d *DetermineElementary) startnewoperation(data *nt.SaveDetElementary, id string) {

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
	d.Out <- &dtmp
	d.Log.Debug("Start operation")
}
func (d *DetermineElementary) saveoperation(data *nt.SaveDetElementary, id string) {
	d.Log.Debugf("DetermineElementary Start saveoperation %s", id)

	lenOpl := len(data.OperationList)
	if lenOpl == 0 {
		d.Log.Debugf("DetermineElementary saveoperation len==0")
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

	dtmp := nt.SendingTopicDeterm{
		IdWell:    id,
		Operation: data.OperationList[lenOpl-1],
	}

	d.Log.Debugf("Send Determine save %v", dtmp)
	d.Out <- &dtmp
	d.Log.Debug("Stop and save  operation ")
}
func (d *DetermineElementary) Read(ctx context.Context, DoneInside chan struct{}, ErrCh chan error) {
	defer func() {
		DoneInside <- struct{}{}
		close(d.Out)

		d.Log.Infof("Exit read DetermineElementary,close(d.Out)")
	}()
	d.Log.Infof("Start Run DetEl ")
	//init
	var res int
	var changeOp bool
	//var ok bool
	// for key, v := range d.DataMapId {
	// 	v.ActiveOperation = -1
	// 	d.DataMapId[key] = v
	// }
	// num interface check
	checkInt := [11]int{0, 1, 2, 3, 4, 5, 0, 6, 7, 8, 9}
	d.ListCheck = []determineOne{&Check0{}, &Check1{},
		&Check2{}, &Check3{}, &Check4{}, &Check5{}, &Check7{}, &Check8{}, &Check9{}, &Check10{}}
	localdone := make(chan struct{})
	defer close(localdone)
	go func() {
		for g := range d.In {

			tmp, ok := g.(nt.ScapeDataD)
			if !ok {
				d.Log.Debugf("data casting failed")
				continue
			}
			data, ok := d.store.DElementaryGet(tmp.Id)
			if !ok {
				d.Log.Debugf("unknown well")
				continue
			}
			//last line data
			if tmp.StatusLastData {
				dtmp := nt.SendingTopicDeterm{
					IdWell:    tmp.Id,
					Operation: nt.OperationOne{Status: "lastline"},
				}
				d.Out <- &dtmp
				continue
			}
			data.ScapeData = tmp
			d.Log.Infof("after parse ScapeData id=%s, data=%v", data.IdWell, data.ScapeData)
			if data.ActiveOperation >= 0 {
				res, changeOp = d.ListCheck[checkInt[data.ActiveOperation]].Check(d, data)
				// d.Log.Infof("after d.ListCheck[] res=%v,changeOp=%v", res, changeOp)
			} else {
				res = -1
			}
			if res == -1 {
				for i := 0; i < len(d.ListCheck) && (res == -1); i++ {
					res, changeOp = d.ListCheck[i].Check(d, data)
				}
			} // select operation
			if res == -1 {
				res = len(d.ListCheck) - 1
				changeOp = false
			}
			d.Log.Debugf("1 before switch data.ActiveOperation =%d,res=%d", data.ActiveOperation, res)
			switch {
			case res == data.ActiveOperation:
				{ //addDatatooperation
					// d.Log.Infof("before d.addDatatooperation id=%s", id)
					d.addDatatooperation(data, data.IdWell, 0)
					// d.Log.Infof("after d.addDatatooperation id=%s", id)
				}
			default:
				{
					d.Log.Infof("before switch default id=%s ,changeOp=%v", data.IdWell, changeOp)
					if !changeOp {
						d.Log.Infof("if !changeOp before switch default id=%s ,changeOp=%v", data.IdWell, changeOp)
						d.addDatatooperation(data, data.IdWell, 1)
						d.saveoperation(data, data.IdWell)
					}
					// d.Log.Debugf(" after d.AtiveOperation = res,data.ActiveOperation =%d,res=%d", data.ActiveOperation, res)
					data.ActiveOperation = res
					d.store.DElementarySet(data.IdWell, data)
					if changeOp {
						d.addDatatooperation(data, data.IdWell, 0)
					}
					if !changeOp {
						d.Log.Debugf(" before startnewoperation data.ActiveOperation =%d,res=%d", data.ActiveOperation, res)

						d.startnewoperation(data, data.IdWell)
						d.addDatatooperation(data, data.IdWell, 0)
					}

					if changeOp {
						changeOp = false
					}
					//d.startnewoperation()
					// d.Log.Infof("after switch default id=%s ,changeOp=%v",data.IdWell, changeOp)
				}

			}
			data.LastScapeData = data.ScapeData
			d.store.DElementarySet(data.IdWell, data)
			// d.Log.Infof("after switch id=%s ,changeOp=%v", data.IdWell, changeOp)

		}
		d.Log.Debug("exit func DetermineElementary.Read(close chanel)")
		localdone <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		{
			d.Log.Info("exit to DetermineElementary(context cancel)")
			return
		}
	case <-localdone:
		{
			d.Log.Info("exit to DetermineElementary(<-localdone)")
			return
		}
	}
}
