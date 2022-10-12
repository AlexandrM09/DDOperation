package determine

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	steam "github.com/AlexandrM09/DDOperation/pkg/steamd"
	"gopkg.in/yaml.v2"
)

type (
	listoperation []string
	//DetermineOne type
	determineOne interface {
		Check(d *nt.DrillDataType) (int, bool)
	}
	//Determine basic data struct
	Determine struct {
		wg        *sync.WaitGroup
		Data      *nt.DrillDataType
		Steam     steam.SteamI
		ListCheck []determineOne
		//  activecheck determineOne
		startTime time.Time
		waitTime  int
		//	Mu        *sync.RWMutex
		itemNew nt.ResultSheet
	}
)

//Wait waiting for completion
func (dt *Determine) Wait() (time.Duration, error) {
	ch := make(chan struct{})
	defer func() {
		dt.Data.Log.Debug(" Done ( Wait())")
		close(ch)
		close(dt.Data.DoneSummary)
		close(dt.Data.TimelimitCh)
		close(dt.Data.DoneCh)
		close(dt.Data.Done)
	}()

	timer1 := time.NewTimer(time.Second * time.Duration(dt.waitTime))
	go func(ch chan struct{}) {
		dt.wg.Wait()
		dt.Data.Log.Info("Wait() ch <- struct{}{} ")
		ch <- struct{}{}

	}(ch)
	for {
		select {
		case <-ch:
			{
				timer1.Stop()
				nw1 := time.Now()
				return nw1.Sub(dt.startTime), nil
			}
		case <-timer1.C:

			{
				//
				dt.Data.Log.Error(" time limit exceeded, dt.Data.Done <- struct{}{}")
				dt.Data.Done <- struct{}{}

				for {

					select {
					case <-ch:
						{
							dt.Data.Log.Error(" time limit exceeded,exit")
							timer1.Stop()
							nw1 := time.Now()
							return nw1.Sub(dt.startTime), errors.New("time limit exceeded,normal output")
						}
					default:
					}
				}
			}
		default:
		}
	}

}

//Start - start loop
func (dt *Determine) Start(wt int) (time.Duration, error) {
	dt.startTime = time.Now()
	dt.waitTime = wt

	dt.wg.Add(1)
	dt.Data.Log.Debug(" Start Steam ")
	go dt.Run()
	//l.Println("Start determine")
	dt.Data.Log.Debug(" Start Steam ")
	go dt.Steam.Read(dt.Data.ScapeDataCh, dt.Data.DoneCh, dt.Data.Done, dt.Data.ErrCh)
	go dt.Summarysheet()
	return dt.Wait() //nil
}

//Summarysheet - fills in the summary sheet
func (dt *Determine) Summarysheet() {
	var resStr nt.OperationOne
	for {
		select {
		case <-dt.Data.DoneSummary:
			{

				len2 := len(dt.itemNew.ResSheet.Details)
				dt.Data.Log.Debug("done and save operation")
				//dt.mu.Lock()
				//defer dt.mu.Unlock()
				dt.itemNew.ResSheet.Sheet.StopData = dt.itemNew.ResSheet.Details[len2-1].StopData
				dt.addSummaryStr(&dt.itemNew.ResSheet)
				dt.wg.Done()
				return
			}
		case resStr = <-dt.Data.SteamCh:
			{
				dt.Data.Log.Debugf("case resStr = <-dt.Data.steamCh:  status:%s", resStr.Status)
				if dt.itemNew.Firstflag == 0 {
					dt.Data.Log.Debug("if dt.itemNew.firstflag == 0 {")
					if resStr.Status == "start" {
						dt.itemNew.StartTime = resStr.StartData.Time
						dt.Data.Log.Debug("one!! if resStr.status == start")
					}
					if resStr.Status == "save" {
						dt.itemNew.Firstflag = 1
						dt.itemNew.SumItemDr = 0
						dt.itemNew.ResSheet.Details = make([]nt.OperationOne, 1, 10)
						dt.itemNew.ResSheet.Sheet = resStr
						dt.itemNew.ResSheet.Details[0] = resStr
						dt.Data.Log.Debug("if resStr.status == save {")
					}
					continue
				}
				if dt.itemNew.Firstflag == 1 {
					dt.Data.Log.Debug("if dt.itemNew.firstflag == 1 {")
					if resStr.Status == "start" {
						//len := len(dt.itemNew.resSheet.Details)
						dt.Data.Log.Debug("if resStr.status == start {")
						if dt.itemNew.NextTime.Flag == 0 {
							dt.itemNew.NextTime.Flag = 1
							dt.itemNew.NextTime.Start = resStr.StartData.Time
						}
						f1 := resStr.Operaton == dt.itemNew.ResSheet.Sheet.Operaton
						f2 := ((resStr.Operaton == dt.Data.Cfg.Operationtype[9]) && (dt.itemNew.ResSheet.Sheet.Operaton == dt.Data.Cfg.Operationtype[4]) || (dt.itemNew.ResSheet.Sheet.Operaton == dt.Data.Cfg.Operationtype[5]))
						if (f1) || (f2) {
							dt.itemNew.NextTime.Flag = 0
						}
						dt.Data.Log.Debug("if resStr.status == start { exit")
					}
					if resStr.Status == "save" {
						dt.itemNew.SumItemDr = 0
						if dt.itemNew.NextTime.Flag == 1 {
							dt.itemNew.SumItemDr = int(resStr.StopData.Time.Sub(dt.itemNew.NextTime.Start).Seconds())
						}
						if dt.itemNew.SumItemDr < dt.Data.Cfg.TimeIntervalAll {
							dt.itemNew.ResSheet.Details = append(dt.itemNew.ResSheet.Details, resStr)
							//len := len(dt.itemNew.resSheet.Details)
							dt.Data.Log.Debug("add new Sheet.Details")
							continue
						}
						len2 := len(dt.itemNew.ResSheet.Details)
						dt.itemNew.NextTime.Flag = 0
						dt.itemNew.ResSheet.Sheet.StopData = dt.itemNew.ResSheet.Details[len2-1].StopData
						dt.Data.Log.Debug("Save Sheet.Operaton - addSummaryStr(")
						dt.addSummaryStr(&dt.itemNew.ResSheet)
						dt.itemNew.ResSheet.Sheet = resStr
						dt.itemNew.ResSheet.Details = nil
						dt.itemNew.ResSheet.Details = make([]nt.OperationOne, 1, 10)
						dt.itemNew.ResSheet.Details[0] = resStr
						//len3 := len(dt.itemNew.resSheet.Details)
						dt.Data.Log.Debug("Start new Sheet.Operaton -before  addSummaryStr(")
					}
				}
			}
		default:
		}
	}
}

