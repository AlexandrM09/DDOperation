package determine

import (
	//	_ "fmt"
	_ "fmt"
	_ "math"
	"strconv"
	"time"
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
func checkOne0(d *DrillDataType) int {
	var res int
	res = -1
	n := d.ScapeData.Values[2] - d.ScapeData.Values[3]
	//	fmt.Printf("Drill n=%v \n",n)
	//	fmt.Printf("CheckOne2(d)==%v \n",checkOne2(d))
	//	fmt.Printf("d.Cfg.DephtTool=%v \n",d.Cfg.DephtTool)
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
func (o *Check1) Check(d *DrillDataType) (int, bool) {
	//res := checkOne9(d)
	//fmt.Printf("Check1 res1=%v \n",res)
	if detCirculation(d) {
		return -1, false
	}
	nz := d.ScapeData.Values[2] - d.ScapeData.Values[3]
	if (nz > 13) && (d.ScapeData.Values[2] < 10) {
		return 9, false
	}
	if nz < d.Cfg.Avgstand {
		if d.ActiveOperation != 1 {
			return 1, false
		}
	}
	if d.ActiveOperation == 1 {
		duratOp := int(d.ScapeData.Time.Sub(d.startActiveOperation).Seconds())
			//fmt.Printf("start %v ,max %v duratOp=%v \n",d.startActiveOperation,strconv.Itoa(int(d.Cfg.TimeIntervalMaxMkconn)) ,duratOp)
		if duratOp < d.Cfg.TimeIntervalMaxMkconn {
			//fmt.Println("res=1")
			return 1, false
		}
		return 9, false
	}
	return -1, false

}

//Check - drill test condition
func (o *Check0) Check(d *DrillDataType) (int, bool) {

	return checkOne0(d), false

}

// local circulation test condition
func checkOne2(d *DrillDataType) int {
	if detCirculation(d) {
		return 2
	}
	return -1

}

//Check - circulation test condition
func (o *Check2) Check(d *DrillDataType) (int, bool) {
	var res, resplus int
	res = checkOne2(d)
	resplus = checkOne0(d)
	if resplus > -1 {
		return resplus, false
	}
	if (res == 2) && (detRotation(d)) {
		return 3, false
	}
	l.Printf("Flow out res=%v,time=%s \n", res, d.ScapeData.Time.Format("15:04:05"))
	return res, false

}

//Check - wiper trip (reapeat Check2)
func (o *Check3) Check(d *DrillDataType) (int, bool) {
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
func (o *Check4) Check(d *DrillDataType) (int, bool) {
	if (detCirculation(d)) || (d.ActiveOperation != 4) {
		return -1, false // is not making a trip
	}
	deltaDepht := d.temp.LastTripData.Values[3] - d.ScapeData.Values[3]
	l.Printf("Up time=%s \n", d.ScapeData.Time.Format("15:04:05"))
	l.Printf("Up deltaDepht=%v \n", deltaDepht)
	if deltaDepht < 0.005 {
		duratOp := int(d.ScapeData.Time.Sub(d.temp.LastTripData.Time).Seconds())
		if duratOp > d.Cfg.TimeIntervalMkTrip {
			//you need to send LastTripData
			d.temp.FlagChangeTrip = 1
			l.Printf("d.temp.FlagChangeTrip = 1 Up time=%s \n", d.ScapeData.Time.Format("15:04:05"))
			l.Printf("Up after duratOp,deltaDepht=%v,duratOp=%v \n", deltaDepht, strconv.Itoa((duratOp)))

			return 9, false
		}
		if (-deltaDepht) > float32(d.Cfg.MinLenforTrip) {
			///you need to send LastTripData
			d.temp.FlagChangeTrip = 1
			l.Printf("Up time=%s \n", d.ScapeData.Time.Format("15:04:05"))
			l.Printf("revers d.temp.FlagChangeTrip = 1 Up res=4 deltaDepht=%v \n", deltaDepht)
			return 5, false
		}
		return 4, false
	}
	if deltaDepht > 0 {
		//That is all right
		d.temp.LastTripData = d.ScapeData
		return 4, false
	}

	return -1, false
}

//Check -  making a trip (Down)
func (o *Check5) Check(d *DrillDataType) (int, bool) {
	if (detCirculation(d)) || (d.ActiveOperation != 5) {
		return -1, false
	} // is not making a trip
	deltaDepht := d.ScapeData.Values[3] - d.temp.LastTripData.Values[3]
	l.Printf("Down time=%s \n", d.ScapeData.Time.Format("15:04:05"))
	l.Printf("Down deltaDepht=%v \n", deltaDepht)
	if deltaDepht < 0.005 {
		duratOp := int(d.ScapeData.Time.Sub(d.temp.LastTripData.Time).Seconds())
		if duratOp > d.Cfg.TimeIntervalMkTrip {
			//you need to pass LastTripData
			d.temp.FlagChangeTrip = 1
			l.Printf("d.temp.FlagChangeTrip = 1 down time=%s \n", d.ScapeData.Time.Format("15:04:05"))
			l.Printf("Down after duratOp,deltaDepht=%v,duratOp=%v \n", deltaDepht, strconv.Itoa((duratOp)))

			return 9, false
		}
		if (-deltaDepht) > float32(d.Cfg.MinLenforTrip) {
			///you need to pass LastTripData
			l.Printf("Down time=%s \n", d.ScapeData.Time.Format("15:04:05"))
			l.Printf("revers d.temp.FlagChangeTrip = 1 Down res=4 deltaDepht=%v \n", deltaDepht)
			d.temp.FlagChangeTrip = 1
			return 4, false
		}
		return 5, false
	}
	if deltaDepht > 0 {
		//That is all right
		d.temp.LastTripData = d.ScapeData
		return 5, false
	}
	l.Printf("Down out time=%s \n", d.ScapeData.Time.Format("15:04:05"))
	l.Printf("Down res=%v \n", -1)
	return -1, false
}

//Check - drill rotor test condition
func (o *Check7) Check(d *DrillDataType) (int, bool) {
	return checkOne0(d), false
}

//Check - drill slide test condition
func (o *Check8) Check(d *DrillDataType) (int, bool) {
	return checkOne0(d), false
}

//Check - temp operation test condition
func (o *Check9) Check(d *DrillDataType) (int, bool) {
	res := checkOne9(d)
	l.Printf("Temp in time=%s \n", d.ScapeData.Time.Format("15:04:05"))
	l.Printf("Temp res=%v \n", res)

	if res != 9 {
		return res, false
	}

	if d.ActiveOperation == 9 {
		l.Printf("Temp time=%s \n", d.ScapeData.Time.Format("15:04:05"))
		l.Printf("Temp cand=%v \n", d.ScapeData.Values[10])
		if d.ScapeData.Values[10] == 0 {
			if d.ScapeData.Values[3] < 0.2 {

				return 9, false
			} //пзр
			if getLastOp(d) != d.Cfg.Operationtype[10] {
				return 10, false
			} //KNBK
		}
		//SPO
		l.Printf("Temp time=%s \n", d.ScapeData.Time.Format("15:04:05"))
		l.Printf("Temp start cand=%v \n", d.temp.LastStartData.Values[10])
		if d.ScapeData.Values[10] < d.temp.LastStartData.Values[10] {
			d.temp.LastTripData = d.ScapeData
			l.Printf("Temp cand=%v \n", d.ScapeData.Values[10])
			l.Printf("Temp start cand=%v \n", d.temp.LastStartData.Values[10])
			return 4, false
		}
		if d.ScapeData.Values[10] > d.temp.LastStartData.Values[10] {
			l.Printf("Temp cand=%v \n", d.ScapeData.Values[10])
			l.Printf("Temp start cand=%v \n", d.temp.LastStartData.Values[10])
			d.temp.LastTripData = d.ScapeData
			return 5, false
		}

	}
	if d.ActiveOperation != 9 {
		d.temp.LastStartData = d.ScapeData
	}
	return 9, false
}

func checkOne9(d *DrillDataType) int {
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
func (o *Check10) Check(d *DrillDataType) (int, bool) {
	if d.ScapeData.Values[3] < 0.2 {
		return 9, false
	} //пзр
	if d.ScapeData.Values[10] > 0 {
		d.temp.LastTripData = d.ScapeData
		return 5, false
	}
	duratOp := int(d.ScapeData.Time.Sub(d.temp.LastStartData.Time).Seconds())
	if d.ActiveOperation == 10 && duratOp > d.Cfg.TimeIntervalKNBK {
		return 9, false
	}
	return 10, false //KNBK

}

//
func getLastOp(d *DrillDataType) string {
	if len(d.operationList) < 2 {
		return ""
	}
	return d.operationList[len(d.operationList)-2].Operaton
}

// determination fluid flow
func detCirculation(d *DrillDataType) bool {
	if d.Cfg.PresFlowCheck == 0 {
		if d.ScapeData.Values[4] > d.Cfg.Pmin {
			return true
		}
	}
	if d.ScapeData.Values[5] > d.Cfg.Flowmin {
		return true
	}
	return false
}

//determination rotation

func detRotation(d *DrillDataType) bool {
	if d.ScapeData.Values[9] > d.Cfg.Rotationmin {
		return true
	}
	return false
}

// tracks the movement of the tool
func getMoveTrip(d *DrillDataType) (float32, float32, time.Time) {
	res := (d.ScapeData.Values[3] - d.temp.LastStartData.Values[3])
	return res, 0, time.Now()
}
