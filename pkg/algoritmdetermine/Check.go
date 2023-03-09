package algoritmdetermine

import (
	//	_ "fmt"

	_ "fmt"
	_ "math"
	"time"

	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
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
10 - КНБК
*/
type (
	//Check0 - drill test condition
	Check0 struct{}
	//Check1 -making a connection
	Check1 struct{}
	//Check2 - circulation test condition
	Check2 struct{}
	//Check3 - borehole reaming (reapeat Check2)
	Check3 struct{}
	//Check4 -  making a trip (Up)
	Check4 struct{}
	//Check5 -  making a trip (Down)
	Check5 struct{}
	//Check7 - drill rotor
	Check7 struct{}
	//Check8 - drill slide test condition
	Check8 struct{}
	//Check9 - temp operation test condition (unknown operation)
	Check9 struct{}
	//Check10 - BHA
	Check10 struct{}
)

// Check Drill
func checkOne0(d *DetermineElementary, data *nt.SaveDetElementary) int {
	var res int
	res = -1
	n := data.ScapeData.Values[2] - data.ScapeData.Values[3]
	if (checkOne2(d, data) == 2) && (n < d.Cfg.DephtTool) {
		if d.Cfg.RotorSl > 0 {
			if detRotation(d, data) {
				return 7
			}
			return 8
		}
		res = 0
	}
	return res

}

// Check -making a connection
func (o *Check1) Check(d *DetermineElementary, data *nt.SaveDetElementary) (int, bool) {
	if detCirculation(d, data) {
		return -1, false
	}
	nz := data.ScapeData.Values[2] - data.ScapeData.Values[3]
	if (nz > 13) && (data.ScapeData.Values[2] < 10) {
		return 9, false
	}
	if nz < d.Cfg.Avgstand {
		if data.ActiveOperation != 1 {
			return 1, false
		}
	}
	if data.ActiveOperation == 1 {
		duratOp := int(data.ScapeData.Time.Sub(data.StartActiveOperation).Seconds())
		if duratOp < d.Cfg.TimeIntervalMaxMkconn {
			return 1, false
		}
		return 9, true
	}
	return -1, false

}

// Check - drill test condition
func (o *Check0) Check(d *DetermineElementary, data *nt.SaveDetElementary) (int, bool) {
	return checkOne0(d, data), false

}

// local circulation test condition
func checkOne2(d *DetermineElementary, data *nt.SaveDetElementary) int {
	if detCirculation(d, data) {
		return 2
	}
	return -1

}

// Check - circulation test condition
func (o *Check2) Check(d *DetermineElementary, data *nt.SaveDetElementary) (int, bool) {
	var res, resplus int
	res = checkOne2(d, data)
	resplus = checkOne0(d, data)
	if resplus > -1 {
		return resplus, false
	}
	if (res == 2) && (detRotation(d, data)) {
		return 3, false
	}
	d.Log.Debugf(" Check - circulation test condition1 res=%v", res)
	return res, false

}

// Check - borehole reaming (reapeat Check2)
func (o *Check3) Check(d *DetermineElementary, data *nt.SaveDetElementary) (int, bool) {
	var res, resplus int
	res = -1
	res = checkOne2(d, data)
	resplus = checkOne0(d, data)
	if resplus > -1 {
		return resplus, false
	}
	if (res == 2) && detRotation(d, data) {
		return 3, false
	}
	return res, false
}

// Check -  making a trip (Up)
func (o *Check4) Check(d *DetermineElementary, data *nt.SaveDetElementary) (int, bool) {
	if (detCirculation(d, data)) || (data.ActiveOperation != 4) {
		return -1, false // is not making a trip
	}
	deltaDepht := data.Temp.LastTripData.Values[3] - data.ScapeData.Values[3]
	d.Log.Debug(" making a trip (Up) ")
	if deltaDepht < 0.005 {
		duratOp := int(data.ScapeData.Time.Sub(data.Temp.LastTripData.Time).Seconds())
		if duratOp > d.Cfg.TimeIntervalMkTrip {
			//you need to send LastTripData
			data.Temp.FlagChangeTrip = 1
			d.Log.Debug("  (Up) if duratOp > d.Cfg.TimeIntervalMkTrip { ")
			return 9, false
		}
		if (-deltaDepht) > float32(d.Cfg.MinLenforTrip) {
			///you need to send LastTripData
			data.Temp.FlagChangeTrip = 1
			d.Log.Debug("  (Up) if (-deltaDepht) > float32.MinLenforTrip) ")
			return 5, false
		}
		return 4, false
	}
	if deltaDepht > 0 {
		//That is all right
		data.Temp.LastTripData = data.ScapeData
		return 4, false
	}

	return -1, false
}

// Check -  making a trip (Down)
func (o *Check5) Check(d *DetermineElementary, data *nt.SaveDetElementary) (int, bool) {
	if (detCirculation(d, data)) || (data.ActiveOperation != 5) {
		return -1, false
	} // is not making a trip
	deltaDepht := data.ScapeData.Values[3] - data.Temp.LastTripData.Values[3]
	d.Log.Debug(" making a trip (Down) ")
	if deltaDepht < 0.005 {
		duratOp := int(data.ScapeData.Time.Sub(data.Temp.LastTripData.Time).Seconds())
		if duratOp > d.Cfg.TimeIntervalMkTrip {
			//you need to pass LastTripData
			data.Temp.FlagChangeTrip = 1
			d.Log.Debug(" (Down) if duratOp > d.Cfg.TimeIntervalMkTrip {")
			return 9, false
		}
		if (-deltaDepht) > float32(d.Cfg.MinLenforTrip) {
			///you need to pass LastTripData
			d.Log.Debug(" (Down) if (-deltaDepht) > float32(d.Cfg.MinLenforTrip) {")
			data.Temp.FlagChangeTrip = 1
			return 4, false
		}
		return 5, false
	}
	if deltaDepht > 0 {
		//That is all right
		data.Temp.LastTripData = data.ScapeData
		return 5, false
	}
	d.Log.Debug(" return -1, false")
	return -1, false
}

// Check - drill rotor test condition
func (o *Check7) Check(d *DetermineElementary, data *nt.SaveDetElementary) (int, bool) {
	return checkOne0(d, data), false
}

// Check - drill slide test condition
func (o *Check8) Check(d *DetermineElementary, data *nt.SaveDetElementary) (int, bool) {
	return checkOne0(d, data), false
}

// Check - temp operation test condition
func (o *Check9) Check(d *DetermineElementary, data *nt.SaveDetElementary) (int, bool) {
	res := checkOne9(d, data)
	d.Log.Debug(" temp operation test condition")
	if res != 9 {
		return res, false
	}

	if data.ActiveOperation == 9 {

		d.Log.Debug(" if d.ActiveOperation == 9 { ")
		if data.ScapeData.Values[10] == 0 {
			d.Log.Debug("if d.ScapeData.Values[10] == 0")
			if data.ScapeData.Values[3] < 0.2 {

				return 9, false
			} //пзр
			if getLastOp(d, data) != d.Cfg.Operationtype[10] {
				return 10, false
			} //KNBK
		}
		//SPO
		d.Log.Debug("SPO")
		nz := data.ScapeData.Values[2] - data.ScapeData.Values[3]
		if data.ScapeData.Values[10] < data.Temp.LastStartData.Values[10] {
			if nz < 13 {
				return 1, true
			}
			data.Temp.LastTripData = data.ScapeData
			d.Log.Debug("if d.ScapeData.Values[10] < d.temp.LastStartData.Values[10] {")
			duratOp := int(data.ScapeData.Time.Sub(data.StartActiveOperation).Seconds())
			if duratOp < d.Cfg.TimeIntervalAll {
				return 4, true
			}
			return 4, false
		}
		if data.ScapeData.Values[10] > data.Temp.LastStartData.Values[10] {
			if nz < 13 {
				return 1, true
			}
			duratOp := int(data.ScapeData.Time.Sub(data.StartActiveOperation).Seconds())
			d.Log.Debug(" Candel now > candel last ")
			data.Temp.LastTripData = data.ScapeData
			if duratOp < d.Cfg.TimeIntervalAll {
				return 5, true
			}
			return 5, false
		}

	}
	if data.ActiveOperation != 9 {
		data.Temp.LastStartData = data.ScapeData
	}
	return 9, false
}

func checkOne9(d *DetermineElementary, data *nt.SaveDetElementary) int {
	res := checkOne0(d, data)
	if res > -1 {
		return res
	}
	res = checkOne2(d, data)
	if res > -1 {
		return res
	}
	return 9
}

// Check - BHA
func (o *Check10) Check(d *DetermineElementary, data *nt.SaveDetElementary) (int, bool) {
	if data.ScapeData.Values[3] < 0.2 {
		return 9, false
	} //пзр
	if data.ScapeData.Values[10] > 0 {
		data.Temp.LastTripData = data.ScapeData
		return 5, false
	}
	duratOp := int(data.ScapeData.Time.Sub(data.Temp.LastStartData.Time).Seconds())
	if data.ActiveOperation == 10 && duratOp > d.Cfg.TimeIntervalKNBK {
		return 9, false
	}
	return 10, false //KNBK

}

func getLastOp(d *DetermineElementary, data *nt.SaveDetElementary) string {
	if len(data.OperationList) < 2 {
		return ""
	}
	return data.OperationList[len(data.OperationList)-2].Operaton
}

// determination fluid flow
func detCirculation(d *DetermineElementary, data *nt.SaveDetElementary) bool {
	// fmt.Printf("d.Cfg=%v\n", d.Cfg)
	if d.Cfg.PresFlowCheck == 0 {
		if data.ScapeData.Values[4] > d.Cfg.Pmin {
			return true
		}
	}
	if data.ScapeData.Values[5] > d.Cfg.Flowmin {
		return true
	}
	return false
}

//determination rotation

func detRotation(d *DetermineElementary, data *nt.SaveDetElementary) bool {
	return data.ScapeData.Values[9] > d.Cfg.Rotationmin
}

// tracks the movement of the tool
func getMoveTrip(d *DetermineElementary, data *nt.SaveDetElementary) (float32, float32, time.Time) {
	res := (data.ScapeData.Values[3] - data.Temp.LastStartData.Values[3])
	return res, 0, time.Now()
}
