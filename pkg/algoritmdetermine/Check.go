package algoritmdetermine

import (
	//	_ "fmt"
	_ "fmt"
	_ "math"
	"time"
	//nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
)

/*
   0 - Бурение
   1 - Наращивание
   2 - Промывка
   3 - Проработка
   4 - Подъем
   5 - Спуск
   6 - Работа т/с
   7 - Бурение (ротор)
   8 - Бурение (слайд)
   9 - ПЗР
*/
type (
	//Check0 - drill test condition
	Check0 struct{}
	//Check1 -making a connection
	Check1 struct{}
	//Check2 - circulation test condition
	Check2 struct{}
	//Check3 - wiper trip (reapeat Check2)
	Check3 struct{}
	//Check4 -  making a trip (Up)
	Check4 struct{}
	//Check5 -  making a trip (Down)
	Check5 struct{}
	//Check7 - drill rotor
	Check7 struct{}
	//Check8 - drill slide test condition
	Check8 struct{}
	//Check9 - temp operation test condition
	Check9 struct{}
	//Check10 - KNBK
	Check10 struct{}
)

//
// Check Drill
func checkOne0(d *DetermineElementary) int {
	var res int
	res = -1
	n := d.data.ScapeData.Values[2] - d.data.ScapeData.Values[3]
	if (checkOne2(d) == 2) && (n < d.Cfg.DephtTool) {
		if d.Cfg.RotorSl > 0 {
			if detRotation(d) {
				return 7
			}
			return 8
		}
		res = 0
	}
	return res

}

//Check -making a connection
func (o *Check1) Check(d *DetermineElementary) (int, bool) {
	if detCirculation(d) {
		return -1, false
	}
	nz := d.data.ScapeData.Values[2] - d.data.ScapeData.Values[3]
	if (nz > 13) && (d.data.ScapeData.Values[2] < 10) {
		return 9, false
	}
	if nz < d.Cfg.Avgstand {
		if d.data.ActiveOperation != 1 {
			return 1, false
		}
	}
	if d.data.ActiveOperation == 1 {
		duratOp := int(d.data.ScapeData.Time.Sub(d.data.StartActiveOperation).Seconds())
		if duratOp < d.Cfg.TimeIntervalMaxMkconn {
			return 1, false
		}
		return 9, true
	}
	return -1, false

}

//Check - drill test condition
func (o *Check0) Check(d *DetermineElementary) (int, bool) {
	return checkOne0(d), false

}

// local circulation test condition
func checkOne2(d *DetermineElementary) int {
	if detCirculation(d) {
		return 2
	}
	return -1

}

//Check - circulation test condition
func (o *Check2) Check(d *DetermineElementary) (int, bool) {
	var res, resplus int
	res = checkOne2(d)
	resplus = checkOne0(d)
	if resplus > -1 {
		return resplus, false
	}
	if (res == 2) && (detRotation(d)) {
		return 3, false
	}
	d.Log.Debug(" Check - circulation test condition ")
	return res, false

}

//Check - wiper trip (reapeat Check2)
func (o *Check3) Check(d *DetermineElementary) (int, bool) {
	var res, resplus int
	res = -1
	res = checkOne2(d)
	resplus = checkOne0(d)
	if resplus > -1 {
		return resplus, false
	}
	if (res == 2) && detRotation(d) {
		return 3, false
	}
	return res, false
}

//Check -  making a trip (Up)
func (o *Check4) Check(d *DetermineElementary) (int, bool) {
	if (detCirculation(d)) || (d.data.ActiveOperation != 4) {
		return -1, false // is not making a trip
	}
	deltaDepht := d.data.Temp.LastTripData.Values[3] - d.data.ScapeData.Values[3]
	d.Log.Debug(" making a trip (Up) ")
	if deltaDepht < 0.005 {
		duratOp := int(d.data.ScapeData.Time.Sub(d.data.Temp.LastTripData.Time).Seconds())
		if duratOp > d.Cfg.TimeIntervalMkTrip {
			//you need to send LastTripData
			d.data.Temp.FlagChangeTrip = 1
			d.Log.Debug("  (Up) if duratOp > d.Cfg.TimeIntervalMkTrip { ")
			return 9, false
		}
		if (-deltaDepht) > float32(d.Cfg.MinLenforTrip) {
			///you need to send LastTripData
			d.data.Temp.FlagChangeTrip = 1
			d.Log.Debug("  (Up) if (-deltaDepht) > float32(d.Cfg.MinLenforTrip) ")
			return 5, false
		}
		return 4, false
	}
	if deltaDepht > 0 {
		//That is all right
		d.data.Temp.LastTripData = d.data.ScapeData
		return 4, false
	}

	return -1, false
}