//Run main dispath function in list
func (dt *Determine) Run() {
	var res int
	var changeOp bool
	dt.Data.ActiveOperation = -1
	defer func() {
		dt.Data.Log.Debug(" Done (defer)")
		close(dt.Data.ScapeDataCh)
		close(dt.Data.SteamCh)
		close(dt.Data.ErrCh)
		dt.wg.Done()

	}()

	for {
		d := dt.Data
		select {
		case <-d.DoneCh:
			{
				dt.saveoperation()
				//dt.Data.Done <- struct{}{}
				//<-dt.Data.DoneCh
				dt.wg.Add(1)
				dt.Data.DoneSummary <- struct{}{}

				return
			}
		case d.ScapeData = <-d.ScapeDataCh:
			{
				if d.ActiveOperation >= 0 {
					res, changeOp = dt.ListCheck[checkInt[d.ActiveOperation]].Check(dt.Data)
				} else {
					res = -1
				}
				if res == -1 {
					for i := 0; i < len(dt.ListCheck) && (res == -1); i++ {
						res, changeOp = dt.ListCheck[i].Check(dt.Data)
					}
				} // select operation
				if res == -1 {
					res = len(dt.ListCheck) - 1
					changeOp = false
				}
				switch {
				case res == d.ActiveOperation:
					{ //addDatatooperation
						dt.addDatatooperation(0)
					}
				default:
					{

						if !changeOp {
							dt.addDatatooperation(1)
							dt.saveoperation()

						}
						dt.Data.Log.Debug(" after d.ActiveOperation = res")
						d.ActiveOperation = res
						if changeOp {
							dt.addDatatooperation(0)
						}
						if !changeOp {
							dt.startnewoperation()
							dt.addDatatooperation(0)
						}

						if changeOp {
							changeOp = false
						}
						//startnewoperation
					}
				}
				d.LastScapeData = d.ScapeData
			}
		default:
		}
	}
}

