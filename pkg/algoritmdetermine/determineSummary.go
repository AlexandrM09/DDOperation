package algoritmdetermine

import (
	"context"
	"fmt"
	"sync"
	_ "time"

	StoreMap "github.com/AlexandrM09/DDOperation/pkg/StoreMap"
	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	"github.com/sirupsen/logrus"
)

type (

	//SummarysheetT -type result list
	// SummaryCalcT struct {
	// 	Sheet   nt.OperationOne
	// 	Details []nt.OperationOne

	// 	Temp struct {
	// 		LastToolDepht     float32
	// 		LastTimeToolDepht time.Time
	// 		StartDepht        float32
	// 		LastStartData     nt.ScapeDataD
	// 		LastTripData      nt.ScapeDataD
	// 		FlagChangeTrip    int
	// 	}
	// }
	SummaryResult struct {
		IdWell       string
		Summarysheet []nt.SummarysheetT
		Sc           nt.ResultSheet
	}
	DetermineSummarys struct {
		DataMapId map[string]SummaryResult
		Log       *logrus.Logger
		//Mu                   *sync.RWMutex
		Wg    *sync.WaitGroup
		Cfg   *nt.ConfigDt
		Store *StoreMap.Brocker
		//In    string   //pub name in busevent
		In chan interface{}
		//Out   []string //pub name in busevent
		Out chan interface{}
		//id скважин должны быть добавлены до старта,
		// в процессе работы скважины с текущей архитектурой добавлять нeльзя
		stateRun bool //true=runing,false=stop
		Ctx      context.Context
	}
)

func New(ctx context.Context, in chan interface{}, out chan interface{}, l *logrus.Logger, cfg *nt.ConfigDt, wellid []string) *DetermineSummarys {
	ds := DetermineSummarys{
		DataMapId: make(map[string]SummaryResult),
		Log:       l,
		Wg:        &sync.WaitGroup{},
		Cfg:       cfg,
		// Store:     store,
		In:  in,
		Out: out,
		Ctx: ctx,
		// done:make(chan struct{},1),
	}
	for i := range wellid {
		ds.AddWell(wellid[i])
	}
	return &ds
}
func (d *DetermineSummarys) AddWell(id string) {
	d.DataMapId[id] = SummaryResult{IdWell: id}
}

// Ожидание пока хотябы в одной скважине поступают данные
func (d *DetermineSummarys) WaitandGetReault() map[string]SummaryResult {

	d.Wg.Wait()
	d.Log.Infof("DetermineSummarys after d.Wg.Wait() ")
	return d.DataMapId
}

