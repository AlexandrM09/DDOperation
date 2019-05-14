package determine

/*func (d *drillData) read() error{
    return nil
}
*/

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

var CheckInt = [10]int{0,0,1, 0, 0, 0, 0, 0, 0, 2}

type (
	listoperation []string
	determineOne  interface {
		Check(d *DrillDataType) int
		// CheckOne(d *DrillDataType) int
	}
	SteamI interface {
		Read(d *DrillDataType)
	}
	Determine struct {
		Data        *DrillDataType
		Steam       SteamI
		ListCheck   []determineOne
		activecheck determineOne
	}
)

func (dt *Determine) Start() error {

	//var err error
	dt.Data.Log.WithFields(logrus.Fields{
		"logger": "LOGRUS",
		//"Time" : time.Now().Format("dd.mm.yy hh:mm"),
	}).Info("Start determine")
	go dt.Run()
	dt.Data.Log.WithFields(logrus.Fields{
		"logger": "LOGRUS",
		//"Time" : time.Now().Format("dd.mm.yy hh:mm"),
	}).Info("Start Steam")
	go dt.Steam.Read(dt.Data)
	return nil
}

// Main dispath function in list
func (dt *Determine) Run() {

	var res int
	defer func() {
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
				d.LastScapeData = d.ScapeData
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
			}
		default:

		}
	}
	fmt.Println("Ooo")
}

// Create new List determine
func NewDetermine(ds *DrillDataType, sm SteamI) *Determine {

	return &Determine{Data: ds,
		Steam:     sm,
		ListCheck: []determineOne{&Check0{}, &Check2{}, &Check9{}},
	}
}

func (dt *Determine) addDatatooperation() {
	//
}
func (dt *Determine) startnewoperation() {

	dt.Data.OperationList = append(dt.Data.OperationList,
		OperationOne{Operaton: dt.Data.Operationtype[dt.Data.ActiveOperation], startData: dt.Data.ScapeData})
	dt.Data.Log.WithFields(logrus.Fields{
		"logger": "LOGRUS",
		"Time":   dt.Data.ScapeData.Time.Format("dd.mm.yy hh:mm:ss"),
	}).Info("Start operation " + dt.Data.Operationtype[dt.Data.ActiveOperation])
}
func (dt *Determine) saveoperation() {
	//
	len := len(dt.Data.OperationList)
	if len == 0 {
		return
	}
	dt.Data.OperationList[len-1].stopData = dt.Data.LastScapeData
	dt.Data.Log.WithFields(logrus.Fields{
		"logger": "LOGRUS",
		"Time":   dt.Data.ScapeData.Time.Format("dd.mm.yy hh:mm:ss"),
	}).Info("Stop and save  operation " + dt.Data.Operationtype[dt.Data.ActiveOperation])
}
func (dt *Determine) Stop() {
	dt.Data.DoneCh <- struct{}{}
}

func GetList() listoperation {
	return []string{"First operation"}
}

// Create Log
