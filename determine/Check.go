package determine

import (
	"fmt"
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
	//Check2 - circulation test condition
	Check2 struct{}
	//Check9 - temp operation test condition
	Check9 struct{}
)
//  
// Check Drill
func checkOne0(d *DrillDataType) int {
	var res int
	res=-1
	n:=d.ScapeData.Values[2]-d.ScapeData.Values[3]
	fmt.Printf("Drill n=%v \n",n)
	fmt.Printf("CheckOne2(d)==%v \n",checkOne2(d))
	fmt.Printf("d.cfg.DephtTool=%v \n",d.cfg.DephtTool)
    if (checkOne2(d)==2)&&(n<d.cfg.DephtTool) { res=0}
	return res
	
}
//Check - drill test condition
func (o *Check0) Check(d *DrillDataType) (int,bool) {
	
	return checkOne0(d),false
	
}
// local circulation test condition
func checkOne2(d *DrillDataType) int {
	var res int
	res=-1
	if detCirculation(d) { res=2}
	return res
	
}
//Check - circulation test condition
func (o *Check2) Check(d *DrillDataType) (int,bool){
	var res,resplus int
	res=-1
	res=checkOne2(d)
	resplus=checkOne0(d)
	if resplus>-1{return resplus,false}
	return res,false
	
}

//Check - temp operation test condition
func (o *Check9) Check(d *DrillDataType) (int,bool) {
	
	res:=checkOne0(d)
	if res>-1 {return res,false}
	res=checkOne2(d)
	if res>-1 {return res,false}
	return 9,false
}
// determination fluid flow
func detCirculation(d *DrillDataType) bool{
	if d.cfg.PresFlowCheck==0{
		if d.ScapeData.Values[4]> d.cfg.Pmin { return true}
	}
	if d.ScapeData.Values[5]> d.cfg.Flowmin{ return true}
	return false
} 
//determination rotation
/*
func detRotation(d *DrillDataType) bool{
	if d.ScapeData.Values[9]> d.cfg.Rotationmin{return true}
	return false
}
*/