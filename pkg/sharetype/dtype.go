package sharetype

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	_ "gopkg.in/yaml.v2"
)

type (

	//DrillDataType drill basic data struct
	DrillDataType struct {
		OperationList []OperationOne
		Summarysheet  []SummarysheetT

		SteamCh         chan OperationOne
		ScapeDataCh     chan ScapeDataD
		ErrCh           chan error
		DoneCh          chan struct{}
		Done            chan struct{}
		TimelimitCh     chan struct{}
		DoneSummary     chan struct{}
		ScapeFullData   bool
		LastScapeData   ScapeDataD
		ScapeData       ScapeDataD
		ActiveOperation int
		Temp            struct {
			LastToolDepht     float32
			LastTimeToolDepht time.Time
			StartDepht        float32
			LastStartData     ScapeDataD
			LastTripData      ScapeDataD
			FlagChangeTrip    int
		}
		StartActiveOperation time.Time
		//Operationtype   OperationtypeD
		Log *logrus.Logger
		Mu  *sync.RWMutex
		Cfg *ConfigDt
	}
	//SummarysheetT -type result list
	SummarysheetT struct {
		Sheet   OperationOne
		Details []OperationOne
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

//ScapeParamtype - scape parametrs yaml
type ScapeParamtype struct {
	Name  string  `yaml:"Name"`
	Gid   int     `yaml:"Gid"`
	Delta float32 `yaml:"Delta"`
}

//ConfigDt - configuration structure yaml type
type ConfigDt struct {
	Pmin                     float32          `yaml:"Pmin"`
	Flowmin                  float32          `yaml:"Flowmin"`
	Rotationmin              float32          `yaml:"Rotationmin"`
	PresFlowCheck            int              `yaml:"PresFlowCheck"`
	DephtTool                float32          `yaml:"DephtTool"`
	RotorSl                  int              `yaml:"RotorSl "`
	DirectionalCheck         int              `yaml:"DirectionalCheck"`
	BeforeDrillString        string           `yaml:"BeforeDrillString"`
	ShowParamRotSl           int              `yaml:"ShowParamRotSl "`
	ShowParamCircl           int              `yaml:"ShowParamCircl"`
	ShowParamWiper           int              `yaml:"ShowParamWiper"`
	ChangeCircWiperfromDrill int              `yaml:"ChangeCircWiperfromDrill"`
	Avgstand                 float32          `yaml:"Avgstand"`
	Wbitmax                  float32          `yaml:"Wbitmax"`
	Pressmax                 float32          `yaml:"Pressmax"`
	TimeIntervalAll          int              `yaml:"TimeIntervalAll"`
	TimeIntervalMkTrip       int              `yaml:"TimeIntervalMkTrip"`
	TimeIntervalMaxMkconn    int              `yaml:"TimeIntervalMaxMkconn"`
	TimeIntervalKNBK         int              `yaml:"TimeIntervalKNBK"`
	MinLenforTrip            int              `yaml:"MinLenforTrip"`
	ScapeParam               []ScapeParamtype `yaml:"Scapeparam"`
	Operationtype            [15]string       `yaml:"Operationtype"`
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
