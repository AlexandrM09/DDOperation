package determine

/*func (d *drillData) read() error{
    return nil
}
*/

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
	go dt.Steam.Read(dt.Data)
	//var err error
	go dt.Run()
	return nil
}

// Main dispath function in list
func (dt *Determine) Run() {

	var res int
	defer func() {
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
				d.LastScapeData = d.ScapeData
				if d.ActiveOperation >= 0 {
					res = dt.ListCheck[d.ActiveOperation].Check(dt.Data)
				} else {
					res = -1
				}

				if res == -1 {

					for i := 0; i < len(dt.ListCheck) && (res == -1); i++ {
						res = dt.ListCheck[i].Check(dt.Data)
					}
				} // select operation
				if res == -1 {
					res = len(dt.ListCheck) - 1
				}
				switch {
				case res == d.ActiveOperation:
					{ //addDatatooperation
						dt.addDatatooperation()
					}
				default:
					{
						d.ActiveOperation = res
						if d.ActiveOperation >= 0 {
							//saveoperation
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

}
func (dt *Determine) saveoperation() {
	//
	len := len(dt.Data.OperationList)
	if len == 0 {
		return
	}
	dt.Data.OperationList[len-1].stopData = dt.Data.LastScapeData
}
func (dt *Determine) Stop() {
	dt.Data.DoneCh <- struct{}{}
}

func GetList() listoperation {
	return []string{"First operation"}
}
