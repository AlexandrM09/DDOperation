package algoritmdetermine

import (
	"fmt"
	"sync"
	"time"

	StoreMap "github.com/AlexandrM09/DDOperation/pkg/StoreMap"
	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	"github.com/sirupsen/logrus"
)

type (

	//SummarysheetT -type result list
	SummaryCalcT struct {
		Sheet   nt.OperationOne
		Details []nt.OperationOne

		Temp struct {
			LastToolDepht     float32
			LastTimeToolDepht time.Time
			StartDepht        float32
			LastStartData     nt.ScapeDataD
			LastTripData      nt.ScapeDataD
			FlagChangeTrip    int
		}
	}
	SummaryResult struct {
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
		In    string   //pub name in busevent
		Out   []string //pub name in busevent
		//id скважин должны быть добавлены до старта,
		// в процессе работы скважины с текущей архитектурой добавлять нeльзя
		Id   map[string]int
		done chan struct{}
	}
)

func New(log *logrus.Logger, cfg *nt.ConfigDt, store *StoreMap.Brocker, in string, out []string) *DetermineSummarys {
	return &DetermineSummarys{
		DataMapId: make(map[string]SummaryResult),
		Log:       log,
		Wg:        &sync.WaitGroup{},
		Cfg:       cfg,
		Store:     store,
		In:        in,
		Out:       out,
		// done:make(chan struct{},1),
	}
}
func (d *DetermineSummarys) AddWell(id string) {
	_, ok := d.Id[id]
	if !ok {
		// d.Id[id] = 1
		d.DataMapId[id] = SummaryResult{}

	}
}
func (d *DetermineSummarys) Stop() map[string]SummaryResult {
	defer close(d.done)
	d.done <- struct{}{}
	d.Wg.Wait()
	return d.DataMapId
}
func (d *DetermineSummarys) Run(ErrCh chan error) {
	d.done = make(chan struct{})
	DoneInside := make(chan struct{})
	go func() {
		defer close(DoneInside)
		d.Wg.Add(1)
		go d.Read(DoneInside, ErrCh)
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
func (d *DetermineSummarys) Read(DoneInside chan struct{}, ErrCh chan error) {
	defer func() {
		DoneInside <- struct{}{}
		d.Log.Infof("Exit read DetermineSummarys ")
	}()
	d.Log.Infof("Start Run DetSummary")
	for {
		for keyid, _ := range d.DataMapId {
			select {
			case <-d.done:
				{
					d.Log.Info("on-demand output DetermineSummary")
					return
				}
			default:
				{
				}
			}
			//read data
			g := d.Store.Receive(d.In, keyid) //"ScapeData"
			//	d.Log.Infof("Read ScapeData id=%s,ii2=%v", d.Id, ii2)
			if g == nil {
				continue
			}
			var resStr *nt.OperationOne
			resStr, ok := g.(*nt.OperationOne)
			if !ok {
				d.Log.Debugf("DetermineElementary:something went very wrong")
				return
			}
			data, ok := d.DataMapId[keyid]
			if !ok {
				d.Log.Debugf("DetermineElementary:something went very wrong")
				return
			}
			d.Log.Debugf("Start dSumm:  status:%s,Id:%s", resStr.Status, keyid)
			if data.Sc.Firstflag == 0 {
				//Самая первая операция в списке
				d.Log.Debug("if dt.itemNew.firstflag == 0 {")
				if resStr.Status == "start" {
					data.Sc.StartTime = resStr.StartData.Time
					d.Log.Debug("one!! if resStr.status == start")
				}
				if resStr.Status == "save" {
					data.Sc.Firstflag = 1
					data.Sc.SumItemDr = 0
					data.Sc.ResSheet.Details = make([]nt.OperationOne, 1, 10)
					data.Sc.ResSheet.Sheet = *resStr
					data.Sc.ResSheet.Details[0] = *resStr
					d.Log.Debug("if resStr.status == save {")
				}
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
						continue
					}
					len2 := len(data.Sc.ResSheet.Details)
					data.Sc.NextTime.Flag = 0
					data.Sc.ResSheet.Sheet.StopData = data.Sc.ResSheet.Details[len2-1].StopData
					d.Log.Debug("Save Sheet.Operaton - addSummaryStr(")
					if res := d.addSummaryStr(keyid, &data.Sc.ResSheet); res != nil {
						data.Summarysheet = append(data.Summarysheet, *res)
					}
					data.Sc.ResSheet.Sheet = *resStr
					data.Sc.ResSheet.Details = nil
					data.Sc.ResSheet.Details = make([]nt.OperationOne, 1, 10)
					data.Sc.ResSheet.Details[0] = *resStr
					//len3 := len(dt.itemNew.resSheet.Details)
					d.Log.Debug("Start new Sheet.Operaton -before  addSummaryStr(")
				}
			}
			d.DataMapId[keyid] = data
		}
	}
}
func (d *DetermineSummarys) addSummaryStr(keyid string, p *nt.SummarysheetT) *nt.SummarysheetT {
	//dt.mu.Lock()
	//defer dt.mu.Unlock()
	if p.Sheet.Status == "save" {
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
	return nil
}
