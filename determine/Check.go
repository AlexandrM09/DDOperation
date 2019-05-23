package determine

import (
//	_ "fmt"
	_ "time"
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
	//Check7 - drill rotor
	Check7 struct{}
	//Check8 - drill slide test condition
	Check8 struct{}
	//Check9 - temp operation test condition
	Check9 struct{}

)

//
// Check Drill
func checkOne0(d *DrillDataType) int {
	var res int
	res = -1
	n := d.ScapeData.Values[2] - d.ScapeData.Values[3]
	//	fmt.Printf("Drill n=%v \n",n)
	//	fmt.Printf("CheckOne2(d)==%v \n",checkOne2(d))
	//	fmt.Printf("d.cfg.DephtTool=%v \n",d.cfg.DephtTool)
	if (checkOne2(d) == 2) && (n < d.cfg.DephtTool) {
		if d.cfg.RotorSl>0{
			if ( detRotation(d)) {return 7}
			return 8
		}
		res = 0
	}
	return res

}

//Check -making a connection
func (o *Check1) Check(d *DrillDataType) (int, bool) {
	res := checkOne9(d)
	//fmt.Printf("Check1 res1=%v \n",res)
	if res == 9 {
		duratOp := int(d.ScapeData.Time.Sub(d.StartActiveOperation).Seconds())
		//nead check 4,5
		//	fmt.Printf("duratOp=%v, start %v \n",duratOp,d.StartActiveOperation)
		if (duratOp < d.cfg.TimeIntervalMaxMkconn) || (d.ActiveOperation == -1) {
			return 1, false
		}
		if d.ActiveOperation == 1 {
			return 9, true
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
	var res int
	res = -1
	if detCirculation(d) {
		res = 2
	}
	return res

}

//Check - circulation test condition
func (o *Check2) Check(d *DrillDataType) (int, bool) {
	var res, resplus int
	res = -1
	res = checkOne2(d)
	resplus = checkOne0(d)
	if resplus > -1 {
		return resplus, false
	}
	if ( detRotation(d)) {return 3, false}
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
	if ( detRotation(d)) {return 3, false}
	return res, false
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
	//if checkOne9(d)>-1 {return 9,false}

	return checkOne9(d), false
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

// determination fluid flow
func detCirculation(d *DrillDataType) bool {
	if d.cfg.PresFlowCheck == 0 {
		if d.ScapeData.Values[4] > d.cfg.Pmin {
			return true
		}
	}
	if d.ScapeData.Values[5] > d.cfg.Flowmin {
		return true
	}
	return false
}

//determination rotation

func detRotation(d *DrillDataType) bool{
	if d.ScapeData.Values[9]> d.cfg.Rotationmin{return true}
	return false
}
