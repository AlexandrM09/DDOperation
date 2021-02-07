package determine

/*func (d *drillData) read() error{
    return nil
}
*/

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type (
	listoperation []string
	determineOne  interface {
		Check(d *DrillDataType) (int, bool)
	}
	//SteamI basic interface for operations recognition
	SteamI interface {
		Read(ScapeDataCh chan ScapeDataD, DoneCh chan struct{})
	}
	//Determine basic data struct
	Determine struct {
		wg        *sync.WaitGroup
		Data      *DrillDataType
		Steam     SteamI
		ListCheck []determineOne
		//  activecheck determineOne
		waitTime int
		itemNew  struct {
			resSheet  SummarysheetT
			firstflag int
			startflag int
			startTime time.Time
			//stopTime  time.Time
			sumItemDr int
			//res       OperationOne
			next     OperationOne
			nextTime struct {
				flag  int
				start time.Time
			}
		}
	}
)

//Wait waiting for completion
func (dt *Determine) Wait() error {
	ch := make(chan struct{})
	go func(ch chan struct{}) {
		//	fmt.Println("!!!!!!before wg.Wait()")
		dt.wg.Wait()
		//	fmt.Println("!!!!!!ch <- struct{}{}")
		ch <- struct{}{}
		//fmt.Println("!!!!!!after wg.Wait()")
	}(ch)
	for {
		select {
		case <-ch:
			return nil
		case <-time.After(time.Duration(dt.waitTime) * time.Second):
			{
				//
				dt.Data.DoneCh <- struct{}{}
				return errors.New("time limit exceeded")

			}
		}
	}
}

//Start - start loop
func (dt *Determine) Start(wt int) error {
	dt.waitTime = wt
	dt.wg = &sync.WaitGroup{}
	dt.wg.Add(1)
	//var err error
	//l.Println("Start determine")
	dt.Data.Log.WithFields(logrus.Fields{
		"package":  "determine",
		"function": "Start",
		//	"error":    nil,

	}).Debug(" Start Steam ")
	dt.Data.mu = &sync.RWMutex{}
	go dt.Run()
	//l.Println("Start determine")
	dt.Data.Log.WithFields(logrus.Fields{
		"package":  "determine",
		"function": "Start",
		//	"error":    nil,

	}).Debug(" Start Steam ")
	go dt.Steam.Read(dt.Data.ScapeDataCh, dt.Data.DoneCh)
	go dt.Summarysheet()
	return nil
}

