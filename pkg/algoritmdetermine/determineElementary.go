package algoritmdetermine

import (
	"time"

	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	"github.com/sirupsen/logrus"
)

type (

	//DrillDataType drill basic data struct
	DrillDataType struct {
		OperationList        []OperationOne
		inputDataCh          chan ScapeDataD
		ScapeFullData        bool
		LastScapeData        ScapeDataD
		ScapeData            ScapeDataD
		ActiveOperation      int
		StartActiveOperation time.Time
		Log                  *logrus.Logger
		Cfg                  *nt.ConfigDt
	}

	//OperationtypeD array drilling type operation
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
	   10- КНБК
	*/
	OperationtypeD [15]string
	//ScapeDataD time series data
	ScapeDataD struct {
		Id     string
		Time   time.Time
		Values [20]float32
		/*0-Дата Время
		  1-Время Дата
		  2=Глубина забоя
		  3=Положение долота
		  4=Давление на манифольде
		  5=Расход на входе
		  6=Нагрузка на долото
		  7=Вес на крюке
		  8=Крутящий момент на роторе
		  9=Число оборотов ротора в мин.
		  10=Число свечей
		  11=Положение долота по свечам
		*/
	}
	//OperationOne description of one operation
	OperationOne struct {
		Status                                     string
		StartData, StopData, MaxData, MinData, Agv ScapeDataD
		Lastchangeoperation                        string
		Count                                      int
		Operaton                                   string
		Params                                     string
	}
)