func (d *DetermineSummarys) Run(ErrCh chan error) {
	d.stateRun = true

	DoneInside := make(chan struct{})
	go func() {
		defer close(DoneInside)
		d.Wg.Add(1)
		go d.Read(d.Ctx, DoneInside, ErrCh)
		for {
			select {
			case <-DoneInside:
				{
					d.stateRun = false
					d.Wg.Done()
					return
				}
			}
		}
	}()

}
func (d *DetermineSummarys) Read(ctx context.Context, DoneInside chan struct{}, ErrCh chan error) {
	defer func() {
		DoneInside <- struct{}{}
		d.Log.Infof("Exit read DetermineSummarys ")
	}()
	d.Log.Infof("Start Run DetSummary")
	localdone := make(chan struct{})
	defer close(localdone)
	go func() {
		for g := range d.In {

			select {
			case <-ctx.Done():
				{
					d.Log.Info("DetermineSummarys Read output by context")
					return
				}
			default:
				{
				}
			}
			//read data
			// g := d.Store.Receive(d.In) //"ScapeData"
			//	d.Log.Infof("Read ScapeData id=%s,ii2=%v", d.Id, ii2)
			// if g == nil {
			// 	continue
			// }
			v, ok := g.(*nt.SendingTopicDeterm)
			if !ok {
				d.Log.Debugf("DetermineElementary:data casting failed")
				continue
			}
			// var resStr *nt.OperationOne
			resStr := &v.Operation

			data, ok := d.DataMapId[v.IdWell]
			if !ok {
				d.Log.Debugf("DetermineElementary:something went very wrong")
				return
			}
			//last line data
			if resStr.Status == "lastline" {
				len2 := len(data.Sc.ResSheet.Details)
				if len2 > 0 {
					d.Log.Debug("done and save operation idwell=%s", v.IdWell)
					data.Sc.ResSheet.Sheet.StopData = data.Sc.ResSheet.Details[len2-1].StopData
					data.Summarysheet = append(data.Summarysheet, *d.addSummaryStr(v.IdWell, &data.Sc.ResSheet))
					d.DataMapId[v.IdWell] = data
				}
				continue
			}
			d.Log.Debugf("Start dSumm:  status:%s,Id:%s", resStr.Status, v.IdWell)
			if data.Sc.Firstflag == 0 {
				//Самая первая операция в списке
				d.Log.Debug("if dt.itemNew.firstflag == 0 {")
				if resStr.Status == "start" {
					data.Sc.StartTime = resStr.StartData.Time
					d.Log.Debugf("one!! if resStr.status == start,id=%s", v.IdWell)
				}
				if resStr.Status == "save" {
					data.Sc.Firstflag = 1
					data.Sc.SumItemDr = 0
					data.Sc.ResSheet.Details = make([]nt.OperationOne, 1, 10)
					data.Sc.ResSheet.Sheet = *resStr
					data.Sc.ResSheet.Details[0] = *resStr
					d.Log.Debug("if resStr.status == save {")
				}
				d.DataMapId[v.IdWell] = data
				continue
			}
			if data.Sc.Firstflag == 1 {
				d.Log.Debug("if dt.itemNew.firstflag == 1 {")
				if resStr.Status == "start" {
					//len := len(dt.itemNew.resSheet.Details)
					d.Log.Debug("if resStr.status == start {")
					if data.Sc.NextTime.Flag == 0 {
						data.Sc.NextTime.Flag = 1
						data.Sc.NextTime.Start = resStr.StartData.Time
					}
					f1 := resStr.Operaton == data.Sc.ResSheet.Sheet.Operaton
					f2 := ((resStr.Operaton == d.Cfg.Operationtype[9]) && (data.Sc.ResSheet.Sheet.Operaton == d.Cfg.Operationtype[4]) || (data.Sc.ResSheet.Sheet.Operaton == d.Cfg.Operationtype[5]))
					if (f1) || (f2) {
						data.Sc.NextTime.Flag = 0
					}
					d.Log.Debug("if resStr.status == start { exit")
				}
				if resStr.Status == "save" {
					data.Sc.SumItemDr = 0
					if data.Sc.NextTime.Flag == 1 {
						data.Sc.SumItemDr = int(resStr.StopData.Time.Sub(data.Sc.NextTime.Start).Seconds())
					}
					if data.Sc.SumItemDr < d.Cfg.TimeIntervalAll {
						data.Sc.ResSheet.Details = append(data.Sc.ResSheet.Details, *resStr)
						//len := len(dt.itemNew.resSheet.Details)
						d.Log.Debug("add new Sheet.Details")
						d.DataMapId[v.IdWell] = data
						continue
					}
					len2 := len(data.Sc.ResSheet.Details)
					data.Sc.NextTime.Flag = 0
					data.Sc.ResSheet.Sheet.StopData = data.Sc.ResSheet.Details[len2-1].StopData
					d.Log.Debug("Save Sheet.Operaton - addSummaryStr(")
					// if data.Sc.ResSheet.Sheet.Status == "save"{
					data.Summarysheet = append(data.Summarysheet, *d.addSummaryStr(v.IdWell, &data.Sc.ResSheet))
					data.Sc.ResSheet.Sheet = *resStr
					data.Sc.ResSheet.Details = nil
					data.Sc.ResSheet.Details = make([]nt.OperationOne, 1, 10)
					data.Sc.ResSheet.Details[0] = *resStr
					//len3 := len(dt.itemNew.resSheet.Details)
					d.Log.Debug("Start new Sheet.Operaton -before  addSummaryStr(")
				}
			}
			d.DataMapId[v.IdWell] = data
		}
		// len2 := len(data.Sc.ResSheet.Details)
		// dt.Data.Log.Debug("done and save operation")
		// data.Sc.ResSheet.Sheet.StopData = data.Sc.ResSheet.Details[len2-1].StopData
		// dt.addSummaryStr(&dt.itemNew.ResSheet)
		d.Log.Debug("exit func DetermineSummarys.Read(close chanel)")
		localdone <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		{
			d.Log.Info("exit to DetermineSummarys(context cancel)")
			return
		}
	case <-localdone:
		{
			d.Log.Info("exit to DetermineSummarys(<-localdone)")
			return
		}
	}
}
func (d *DetermineSummarys) addSummaryStr(keyid string, p *nt.SummarysheetT) *nt.SummarysheetT {
	rs := nt.SummarysheetT{Sheet: p.Sheet}
	rs.Details = p.Details[0:len(p.Details)]
	data := rs.Sheet
	switch data.Operaton {
	case "Бурение", "Бурение ротор", "Бурение (слайд)":
		rs.Sheet.Params =
			fmt.Sprintf(" в инт. %.1f - %.1fм (Р=%.1fатм,Q=%.1fл/с,W=%.1fт)",
				data.StartData.Values[3], data.StopData.Values[3],
				data.Agv.Values[4], data.Agv.Values[5], data.Agv.Values[6])
	case "Наращивание":
		rs.Sheet.Params = fmt.Sprintf(" %.1fсв.", data.StopData.Values[10])
	case "Промывка", "Проработка":
		rs.Sheet.Params =
			fmt.Sprintf(" в инт. %.1f - %.1fм (Р=%.1fатм,Q=%.1fл/с)",
				data.StartData.Values[3], data.StopData.Values[3], data.Agv.Values[4], data.Agv.Values[5])
	case "Подъем", "Спуск":
		rs.Sheet.Params =
			fmt.Sprintf(" в инт.%.1f - %.1fм", data.StartData.Values[3], data.StopData.Values[3])
	}
	// d.DataMapId[keyid].Summarysheet = append(dt.Data.Summarysheet, rs)
	return &rs

}
