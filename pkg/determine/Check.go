package determine

import (
	//	_ "fmt"
	_ "fmt"
	_ "math"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
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
		if duratOp < d.Cfg.TimeIntervalMaxMkconn {
			return 1, false
		}
		return 9, true
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
	d.Log.WithFields(logrus.Fields{
		"package":  "determine (Chek2)",
		"function": "Check",
		//	"error":    nil,
		"Flow out ": res,
		"time":      d.ScapeData.Time.Format("15:04:05"),
	}).Debug(" Check - circulation test condition ")
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
	d.Log.WithFields(logrus.Fields{
		"package":  "determine (Chek4)",
		"function": "Check",
		//	"error":    nil,
		"Up deltaDepht": deltaDepht,
		"Up time":       d.ScapeData.Time.Format("15:04:05"),
	}).Debug(" making a trip (Up) ")
	if deltaDepht < 0.005 {
		duratOp := int(d.ScapeData.Time.Sub(d.temp.LastTripData.Time).Seconds())
		if duratOp > d.Cfg.TimeIntervalMkTrip {
			//you need to send LastTripData
			d.temp.FlagChangeTrip = 1
			d.Log.WithFields(logrus.Fields{
				"package":  "determine (Chek4)",
				"function": "Check",
				//	"error":    nil,
				"Up deltaDepht": deltaDepht,
				"duratOp":       strconv.Itoa((duratOp)),
				"Up time":       d.ScapeData.Time.Format("15:04:05"),
			}).Debug("  (Up) if duratOp > d.Cfg.TimeIntervalMkTrip { ")
			return 9, false
		}
		if (-deltaDepht) > float32(d.Cfg.MinLenforTrip) {
			///you need to send LastTripData
			d.temp.FlagChangeTrip = 1
			d.Log.WithFields(logrus.Fields{
				"package":       "determine (Chek4)",
				"function":      "Check",
				"Up deltaDepht": deltaDepht,
				"duratOp":       strconv.Itoa((duratOp)),
				"Up time":       d.ScapeData.Time.Format("15:04:05"),
			}).Debug("  (Up) if (-deltaDepht) > float32(d.Cfg.MinLenforTrip) ")
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
	d.Log.WithFields(logrus.Fields{
		"package":  "determine (Chek5)",
		"function": "Check",
		//	"error":    nil,
		"Down deltaDepht": deltaDepht,
		"Down time":       d.ScapeData.Time.Format("15:04:05"),
	}).Debug(" making a trip (Down) ")
	if deltaDepht < 0.005 {
		duratOp := int(d.ScapeData.Time.Sub(d.temp.LastTripData.Time).Seconds())
		if duratOp > d.Cfg.TimeIntervalMkTrip {
			//you need to pass LastTripData
			d.temp.FlagChangeTrip = 1
			d.Log.WithFields(logrus.Fields{
				"package":  "determine (Chek5)",
				"function": "Check",
				//	"error":    nil,
				"Down deltaDepht": deltaDepht,
				"Down time":       d.ScapeData.Time.Format("15:04:05"),
				"duratOp":         strconv.Itoa((duratOp)),
			}).Debug(" (Down) if duratOp > d.Cfg.TimeIntervalMkTrip {")
			return 9, false
		}
		if (-deltaDepht) > float32(d.Cfg.MinLenforTrip) {
			///you need to pass LastTripData
			d.Log.WithFields(logrus.Fields{
				"package":  "determine (Chek5)",
				"function": "Check",
				//	"error":    nil,
				"Down deltaDepht": deltaDepht,
				"Down time":       d.ScapeData.Time.Format("15:04:05"),
				"duratOp":         strconv.Itoa((duratOp)),
			}).Debug(" (Down) if (-deltaDepht) > float32(d.Cfg.MinLenforTrip) {")
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
	d.Log.WithFields(logrus.Fields{
		"package":  "determine (Chek5)",
		"function": "Check",
		//	"error":    nil,
		//"Down deltaDepht": deltaDepht,
		"Down time": d.ScapeData.Time.Format("15:04:05"),
		"res":       -1,
		//"duratOp":         strconv.Itoa((duratOp)),
	}).Debug(" return -1, false")
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
	d.Log.WithFields(logrus.Fields{
		"package":   "determine (Chek9)",
		"function":  "Check",
		"Down time": d.ScapeData.Time.Format("15:04:05"),
		"res":       res,
	}).Debug(" temp operation test condition")
	if res != 9 {
		return res, false
	}

	if d.ActiveOperation == 9 {

		d.Log.WithFields(logrus.Fields{
			"package":  "determine (Chek9)",
			"function": "Check",
			//	"error":    nil,
			"Temp cand":       d.ScapeData.Values[10],
			"Temp start cand": d.temp.LastStartData.Values[10],
		}).Debug(" if d.ActiveOperation == 9 { ")
		if d.ScapeData.Values[10] == 0 {
			d.Log.WithFields(logrus.Fields{
				"package":  "determine (Chek9)",
				"function": "Check",
				//	"error":    nil,
				"Temp cand":       d.ScapeData.Values[10],
				"Temp start cand": d.temp.LastStartData.Values[10],
			}).Debug("if d.ScapeData.Values[10] == 0")
			if d.ScapeData.Values[3] < 0.2 {

				return 9, false
			} //пзр
			if getLastOp(d) != d.Cfg.Operationtype[10] {
				return 10, false
			} //KNBK
		}
		//SPO
		d.Log.WithFields(logrus.Fields{
			"package":  "determine (Chek9)",
			"function": "Check",
			//	"error":    nil,
			"Temp cand":       d.ScapeData.Values[10],
			"Temp start cand": d.temp.LastStartData.Values[10],
		}).Debug("SPO")
		nz := d.ScapeData.Values[2] - d.ScapeData.Values[3]
		if d.ScapeData.Values[10] < d.temp.LastStartData.Values[10] {
			if nz < 13 {
				return 1, true
			}
			d.temp.LastTripData = d.ScapeData
			d.Log.WithFields(logrus.Fields{
				"package":  "determine (Chek9)",
				"function": "Check",
				//	"error":    nil,
				"Temp cand":       d.ScapeData.Values[10],
				"Temp start cand": d.temp.LastStartData.Values[10],
			}).Debug("if d.ScapeData.Values[10] < d.temp.LastStartData.Values[10] {")
			duratOp := int(d.ScapeData.Time.Sub(d.startActiveOperation).Seconds())
			if duratOp < d.Cfg.TimeIntervalAll {
				return 4, true
			}
			return 4, false
		}
		if d.ScapeData.Values[10] > d.temp.LastStartData.Values[10] {
			if nz < 13 {
				return 1, true
			}
			duratOp := int(d.ScapeData.Time.Sub(d.startActiveOperation).Seconds())
			d.Log.WithFields(logrus.Fields{
				"package":  "determine (Chek9)",
				"function": "Check",
				//	"error":    nil,
				"Temp cand":       d.ScapeData.Values[10],
				"Temp start cand": d.temp.LastStartData.Values[10],
			}).Debug(" Candel now > candel last ")
			d.temp.LastTripData = d.ScapeData
			if duratOp < d.Cfg.TimeIntervalAll {
				return 5, true
			}
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