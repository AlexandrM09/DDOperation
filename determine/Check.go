package determine

/*
   0 - Бурение
   1 - Наращивание
   2 - Промывка
   3 - Проработка
   4 - Подъем
   5 - Спуск
   6 - Работа т/с
   7 - Бурение ротор
   8 - Бурение (слайд)
   9 - ПЗР
*/
type (
	Check2 struct{}
)

func (o *Check2) check(d *DrillDataType) string {
	for{
	if !d.ScapeFullData {d.OneMoreScape=<-d.ScapeDataCh}
	return ""
	}
}
func (o *Check2) getname() string {
	return "Промывка"
}
