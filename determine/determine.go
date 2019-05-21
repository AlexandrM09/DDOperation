package determine

/*func (d *drillData) read() error{
    return nil
}
*/

import (
	"encoding/json"
	"os"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)
// num interface check
var CheckInt = [10]int{0, 0, 1, 0, 0, 0, 0, 0, 0, 2}

type (
	listoperation []string
	determineOne  interface {
		Check(d *DrillDataType) int
		//Check(d *DrillDataType) (int,boll)//return num of operation, state of change operation
		// CheckOne(d *DrillDataType) int
	}
	SteamI interface {
		Read(ScapeDataCh chan ScapeDataD, DoneCh chan struct{})
	}
	Determine struct {
		wg          *sync.WaitGroup
		Data        *DrillDataType
		Steam       SteamI
		ListCheck   []determineOne
		activecheck determineOne
		waitTime    int
	}
)

func (dt *Determine) Wait() error {
	ch := make(chan struct{})
	go func(ch chan struct{}) {
		fmt.Println("!!!!!!before wg.Wait()")
		dt.wg.Wait()
		fmt.Println("!!!!!!ch <- struct{}{}")
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
	return nil
}

// Main dispath function in list
func (dt *Determine) Run() {

	var res int
	defer func() {
		dt.wg.Done()
		dt.Data.Log.WithFields(logrus.Fields{
			"logger": "LOGRUS",
		}).Info("Done")
		fmt.Println("Close ScapeDataCh")
		close(dt.Data.ScapeDataCh)
		close(dt.Data.SteamCh)
		close(dt.Data.ErrCh)
		dt.Data.DoneScapeCh <- struct{}{}

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
				fmt.Println("after read Scapedata")

				if d.ActiveOperation >= 0 {
					fmt.Println("Run ActiveCheck", fmt.Sprint(d.ActiveOperation))
					res = dt.ListCheck[CheckInt[d.ActiveOperation]].Check(dt.Data)
					fmt.Println("after Chek res= ", fmt.Sprint(res))
				} else {
					res = -1
				}

				if res == -1 {

					for i := 0; i < len(dt.ListCheck) && (res == -1); i++ {
						fmt.Println("Run Check", fmt.Sprint(i))
						res = dt.ListCheck[i].Check(dt.Data)
						fmt.Println("after Chek res= ", fmt.Sprint(res))
					}
				} // select operation
				if res == -1 {
					res = len(dt.ListCheck) - 1
				}
				fmt.Println("res= ", fmt.Sprint(res))
				fmt.Println("ActiveOperation ", fmt.Sprint(d.ActiveOperation))
				switch {
				case res == d.ActiveOperation:
					{ //addDatatooperation
						dt.addDatatooperation()
					}
				default:
					{
						fmt.Println("startnewoperation()")
						d.ActiveOperation = res
						if d.ActiveOperation >= 0 {
							//saveoperation
							fmt.Println("saveoperation()")
							dt.saveoperation()

						}
						//startnewoperation
						dt.startnewoperation()
					}
				}

				d.LastScapeData = d.ScapeData
			}
		default:

		}
	}
//	fmt.Println("Ooo")
}

// Create new List determine
func NewDetermine(ds *DrillDataType, sm SteamI) *Determine {

	return &Determine{Data: ds,
		Steam:     sm,
		ListCheck: []determineOne{&Check0{}, &Check2{}, &Check9{}},
	}
}

func (dt *Determine) addDatatooperation() {
	//dt.Data.mu.Lock()
	//defer dt.Data.mu.Unlock()
}
func (dt *Determine) startnewoperation() {

	dt.Data.mu.Lock()
	defer dt.Data.mu.Unlock()
	dt.Data.OperationList = append(dt.Data.OperationList,
		OperationOne{Operaton: dt.Data.Operationtype[dt.Data.ActiveOperation], startData: dt.Data.ScapeData})
	dt.Data.Log.WithFields(logrus.Fields{
		"logger": "LOGRUS",
		"Time":   dt.Data.ScapeData.Time.Format("2006-01-02 15:04:05"),
	}).Info("Start operation " + dt.Data.Operationtype[dt.Data.ActiveOperation])
}
func (dt *Determine) saveoperation() {
	//
	dt.Data.mu.Lock()
	defer dt.Data.mu.Unlock()
	len := len(dt.Data.OperationList)
	if len == 0 {
		return
	}
	dt.Data.OperationList[len-1].stopData = dt.Data.LastScapeData
	dt.Data.Log.WithFields(logrus.Fields{
		"logger": "LOGRUS",
		"Time":   dt.Data.ScapeData.Time.Format("2006-01-02 15:04:05"),
	}).Info("Stop and save  operation " + dt.Data.Operationtype[dt.Data.ActiveOperation])
}
func (dt *Determine) Stop() {
	dt.Data.DoneCh <- struct{}{}
}

func GetList() listoperation {
	return []string{"First operation"}
}

// Create Log
//LoadConfig - load config file
func LoadConfig(path string,cf *ConfigDt) error{
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