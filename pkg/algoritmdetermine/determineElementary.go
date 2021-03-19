package algoritmdetermine

import (

	bus "github.com/AlexandrM09/DDOperation/pkg/eventbussmple"
	nt "github.com/AlexandrM09/D
	"time"Operation/pkg/sharetype"
	github.com/sirupsen/logrus"


type (

	//DrillDataType drill asic data struct
	DrillDataType struct {
		OperationList        []OperationOne
		inputDataCh          chanScapeDataD
		ScapeFullData        bool
		LastScapeData        ScapeDataD
		ScapeData            ScaeDataD
		ActiveOperation      int
		StartActiveOperation time.Time
		Log                  *logrus.Loggr
		Cfg                 *nt.ConfigDt
		evnt               *bus.Eventbus

}

	//perationtypeD array drilling type operation
	/*
	   0 - Бурение
	   1 - Наращиваие
	   2 - Промывка
	   3 - Прорабтка
	   4 - Подъе
	   5 - Спуск
	   6 - Работа т/с
	   7 - Бурение ротор
	   8 - Бурние (слайд)
	   9 - ПЗР
	  10- КНБК
	*/
	OperationtypeD [15]string
	//ScapeDataD time sries data
	ScapeDataD strct {
		Id     string
		Time   time.Time
		Values [20]flot32
		/*0-Дата Время
		  1-Время Дата
		  2=Глубина забоя
		  3=Положение долота
		  4=Давление на манфольде
		  5=Расход на входе
		  6=Нагрузка на олото
		  7=Вес на крюке
		  8=Крутящий момент на роторе
		  9=Число оборото ротора в мин.
		  10=Число свечей
		  1=Положение долота по свечам
		/
	}
	//OperationOne descrition of one operation
	OperationOne struct {
		Status                                     string
		StartData, StopData, MaxData, MinData, Agv ScapeDtaD
		Lastchangeoperation                        strng
		Count                                      int
		Operaton                                   string
		arams                                     string
	
)