//Check -  making a trip (Down)
func (o *Check5) Check(d *DetermineElementary) (int, bool) {
	if (detCirculation(d)) || (d.data.ActiveOperation != 5) {
		return -1, false
	} // is not making a trip
	deltaDepht := d.data.ScapeData.Values[3] - d.data.Temp.LastTripData.Values[3]
	d.Log.Debug(" making a trip (Down) ")
	if deltaDepht < 0.005 {
		duratOp := int(d.data.ScapeData.Time.Sub(d.data.Temp.LastTripData.Time).Seconds())
		if duratOp > d.Cfg.TimeIntervalMkTrip {
			//you need to pass LastTripData
			d.data.Temp.FlagChangeTrip = 1
			d.Log.Debug(" (Down) if duratOp > d.Cfg.TimeIntervalMkTrip {")
			return 9, false
		}
		if (-deltaDepht) > float32(d.Cfg.MinLenforTrip) {
			///you need to pass LastTripData
			d.Log.Debug(" (Down) if (-deltaDepht) > float32(d.Cfg.MinLenforTrip) {")
			d.data.Temp.FlagChangeTrip = 1
			return 4, false
		}
		return 5, false
	}
	if deltaDepht > 0 {
		//That is all right
		d.data.Temp.LastTripData = d.data.ScapeData
		return 5, false
	}
	d.Log.Debug(" return -1, false")
	return -1, false
}

//Check - drill rotor test condition
func (o *Check7) Check(d *DetermineElementary) (int, bool) {
	return checkOne0(d), false
}

//Check - drill slide test condition
func (o *Check8) Check(d *DetermineElementary) (int, bool) {
	return checkOne0(d), false
}

//Check - temp operation test condition
func (o *Check9) Check(d *DetermineElementary) (int, bool) {
	res := checkOne9(d)
	d.Log.Debug(" temp operation test condition")
	if res != 9 {
		return res, false
	}

	if d.data.ActiveOperation == 9 {

		d.Log.Debug(" if d.ActiveOperation == 9 { ")
		if d.data.ScapeData.Values[10] == 0 {
			d.Log.Debug("if d.ScapeData.Values[10] == 0")
			if d.data.ScapeData.Values[3] < 0.2 {

				return 9, false
			} //пзр
			if getLastOp(d) != d.Cfg.Operationtype[10] {
				return 10, false
			} //KNBK
		}
		//SPO
		d.Log.Debug("SPO")
		nz := d.data.ScapeData.Values[2] - d.data.ScapeData.Values[3]
		if d.data.ScapeData.Values[10] < d.data.Temp.LastStartData.Values[10] {
			if nz < 13 {
				return 1, true
			}
			d.data.Temp.LastTripData = d.data.ScapeData
			d.Log.Debug("if d.ScapeData.Values[10] < d.temp.LastStartData.Values[10] {")
			duratOp := int(d.data.ScapeData.Time.Sub(d.data.StartActiveOperation).Seconds())
			if duratOp < d.Cfg.TimeIntervalAll {
				return 4, true
			}
			return 4, false
		}
		if d.data.ScapeData.Values[10] > d.data.Temp.LastStartData.Values[10] {
			if nz < 13 {
				return 1, true
			}
			duratOp := int(d.data.ScapeData.Time.Sub(d.data.StartActiveOperation).Seconds())
			d.Log.Debug(" Candel now > candel last ")
			d.data.Temp.LastTripData = d.data.ScapeData
			if duratOp < d.Cfg.TimeIntervalAll {
				return 5, true
			}
			return 5, false
		}

	}
	if d.data.ActiveOperation != 9 {
		d.data.Temp.LastStartData = d.data.ScapeData
	}
	return 9, false
}

func checkOne9(d *DetermineElementary) int {
	res := checkOne0(d)
	if res > -1 {
		return res
	}
	res = checkOne2(d)
	if res > -1 {
		return res
	}
	return 9
}

//Check - KNBK
func (o *Check10) Check(d *DetermineElementary) (int, bool) {
	if d.data.ScapeData.Values[3] < 0.2 {
		return 9, false
	} //пзр
	if d.data.ScapeData.Values[10] > 0 {
		d.data.Temp.LastTripData = d.data.ScapeData
		return 5, false
	}
	duratOp := int(d.data.ScapeData.Time.Sub(d.data.Temp.LastStartData.Time).Seconds())
	if d.data.ActiveOperation == 10 && duratOp > d.Cfg.TimeIntervalKNBK {
		return 9, false
	}
	return 10, false //KNBK

}

//
func getLastOp(d *DetermineElementary) string {
	if len(d.data.OperationList) < 2 {
		return ""
	}
	return d.data.OperationList[len(d.data.OperationList)-2].Operaton
}

// determination fluid flow
func detCirculation(d *DetermineElementary) bool {
	if d.Cfg.PresFlowCheck == 0 {
		if d.data.ScapeData.Values[4] > d.Cfg.Pmin {
			return true
		}
	}
	if d.data.ScapeData.Values[5] > d.Cfg.Flowmin {
		return true
	}
	return false
}

//determination rotation

func detRotation(d *DetermineElementary) bool {
	if d.data.ScapeData.Values[9] > d.Cfg.Rotationmin {
		return true
	}
	return false
}

// tracks the movement of the tool
func getMoveTrip(d *DetermineElementary) (float32, float32, time.Time) {
	res := (d.data.ScapeData.Values[3] - d.data.Temp.LastStartData.Values[3])
	return res, 0, time.Now()
}