func (dt *Determine) addDatatooperation(flag int) {
	dt.Data.Mu.Lock()
	defer dt.Data.Mu.Unlock()
	len := len(dt.Data.OperationList)
	if len == 0 {
		return
	}
	if !(dt.Data.OperationList[len-1].Operaton == dt.Data.Cfg.Operationtype[dt.Data.ActiveOperation]) {
		dt.Data.OperationList[len-1].Lastchangeoperation = dt.Data.OperationList[len-1].Operaton
		dt.Data.OperationList[len-1].Operaton = dt.Data.Cfg.Operationtype[dt.Data.ActiveOperation]
	}
	Op := &dt.Data.OperationList[len-1]
	data := &dt.Data.ScapeData
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

//
func (dt *Determine) startnewoperation() {

	dt.Data.Mu.Lock()
	defer dt.Data.Mu.Unlock()
	tempData := dt.Data.ScapeData
	if dt.Data.Temp.FlagChangeTrip == 1 {
		dt.Data.Temp.FlagChangeTrip = 0
		tempData = dt.Data.Temp.LastTripData
	}
	dt.Data.OperationList = append(dt.Data.OperationList,
		nt.OperationOne{Operaton: dt.Data.Cfg.Operationtype[dt.Data.ActiveOperation], StartData: tempData, Status: "start"})
	dt.Data.Temp.LastStartData = tempData
	dt.Data.StartActiveOperation = tempData.Time
	dt.Data.SteamCh <- dt.Data.OperationList[len(dt.Data.OperationList)-1]
	dt.Data.Log.Debug("Start operation")
}

//
func (dt *Determine) saveoperation() {
	//
	dt.Data.Mu.Lock()
	defer dt.Data.Mu.Unlock()
	len := len(dt.Data.OperationList)
	if len == 0 {
		return
	}
	if dt.Data.Temp.FlagChangeTrip == 1 {
		//dt.Data.temp.FlagChangeTrip=0
		dt.Data.OperationList[len-1].StopData = dt.Data.Temp.LastTripData
		//l.Printf("FlagChangeTrip == 1")
		dt.Data.Log.Debug("FlagChangeTrip == 1")
	} else {
		dt.Data.OperationList[len-1].StopData = dt.Data.LastScapeData
	}
	dt.Data.OperationList[len-1].Status = "save"

	dt.Data.SteamCh <- dt.Data.OperationList[len-1]
	dt.Data.Log.Debug("Stop and save  operation ")
}
func (dt *Determine) addSummaryStr(p *nt.SummarysheetT) {
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
		dt.Data.Summarysheet = append(dt.Data.Summarysheet, rs)
	}
	return
}
func (dt *Determine) saveSummaryStr(p *nt.OperationOne) {
	return
}

//GetSummarysheet - return summary sheet
func (dt *Determine) GetSummarysheet() []nt.SummarysheetT {
	return dt.Data.Summarysheet
}

//GetOperationList - return operation list
func (dt *Determine) GetOperationList() []nt.OperationOne {
	return dt.Data.OperationList
}

//Stop stoping loop
func (dt *Determine) Stop() {
	dt.Data.DoneCh <- struct{}{}
}

//LoadConfig - load config file json
func LoadConfig(path string, cf *nt.ConfigDt) error {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	//json.Unmarshal()
	err = decoder.Decode(&cf)
	return nil
}

//LoadConfigYaml - load config file yaml
func LoadConfigYaml(path string, cf *nt.ConfigDt) error {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &cf)
	//json.Unmarshal()
	if err != nil {
		return err
	}
	return nil
}

// num interface check
var checkInt = [11]int{0, 1, 2, 3, 4, 5, 0, 6, 7, 8, 9}

//NewDetermine  create new List determine
func NewDetermine(ds *nt.DrillDataType, sm steam.SteamI) *Determine {
	ds.OperationList = make([]nt.OperationOne, 0, 500)
	ds.Summarysheet = make([]nt.SummarysheetT, 0, 200)
	ds.SteamCh = make(chan nt.OperationOne)
	ds.ScapeDataCh = make(chan nt.ScapeDataD)
	ds.ErrCh = make(chan error, 2)
	ds.DoneCh = make(chan struct{})
	ds.Done = make(chan struct{})

	ds.DoneSummary = make(chan struct{})
	ds.TimelimitCh = make(chan struct{})
	ds.ActiveOperation = 1
	ds.Mu = &sync.RWMutex{}
	return &Determine{Data: ds,
		wg: &sync.WaitGroup{},
		//	Mu:    &sync.RWMutex{},
		Steam: sm,
		ListCheck: []determineOne{&Check0{}, &Check1{},
			&Check2{}, &Check3{}, &Check4{}, &Check5{}, &Check7{}, &Check8{}, &Check9{}, &Check10{}},
	}
}

//FormatSheet format string function
func FormatSheet(sh nt.SummarysheetT) string {
	return fmt.Sprintf("%s | %s |%s%s",
		sh.Sheet.StartData.Time.Format("2006-01-02 15:04:05"),
		sh.Sheet.StopData.Time.Format("15:04:05"),
		sh.Sheet.Operaton,
		sh.Sheet.Params)
}

//FormatSheetDetails format string function
func FormatSheetDetails(Det nt.OperationOne) string {
	return fmt.Sprintf("____%s | %s |%s",
		Det.StartData.Time.Format("15:04:05"),
		Det.StopData.Time.Format("15:04:05"),
		Det.Operaton)
}

//FormatSheet2 format string function
func FormatSheet2(sh nt.SummarysheetT) string {
	tempt, _ := time.Parse("15:04:01", "00:00:00")
	dur := sh.Sheet.StopData.Time.Sub(sh.Sheet.StartData.Time)
	return fmt.Sprintf("%s | %s |%s |%s %s",
		sh.Sheet.StartData.Time.Format("15:04"),
		sh.Sheet.StopData.Time.Format("15:04"),
		tempt.Add(dur).Format("15:04"),
		sh.Sheet.Operaton,
		sh.Sheet.Params)
}
