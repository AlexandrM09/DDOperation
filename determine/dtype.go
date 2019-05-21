package determine

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type (
	//DrillDataType drill basic data struct
	DrillDataType struct {
		OperationList        []OperationOne
		SteamCh              chan OperationOne
		ScapeDataCh          chan ScapeDataD
		ErrCh                chan error
		DoneCh               chan struct{}
		DoneScapeCh          chan struct{}
		ScapeFullData        bool
		LastScapeData        ScapeDataD
		ScapeData            ScapeDataD
		ActiveOperation      int
		StartActiveOperation time.Time
		//Operationtype   OperationtypeD
		Log *logrus.Logger
		mu  *sync.RWMutex
		cfg *ConfigDt
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
	*/
	OperationtypeD [15]string
	//ScapeDataD time series data
	ScapeDataD struct {
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
		startData, stopData, maxData, minData, sum, agv ScapeDataD
		//buf_count,count int;
		//buf:array [0..bufSize] of ageooscape_data;
		Operaton string
		Params   string
	}
)

/*
var DrillOperationConst = [15]string{"Бурение",
	"Наращивание",
	"Промывка",
	"Проработка",
	"Подъем",
	"Спуск",
	"Работа т/с",
	"Бурение (ротор)", "Бурение (слайд)", "ПЗР", "", "", "", "", ""}
*/

//ScapeParamtype - scape parametrs json type
type ScapeParamtype struct {
	Name  string  `json:"name,string"`
	Gid   int     `json:"gid"`
	Delta float32 `json:"delta"`
}

//ConfigDt - configuration structure json type
type ConfigDt struct {
	Pmin                     float32
	Flowmin                  float32
	Rotationmin              float32
	PresFlowCheck            int
	DephtTool                float32
	RotorSl                  int
	DirectionalCheck         int
	BeforeDrillString        string
	ShowParamRotSl           int
	ShowParamCircl           int
	ShowParamWiper           int
	ChangeCircWiperfromDrill int
	Avgstand                 float32
	Wbitmax                  float32
	Pressmax                 float32
	TimeIntervalAll          int
	TimeIntervalMkTrip       int
	TimeIntervalMaxMkconn    int
	MinLenforTrip            int
	ScapeParam               []ScapeParamtype
	Operationtype            [15]string
}

/*
type ConfigDt struct {
	Pmin                     float32 `json:"Pmin"`
	Flowmin                  float32 `json:"Flowmin"`
	PresFlowCheck            int     `json:"PresFlowCheck"`
	DephtTool                float32 `json:"DephtTool"`
	RotorSl                  int     `json:"RotorSl"`
	DirectionalCheck         int     `json:"DirectionalCheck"`
	BeforeDrillString        string
	ShowParamRotSl           int     `json:"ShowParamRotSl"`
	ShowParamCircl           int     `json:"ShowParamCircl"`
	ShowParamWiper           int     `json:"ShowParamWiper"`
	ChangeCircWiperfromDrill int     `json:"ChangeCircWiperfromDrill"`
	Avgstand                 float32 `json:"Avgstand"`
	Wbitmax                  float32 `json:"Wbitmax"`
	Pressmax                 float32 `json:"Pressmax"`
	TimeIntervalAll          int     `json:"TimeIntervalAll"`
	TimeIntervalMkTrip       int     `json:"TimeIntervalMkTrip"`
	MinLenforTrip            int     `json:"MinLenforTrip"`
	ScapeParam               []ScapeParamtype
	Operationtype [15]string
}
*/
