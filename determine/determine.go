package determine

/*func (d *drillData) read() error{
    return nil
}
*/

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type (
	listoperation []string
	determineOne  interface {
		Check(d *DrillDataType) (int, bool)
		//Check(d *DrillDataType) (int,boll)//return num of operation, state of change operation
		// CheckOne(d *DrillDataType) int
	}
	//SteamI basic interface for operations recognition
	SteamI interface {
		Read(ScapeDataCh chan ScapeDataD, DoneCh chan struct{})
	}
	//Determine basic data struct
	Determine struct {
		wg          *sync.WaitGroup
		Data        *DrillDataType
		Steam       SteamI
		ListCheck   []determineOne
		activecheck determineOne
		waitTime    int
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
		fmt.Println("!!!!!!after wg.Wait()")
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
	dt.Data.Log.WithFields(logrus.Fields{
		"logger": "LOGRUS",
		//"Time" : time.Now().Format("dd.mm.yy hh:mm"),
	}).Info("Start determine")
	dt.Data.mu = &sync.RWMutex{}
	go dt.Run()
	dt.Data.Log.WithFields(logrus.Fields{
		"logger": "LOGRUS",
		//"Time" : time.Now().Format("dd.mm.yy hh:mm"),
	}).Info("Start Steam")
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
				dt.saveSummaryStr(&resStr)
				return
			}
		case resStr = <-dt.Data.steamCh:
			{
				dt.addSummaryStr(&resStr)
			}
		}

	}
}

//Run main dispath function in list
func (dt *Determine) Run() {

	var res int
	var changeOp bool
	dt.Data.ActiveOperation = -1
	defer func() {
		dt.wg.Done()
		dt.Data.Log.WithFields(logrus.Fields{
			"logger": "LOGRUS",
		}).Info("Done")
		fmt.Println("Close ScapeDataCh")
		close(dt.Data.ScapeDataCh)
		close(dt.Data.steamCh)
		close(dt.Data.ErrCh)
		dt.Data.DoneSummary <- struct{}{}

	}()

	for {
		d := dt.Data
		//resSt = ""
		select {
		case <-d.DoneCh:
			{ //close all

				dt.saveoperation()

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
						dt.addDatatooperation()
					}
				default:
					{
						//		fmt.Println("startnewoperation()")
						d.ActiveOperation = res
						//if d.ActiveOperation >= 0 {
						if !changeOp {
							//saveoperation
							//		fmt.Println("saveoperation()")
							dt.saveoperation()
							dt.startnewoperation()
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

func (dt *Determine) addDatatooperation() {
	//dt.Data.mu.Lock()
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
		OperationOne{Operaton: dt.Data.cfg.Operationtype[dt.Data.ActiveOperation], startData: tempData, status: "start"})
	dt.Data.temp.LastStartData = tempData
	dt.Data.startActiveOperation = tempData.Time
	dt.Data.steamCh <- dt.Data.operationList[len(dt.Data.operationList)-1]
	dt.Data.Log.WithFields(logrus.Fields{
		"logger": "LOGRUS",
		"Time":   dt.Data.ScapeData.Time.Format("2006-01-02 15:04:05"),
	}).Info("Start operation " + dt.Data.cfg.Operationtype[dt.Data.ActiveOperation])
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
		dt.Data.operationList[len-1].stopData = dt.Data.temp.LastTripData
	} else {
		dt.Data.operationList[len-1].stopData = dt.Data.LastScapeData
	}
	dt.Data.operationList[len-1].status = "save"
	dt.Data.steamCh <- dt.Data.operationList[len-1]
	dt.Data.Log.WithFields(logrus.Fields{
		"logger": "LOGRUS",
		"Time":   dt.Data.ScapeData.Time.Format("2006-01-02 15:04:05"),
	}).Info("Stop and save  operation " + dt.Data.cfg.Operationtype[dt.Data.ActiveOperation])
}
func (dt *Determine) addSummaryStr(p *OperationOne) {
	if p.status == "save" {
		dt.Data.summarysheet = append(dt.Data.summarysheet, *p)
	}
	return
}
func (dt *Determine) saveSummaryStr(p *OperationOne) {
	return
}

//GetSummarysheet - return summary sheet
func (dt *Determine) GetSummarysheet() []OperationOne {
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

//LoadConfig - load config file
func LoadConfig(path string, cf *ConfigDt) error {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	//json.Unmarshal()
	err = decoder.Decode(&cf)
	//fmt.Printf("cfg=%v \n",cf)
	return nil
}

// num interface check
var checkInt = [10]int{0, 1, 2, 3, 4, 5, 0, 6, 7, 8}

//NewDetermine  create new List determine
func NewDetermine(ds *DrillDataType, sm SteamI) *Determine {
	ds.operationList = make([]OperationOne, 0, 500)
	ds.summarysheet = make([]OperationOne, 0, 200)
	ds.steamCh = make(chan OperationOne)
	ds.ScapeDataCh = make(chan ScapeDataD)
	ds.ErrCh = make(chan error, 2)
	ds.DoneCh = make(chan struct{})
	ds.DoneSummary = make(chan struct{})
	ds.ActiveOperation = 1
	return &Determine{Data: ds,
		Steam: sm,
		ListCheck: []determineOne{&Check0{}, &Check1{},
			&Check2{}, &Check3{}, &Check4{}, &Check5{}, &Check7{}, &Check8{}, &Check9{}},
	}
}
