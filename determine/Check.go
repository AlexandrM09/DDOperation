package determine

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
	Check0 struct{}
	Check2 struct{}
	Check9 struct{}
)
//  
// Check Drill
func checkOne0(d *DrillDataType) int {
	var res int
	res=-1
	n:=d.ScapeData.Values[2]-d.ScapeData.Values[3]
    if (CheckOne2(d)==2)&&(n<0.05) { res=0}
	return res
	
}
func (o *Check0) Check(d *DrillDataType) int {
	
	return checkOne0(d)
	
}
//determination fluid flow
func CheckOne2(d *DrillDataType) int {
	var res int
	res=-1
	if d.ScapeData.Values[4]>10 { res=2}
	return res
	
}
func (o *Check2) Check(d *DrillDataType) int {
	var res,resplus int
	res=-1
	res=CheckOne2(d)
	resplus=checkOne0(d)
	if resplus>-1{return resplus}
	return res
	
}

//ПЗР
func (o *Check9) Check(d *DrillDataType) int {
	
	res:=checkOne0(d)
	if res>-1 {return res}
	res=CheckOne2(d)
	if res>-1 {return res}
	return 9
}