//Summarysheet - fills in the summary sheet
func (dt *Determine) Summarysheet() {
	var resStr OperationOne
	for {
		select {
		case <-dt.Data.DoneSummary:
			{
				//	if dt.itemNew.flag == 1 {
				//		dt.saveSummaryStr(&dt.itemNew.res)
				//	}
				len2 := len(dt.itemNew.resSheet.Details)
				//fmt.Println("save op=", dt.itemNew.resSheet.Sheet.Operaton)
				//l.Println("save op=", dt.itemNew.resSheet.Sheet.Operaton)
				dt.Data.Log.WithFields(logrus.Fields{
					"package":  "determine",
					"function": "Summarysheet",
					//	"error":    nil,
					"status":    resStr.status,
					"firstflag": dt.itemNew.firstflag,
				}).Debug("done and save operation")
				//dt.itemNew.nextTime.flag = 0
				dt.itemNew.resSheet.Sheet.StopData = dt.itemNew.resSheet.Details[len2-1].StopData
				dt.addSummaryStr(&dt.itemNew.resSheet)
				//fmt.Println(dt.itemNew.resSheet.Sheet.Operaton)
				return
			}
		case resStr = <-dt.Data.steamCh:
			{
				dt.Data.Log.WithFields(logrus.Fields{
					"package":  "determine",
					"function": "Summarysheet",
					//	"error":    nil,
					"status":    resStr.status,
					"firstflag": dt.itemNew.firstflag,
				}).Debug("case resStr = <-dt.Data.steamCh:")
				if dt.itemNew.firstflag == 0 {
					//l.Printf("Summarysheet() Start firstflag == 0 :fistflag =%d, resStr.status =%s", dt.itemNew.firstflag, resStr.status)
					dt.Data.Log.WithFields(logrus.Fields{
						"package":  "determine",
						"function": "Summarysheet",
						//	"error":    nil,
						"status":    resStr.status,
						"firstflag": dt.itemNew.firstflag,
					}).Debug("if dt.itemNew.firstflag == 0 {")
					if resStr.status == "start" {
						//dt.itemNew.flag=1
						//	fmt.Println("firstflag == 0, resStr.status == start time=",dt.itemNew.startTime)
						//	fmt.Println("newData=",resStr.Operaton," resSheet epty")
						dt.itemNew.startTime = resStr.StartData.Time
						dt.Data.Log.WithFields(logrus.Fields{
							"package":  "determine",
							"function": "Summarysheet",
							//	"error":    nil,
							"status":    resStr.status,
							"firstflag": dt.itemNew.firstflag,
						}).Debug("if resStr.status == start")
						//l.Printf("Summarysheet():fistflag =%d, resStr.status =%s", dt.itemNew.firstflag, resStr.status)
						//dt.itemNew.res=resStr
					}
					if resStr.status == "save" {
						//	fmt.Println("firstflag == 0, resStr.status == save ")

						dt.itemNew.firstflag = 1
						dt.itemNew.sumItemDr = 0
						dt.itemNew.resSheet.Details = make([]OperationOne, 1, 10)
						dt.itemNew.resSheet.Sheet = resStr
						//	dt.itemNew.resSheet.Sheet.StartData.Time = dt.itemNew.startTime
						dt.itemNew.resSheet.Details[0] = resStr
						dt.Data.Log.WithFields(logrus.Fields{
							"package":  "determine",
							"function": "Summarysheet",
							//	"error":    nil,
							"status":    resStr.status,
							"firstflag": dt.itemNew.firstflag,
						}).Debug("if resStr.status == save {")
						//l.Printf("Summarysheet():fistflag =%d, resStr.status =%s", dt.itemNew.firstflag, resStr.status)
						//	fmt.Println("newData=",resStr.Operaton," resSheet=",dt.itemNew.resSheet.Sheet.Operaton)
					}
					continue
				}
				if dt.itemNew.firstflag == 1 {
					//l.Printf("Summarysheet() Start firstflag == 1 :fistflag =%d, resStr.status =%s", dt.itemNew.firstflag, resStr.status)
					dt.Data.Log.WithFields(logrus.Fields{
						"package":  "determine",
						"function": "Summarysheet",
						//	"error":    nil,
						"status":    resStr.status,
						"firstflag": dt.itemNew.firstflag,
					}).Debug("if dt.itemNew.firstflag == 1 {")
					if resStr.status == "start" {
						//l.Printf("Summarysheet():fistflag =%d, resStr.status =%s", dt.itemNew.firstflag, resStr.status)
						//l.Println("newData=", resStr.Operaton, " resSheet=", dt.itemNew.resSheet.Sheet.Operaton)
						len := len(dt.itemNew.resSheet.Details)
						//l.Println("len Details =", len, " last op details=", dt.itemNew.resSheet.Details[len-1].Operaton)
						dt.Data.Log.WithFields(logrus.Fields{
							"package":  "determine",
							"function": "Summarysheet",
							//	"error":    nil,
							"status":           resStr.status,
							"firstflag":        dt.itemNew.firstflag,
							"resSheet=":        dt.itemNew.resSheet.Sheet.Operaton,
							"len Details":      len,
							"last op details=": dt.itemNew.resSheet.Details[len-1].Operaton,
						}).Debug("if resStr.status == start {")
						if dt.itemNew.nextTime.flag == 0 {
							dt.itemNew.nextTime.flag = 1
							dt.itemNew.nextTime.start = resStr.StartData.Time
						}
						//l.Println("dt.itemNew.nextTime.flag = 0 resStr=", resStr.Operaton, " dt.Operaton=", dt.itemNew.resSheet.Sheet.Operaton)
						f1 := resStr.Operaton == dt.itemNew.resSheet.Sheet.Operaton
						f2 := ((resStr.Operaton == dt.Data.Cfg.Operationtype[9]) && (dt.itemNew.resSheet.Sheet.Operaton == dt.Data.Cfg.Operationtype[4]) || (dt.itemNew.resSheet.Sheet.Operaton == dt.Data.Cfg.Operationtype[5]))
						if (f1) || (f2) {

							dt.itemNew.nextTime.flag = 0
						}
						//l.Println("dt.itemNew.nextTime.flag=", dt.itemNew.nextTime.flag)
						dt.Data.Log.WithFields(logrus.Fields{
							"package":  "determine",
							"function": "Summarysheet",
							//	"error":    nil,
							"status":                resStr.status,
							"firstflag":             dt.itemNew.firstflag,
							"itemNew.nextTime.flag": dt.itemNew.nextTime.flag,
						}).Debug("if resStr.status == start { exit")
					}
					if resStr.status == "save" {
						dt.itemNew.sumItemDr = 0
						if dt.itemNew.nextTime.flag == 1 {
							dt.itemNew.sumItemDr = int(resStr.StopData.Time.Sub(dt.itemNew.nextTime.start).Seconds())
						}
						//l.Printf("Scape time=%s \n", dt.Data.ScapeData.Time.Format("15:04:05"))
						//l.Printf("Summarysheet():fistflag =%d, resStr.status =%s,dur=%s ", dt.itemNew.firstflag, resStr.status, strconv.Itoa(dt.itemNew.sumItemDr))
						//l.Println("fistflag == 1, resStr.status == save, dur= ", strconv.Itoa(dt.itemNew.sumItemDr))
						//l.Println("TimeInterval for save=", strconv.Itoa(dt.Data.Cfg.TimeIntervalAll))
						//l.Println("newData=", resStr.Operaton, " resSheet=", dt.itemNew.resSheet.Sheet.Operaton)

						if dt.itemNew.sumItemDr < dt.Data.Cfg.TimeIntervalAll {
							dt.itemNew.resSheet.Details = append(dt.itemNew.resSheet.Details, resStr)
							len := len(dt.itemNew.resSheet.Details)
							//l.Println("save Details, len Details =", len, " last op details=", dt.itemNew.resSheet.Details[len-1].Operaton)
							dt.Data.Log.WithFields(logrus.Fields{
								"package":  "determine",
								"function": "Summarysheet",
								//	"error":    nil,
								"status":                  resStr.status,
								"firstflag":               dt.itemNew.firstflag,
								"len Details":             len,
								" last operation details": dt.itemNew.resSheet.Details[len-1].Operaton,
							}).Debug("add new Sheet.Details")
							continue
						}
						//append new sheet
						len2 := len(dt.itemNew.resSheet.Details)
						//l.Println("save op=", dt.itemNew.resSheet.Sheet.Operaton)
						dt.itemNew.nextTime.flag = 0
						dt.itemNew.resSheet.Sheet.StopData = dt.itemNew.resSheet.Details[len2-1].StopData
						dt.Data.Log.WithFields(logrus.Fields{
							"package":  "determine",
							"function": "Summarysheet",
							//	"error":    nil,
							"status":                 resStr.status,
							"firstflag":              dt.itemNew.firstflag,
							"Scape time":             dt.Data.ScapeData.Time.Format("15:04:05"),
							"dur":                    strconv.Itoa(dt.itemNew.sumItemDr),
							"TimeInterval for save=": strconv.Itoa(dt.Data.Cfg.TimeIntervalAll),
							"newData":                resStr.Operaton,
							"resSheet":               dt.itemNew.resSheet.Sheet.Operaton,
						}).Debug("Save Sheet.Operaton - addSummaryStr(")
						dt.addSummaryStr(&dt.itemNew.resSheet)
						dt.itemNew.resSheet.Sheet = resStr
						dt.itemNew.resSheet.Details = nil
						dt.itemNew.resSheet.Details = make([]OperationOne, 1, 10)
						dt.itemNew.resSheet.Details[0] = resStr

						//l.Println("after save time=", dt.itemNew.startTime)
						len3 := len(dt.itemNew.resSheet.Details)
						//l.Println(" resSheet=", dt.itemNew.resSheet.Sheet.Operaton)
						//l.Println("first Details, len Details =", len3, " last op details=", dt.itemNew.resSheet.Details[len3-1].Operaton)
						dt.Data.Log.WithFields(logrus.Fields{
							"package":  "determine",
							"function": "Summarysheet",
							//	"error":    nil,
							"status":                  resStr.status,
							"firstflag":               dt.itemNew.firstflag,
							"new start time:":         dt.itemNew.startTime,
							"resSheet":                dt.itemNew.resSheet.Sheet.Operaton,
							"len Details":             len3,
							"last operation details=": dt.itemNew.resSheet.Details[len3-1].Operaton,
						}).Debug("Start new Sheet.Operaton -before  addSummaryStr(")
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

		//l.Println("Done")
		dt.Data.Log.WithFields(logrus.Fields{
			"package":  "determine",
			"function": "run",
			//	"error":    nil,

		}).Debug(" Done (defer)")
		//fmt.Println("Close ScapeDataCh")
		close(dt.Data.ScapeDataCh)
		close(dt.Data.steamCh)
		close(dt.Data.ErrCh)
		dt.wg.Done()

	}()

	for {
		d := dt.Data
		//resSt = ""
		select {
		case <-d.DoneCh:
			{ //close all
				//	fmt.Println("<-d.DoneCh dt.saveoperation() ")
				dt.saveoperation()
				//	fmt.Println("dt.Data.DoneSummary <- struct{}{} ")
				//	l.Println("dt.Data.DoneSummary <- struct{}{}")
				dt.Data.DoneSummary <- struct{}{}

				return
			}
		//case err := <-d.ErrCh:{return}

		case d.ScapeData = <-d.ScapeDataCh:
			{
				//	fmt.Println("after read Scapedata")

				if d.ActiveOperation >= 0 {
					//		fmt.Println("Run ActiveCheck", fmt.Sprint(d.ActiveOperation))
					res, changeOp = dt.ListCheck[checkInt[d.ActiveOperation]].Check(dt.Data)
					//		fmt.Println("after Chek res= ", fmt.Sprint(res))
				} else {
					res = -1
				}

				if res == -1 {

					for i := 0; i < len(dt.ListCheck) && (res == -1); i++ {
						//			fmt.Println("Run Check", fmt.Sprint(i))
						res, changeOp = dt.ListCheck[i].Check(dt.Data)
						//			fmt.Println("after Chek res= ", fmt.Sprint(res))
					}
				} // select operation
				if res == -1 {
					res = len(dt.ListCheck) - 1
					changeOp = false
				}
				//	fmt.Println("res= ", fmt.Sprint(res))
				//	fmt.Println("ActiveOperation ", fmt.Sprint(d.ActiveOperation))

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
						//		fmt.Println("startnewoperation()")
						//l.Printf("d.ActiveOperation = res=%v", res)
						dt.Data.Log.WithFields(logrus.Fields{
							"package":  "determine",
							"function": "run",
							//	"error":    nil,
							"ActiveOperation": dt.Data.ActiveOperation,
							"res":             res,
						}).Debug(" after d.ActiveOperation = res")
						d.ActiveOperation = res
						if changeOp {
							dt.addDatatooperation(0)
						}
						//if d.ActiveOperation >= 0 {
						if !changeOp {
							dt.startnewoperation()
							dt.addDatatooperation(0)
						}

						//}
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
	//	fmt.Println("Ooo")
}

func (dt *Determine) addDatatooperation(flag int) {
	//dt.Data.mu.Lock()
	len := len(dt.Data.operationList)
	if len == 0 {
		return
	}
	if !(dt.Data.operationList[len-1].Operaton == dt.Data.Cfg.Operationtype[dt.Data.ActiveOperation]) {
		dt.Data.operationList[len-1].Lastchangeoperation = dt.Data.operationList[len-1].Operaton
		dt.Data.operationList[len-1].Operaton = dt.Data.Cfg.Operationtype[dt.Data.ActiveOperation]
	}

	Op := &dt.Data.operationList[len-1]
	data := &dt.Data.ScapeData
	Op.count++
	for i := 4; i < 12; i++ {
		if data.Values[i] < Op.minData.Values[i] {
			Op.minData.Values[i] = data.Values[i]
		}
		if data.Values[i] > Op.maxData.Values[i] {
			Op.maxData.Values[i] = data.Values[i]
		}
		Op.Agv.Values[i] = Op.Agv.Values[i] + data.Values[i]
		if flag == 1 {
			Op.Agv.Values[i] = Op.Agv.Values[i] / float32(Op.count)
		}

	}

	//defer dt.Data.mu.Unlock()
}

/*
func (dt *Determine) changeOperation() {
	dt.Data.mu.Lock()
	defer dt.Data.mu.Unlock()

}
*/
func (dt *Determine) startnewoperation() {

	dt.Data.mu.Lock()
	defer dt.Data.mu.Unlock()
	tempData := dt.Data.ScapeData
	if dt.Data.temp.FlagChangeTrip == 1 {
		dt.Data.temp.FlagChangeTrip = 0
		tempData = dt.Data.temp.LastTripData
	}
	dt.Data.operationList = append(dt.Data.operationList,
		OperationOne{Operaton: dt.Data.Cfg.Operationtype[dt.Data.ActiveOperation], StartData: tempData, status: "start"})
	dt.Data.temp.LastStartData = tempData
	dt.Data.startActiveOperation = tempData.Time
	dt.Data.steamCh <- dt.Data.operationList[len(dt.Data.operationList)-1]
	/*dt.Data.Log.WithFields(logrus.Fields{
		"logger": "LOGRUS",
		"Time":   dt.Data.ScapeData.Time.Format("2006-01-02 15:04:05"),
	}).Info("Start operation " + dt.Data.Cfg.Operationtype[dt.Data.ActiveOperation])
	*/
	//l.Printf("time=%s \n", dt.Data.ScapeData.Time.Format("15:04:05"))
	//l.Printf("Start operation " + dt.Data.Cfg.Operationtype[dt.Data.ActiveOperation] + "  ActiveOperation=" + strconv.Itoa(dt.Data.ActiveOperation))
	dt.Data.Log.WithFields(logrus.Fields{
		"package":  "determine",
		"function": "startnewoperation",
		//	"error":    nil,
		"time":            dt.Data.ScapeData.Time.Format("15:04:05"),
		"Operation":       dt.Data.Cfg.Operationtype[dt.Data.ActiveOperation],
		"ActiveOperation": strconv.Itoa(dt.Data.ActiveOperation),
	}).Debug("Start operation")
}
func (dt *Determine) saveoperation() {
	//
	dt.Data.mu.Lock()
	defer dt.Data.mu.Unlock()
	len := len(dt.Data.operationList)
	if len == 0 {
		return
	}
	if dt.Data.temp.FlagChangeTrip == 1 {
		//dt.Data.temp.FlagChangeTrip=0
		dt.Data.operationList[len-1].StopData = dt.Data.temp.LastTripData
		//l.Printf("FlagChangeTrip == 1")
		dt.Data.Log.WithFields(logrus.Fields{
			"package":  "determine",
			"function": "saveoperation",
			//	"error":    nil,
			"FlagChangeTrip": dt.Data.temp.FlagChangeTrip,
		}).Debug("FlagChangeTrip == 1")
	} else {
		dt.Data.operationList[len-1].StopData = dt.Data.LastScapeData
	}
	dt.Data.operationList[len-1].status = "save"

	dt.Data.steamCh <- dt.Data.operationList[len-1]
	/*dt.Data.Log.WithFields(logrus.Fields{
		"logger": "LOGRUS",
		"Time":   dt.Data.ScapeData.Time.Format("2006-01-02 15:04:05"),
	}).Info("Stop and save  operation " + dt.Data.Cfg.Operationtype[dt.Data.ActiveOperation])
	*/
	//l.Printf("Temp time=%s \n", dt.Data.operationList[len-1].StopData.Time.Format("15:04:05"))
	//l.Printf("Stop and save  operation " + dt.Data.Cfg.Operationtype[dt.Data.ActiveOperation] +
	//	"  ActiveOperation=" + strconv.Itoa(dt.Data.ActiveOperation))
	dt.Data.Log.WithFields(logrus.Fields{
		"package":  "determine",
		"function": "saveoperation",
		//	"error":    nil,
		"Temp time":       dt.Data.operationList[len-1].StopData.Time.Format("15:04:05"),
		"Save operation ": dt.Data.Cfg.Operationtype[dt.Data.ActiveOperation],
		"ActiveOperation": strconv.Itoa(dt.Data.ActiveOperation),
	}).Debug("Stop and save  operation")
}
func (dt *Determine) addSummaryStr(p *SummarysheetT) {
	if p.Sheet.status == "save" {
		//ss:=SummarysheetT{Sheet:p,Details:make([]OperationOne,1)}
		//ss.Details=append(ss.Details,)
		rs := SummarysheetT{Sheet: p.Sheet}
		rs.Details = p.Details[0:len(p.Details)]
		data := rs.Sheet
		switch data.Operaton {
		case "Бурение", "Бурение ротор", "Бурение (слайд)":
			rs.Sheet.Params =
				fmt.Sprintf(" в инт. %.1f - %.1fм (Р=%.1fатм,Q=%.1fл/с,W=%.1fт) ",
					data.StartData.Values[3], data.StopData.Values[3],
					data.Agv.Values[4], data.Agv.Values[5], data.Agv.Values[6])
		case "Наращивание":
			rs.Sheet.Params = fmt.Sprintf(" %.1fсв.", data.StopData.Values[10])
		case "Промывка", "Проработка":
			rs.Sheet.Params =
				fmt.Sprintf(" в инт. %.1f - %.1f м(Р=%.1fатм,Q=%.1fл/с) ",
					data.StartData.Values[3], data.StopData.Values[3], data.Agv.Values[4], data.Agv.Values[5])
		case "Подъем", "Спуск":
			rs.Sheet.Params =
				fmt.Sprintf(" в инт. %.1f - %.1fм ", data.StartData.Values[3], data.StopData.Values[3])
		}
		//fmt.Println("Save item Summarysheet ",rs.Details)
		dt.Data.summarysheet = append(dt.Data.summarysheet, rs)
	}
	return
}
func (dt *Determine) saveSummaryStr(p *OperationOne) {
	return
}

//GetSummarysheet - return summary sheet
func (dt *Determine) GetSummarysheet() []SummarysheetT {
	return dt.Data.summarysheet
}

//GetOperationList - return operation list
func (dt *Determine) GetOperationList() []OperationOne {
	return dt.Data.operationList
}

//Stop stoping loop
func (dt *Determine) Stop() {
	dt.Data.DoneCh <- struct{}{}
}

//LoadConfig - load config file json
func LoadConfig(path string, cf *ConfigDt) error {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	//json.Unmarshal()
	err = decoder.Decode(&cf)
	//fmt.Printf("Cfg=%v \n",cf)
	return nil
}

//LoadConfigYaml - load config file yaml
func LoadConfigYaml(path string, cf *ConfigDt) error {
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
func NewDetermine(ds *DrillDataType, sm SteamI) *Determine {
	ds.operationList = make([]OperationOne, 0, 500)
	ds.summarysheet = make([]SummarysheetT, 0, 200)
	//ds.summarysheet.Details= make([]OperationOne, 0, 10)
	ds.steamCh = make(chan OperationOne)
	ds.ScapeDataCh = make(chan ScapeDataD)
	ds.ErrCh = make(chan error, 2)
	ds.DoneCh = make(chan struct{})
	ds.DoneSummary = make(chan struct{})
	ds.ActiveOperation = 1
	//file, _ := os.OpenFile("temp.log", os.O_CREATE|os.O_WRONLY, 0666)
	//l = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
	return &Determine{Data: ds,
		Steam: sm,
		ListCheck: []determineOne{&Check0{}, &Check1{},
			&Check2{}, &Check3{}, &Check4{}, &Check5{}, &Check7{}, &Check8{}, &Check9{}, &Check10{}},
	}
}

//var l *log.Logger
