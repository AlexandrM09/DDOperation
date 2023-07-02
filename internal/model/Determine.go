package model

import (
	"io/ioutil"
	"time"

	yaml "gopkg.in/yaml.v2"
)

const ScapeDataCount = 50

type (

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
	// golang CRUD with gin and ent
	ScapeDataD struct {
		Id       string
		SchemaId int64
		Time     time.Time
		// Count          int
		Values           [ScapeDataCount]float32
		StatusLastData   bool //true если последняя запись
		StatusChangeDate bool //true если изменилась дата

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
		Lastchangeoperation                        OperationtypeT
		Count                                      int
		Operaton                                   OperationtypeT
		Params                                     string
	}
	//OperationOne for sending other servises
	SendingTopicDeterm struct {
		IdWell    string
		Operation OperationOne
		Status string
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

type ResultSheet struct {
	Id        string
	ResSheet  SummarysheetT
	Firstflag int
	Startflag int
	StartTime time.Time
	//stopTime  time.Time
	SumItemDr int
	//res       OperationOne
	//next     nt.OperationOne
	NextTime struct {
		Flag  int
		Start time.Time
	}
}
type SaveDetElementary = struct {
	IdWell               string
	OperationList        []OperationOne
	ScapeFullData        bool
	LastScapeData        ScapeDataD
	ScapeData            ScapeDataD
	Sm                   map[string]int //Мапа индексов ScapeData
	ActiveOperation      int
	StartActiveOperation time.Time
	Temp                 struct {
		LastToolDepht     float32
		LastTimeToolDepht time.Time
		StartDepht        float32
		LastStartData     ScapeDataD
		LastTripData      ScapeDataD
		FlagChangeTrip    int
	}
}

// Возвращает значение по имени датчика
func (scd *ScapeDataD) GetSensorsNV(name string, sm map[string]int) float32 {
	v, ok := sm[name]
	if ok && v >= 0 && v < len(scd.Values) {
		return scd.Values[v]
	}
	return 0
}

type SummaryResult struct {
	IdWell       string
	Summarysheet []SummarysheetT
	Sc           ResultSheet
}

// ScapeParamtype - scape parametrs yaml
type ScapeParamtype struct {
	Name  string  `yaml:"Name"`
	Gid   int     `yaml:"Gid"`
	Delta float32 `yaml:"Delta"`
}
type OperationtypeT struct {
	Description string `yaml:"Description"`
	IdCodeOper  int64  `yaml:"IdCodeOper"`
}

// ConfigDt - configuration structure yaml type
type ConfigDt struct {
	Pmin                     float32            `yaml:"Pmin"`
	Flowmin                  float32            `yaml:"Flowmin"`
	Rotationmin              float32            `yaml:"Rotationmin"`
	PresFlowCheck            int                `yaml:"PresFlowCheck"`
	DephtTool                float32            `yaml:"DephtTool"`
	RotorSl                  int                `yaml:"RotorSl "`
	DirectionalCheck         int                `yaml:"DirectionalCheck"`
	BeforeDrillString        string             `yaml:"BeforeDrillString"`
	ShowParamRotSl           int                `yaml:"ShowParamRotSl "`
	ShowParamCircl           int                `yaml:"ShowParamCircl"`
	ShowParamWiper           int                `yaml:"ShowParamWiper"`
	ChangeCircWiperfromDrill int                `yaml:"ChangeCircWiperfromDrill"`
	Avgstand                 float32            `yaml:"Avgstand"`
	Wbitmax                  float32            `yaml:"Wbitmax"`
	Pressmax                 float32            `yaml:"Pressmax"`
	TimeIntervalAll          int                `yaml:"TimeIntervalAll"`
	TimeIntervalMkTrip       int                `yaml:"TimeIntervalMkTrip"`
	TimeIntervalMaxMkconn    int                `yaml:"TimeIntervalMaxMkconn"`
	TimeIntervalKNBK         int                `yaml:"TimeIntervalKNBK"`
	MinLenforTrip            int                `yaml:"MinLenforTrip"`
	ScapeParam               []ScapeParamtype   `yaml:"Scapeparam"`
	Operationtype            [15]OperationtypeT `yaml:"Operationtype"`
}

func LoadDetermineConfigYaml(path string) (*ConfigDt, error) {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c ConfigDt
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Функция фозвращает мапу с индексами датчиков в sensors
func SensorsMap() map[string]int {
	res := make(map[string]int, len(Sensors))
	for i, v := range Sensors {
		res[v.Name] = i
	}
	return res
}
type Sensor struct {
	Gid      int
	FullName string
	Name     string
	Units    string
	Active   bool
}
var Sensors = [...]Sensor{{Gid: 101, FullName: "Глубина забоя", Name: "Глубина забоя", Units: "м.", Active: true},
	{Gid: 110, FullName: "Положение долота по свечам", Name: "Пол.дол. по св.", Units: "м.", Active: true},
	{Gid: 111, FullName: "Число свечей", Name: "Число свечей", Units: " ", Active: true},
	{Gid: 115, FullName: "Положение долота", Name: "Пол. долота", Units: "м.", Active: true},
	{Gid: 103, FullName: "Положение тальблока", Name: "Тальблок", Units: "м.", Active: true},
	{Gid: 104, FullName: "Скорость движения тальблока", Name: "Скор. тальблока", Units: "м/с", Active: true},
	{Gid: 112, FullName: "Скорость бурения по времени", Name: "Скор.бур.по вр.", Units: "м/ч", Active: true},
	{Gid: 105, FullName: "Положение клиньев", Name: "Клинья", Units: " ", Active: true},
	{Gid: 200, FullName: "Вес на крюке", Name: "Вес на крюке", Units: "т.", Active: true},
	{Gid: 201, FullName: "Вес колонны", Name: "Вес колонны", Units: "т.", Active: true},

	{Gid: 300, FullName: "Давление на манифольде", Name: "Давление нагн.", Units: "атм.", Active: true},
	{Gid: 600, FullName: "Плотность р-ра на входе", Name: "Плотн. на входе", Units: "г/см3", Active: true},
	{Gid: 604, FullName: "Плотность р-ра под виброситом", Name: "Плотн. под виб.", Units: "г/см3", Active: true},
	{Gid: 605, FullName: "Плотность р-ра на выходе", Name: "Плотн. на вых.", Units: "г/см3", Active: true},
	{Gid: 711, FullName: "Объем р-ра в емкости 1", Name: "Объем в 1 емк.", Units: "м3", Active: true},
	{Gid: 712, FullName: "Объем р-ра в емкости 2", Name: "Объем во 2 емк.", Units: "м3", Active: true},
	{Gid: 713, FullName: "Объем р-ра в емкости 3", Name: "Объем в 3 емк.", Units: "м3", Active: true},
	{Gid: 714, FullName: "Объем р-ра в емкости под виброситом", Name: "Объем под вибр.", Units: "м3", Active: true},
	{Gid: 715, FullName: "Объем р-ра в доливочной емкости", Name: "Объем в дол.емк", Units: "м3", Active: true},
	{Gid: 716, FullName: "Объем р-ра в емкости 4", Name: "Объем в 4 емк.", Units: "м3", Active: true},

	{Gid: 717, FullName: "Объем р-ра в емкости 5", Name: "Объем в 5 емк.", Units: "м3", Active: true},
	{Gid: 718, FullName: "Объем р-ра в емкости 6", Name: "Объем в 6 емк.", Units: "м3", Active: true},
	{Gid: 720, FullName: "Суммарный объем в емкостях", Name: "Сум.объем в емк", Units: "м3", Active: true},
	{Gid: 800, FullName: "Температура на входе", Name: "Темпер. на вх.", Units: "°С", Active: true},
	{Gid: 900, FullName: "Температура раствора на выходе", Name: "Темпер. на вых.", Units: "°С", Active: true},
	{Gid: 1300, FullName: "Крутящий момент на роторе", Name: "Кр.м на роторе", Units: "т*м", Active: true},
	{Gid: 1600, FullName: "Сумма газов", Name: "Сумма газов", Units: "%", Active: true},
	{Gid: 1601, FullName: "C1", Name: "C1", Units: "%", Active: true},
	{Gid: 1602, FullName: "C2", Name: "C2", Units: "%", Active: true},
	{Gid: 1603, FullName: "C3", Name: "C3", Units: "%", Active: true},

	{Gid: 1604, FullName: "C4", Name: "C4", Units: "%", Active: true},
	{Gid: 1605, FullName: "C5", Name: "C5", Units: "%", Active: true},
	{Gid: 1626, FullName: "iC4", Name: "iC4", Units: "%", Active: true},
	{Gid: 1627, FullName: "iC5", Name: "iC5", Units: "%", Active: true},
	{Gid: 1003, FullName: "Расход на выходе", Name: "Расх. на вых.", Units: "л/c", Active: true},
	{Gid: 1200, FullName: "Число оборотов ротора в мин.", Name: "Обороты рот.", Units: "мин-1", Active: true},
	{Gid: 50, FullName: "Число ходов 1 насоса", Name: "Ходы 1 нас.", Units: "мин-1", Active: true},
	{Gid: 51, FullName: "Число ходов 2 насоса", Name: "Ходы 2 нас.", Units: "мин-1", Active: true},
	{Gid: 52, FullName: "Число ходов 3 насоса", Name: "Ходы 3 нас.", Units: "мин-1", Active: true},
	{Gid: 1001, FullName: "Расход на входе", Name: "Расх на вх.", Units: "л/c", Active: true},

	{Gid: 113, FullName: "Глубина для газа", Name: "Глубина газа", Units: "м.", Active: true},
	{Gid: 114, FullName: "Прогнозируемое время вых. газа", Name: "П.вр. вых. газа", Units: "мин.", Active: true},
	{Gid: 106, FullName: "Скорость бурения {по глубине}", Name: "Скорость бур.", Units: "м/ч", Active: true},
	{Gid: 107, FullName: "ДМК", Name: "ДМК", Units: "мин/м", Active: true},
	{Gid: 202, FullName: "Нагрузка на долото", Name: "Нагрузка на дол", Units: "т.", Active: true},
	{Gid: 30, FullName: "Температура наружного воздуха", Name: "Темп. воздуха", Units: "°", Active: true},
	{Gid: 1301, FullName: "Момент на машинном ключе", Name: "Момент на ключе", Units: "т*м", Active: true},
	//calculated
	{Gid: 30100, FullName: "Скорость инструмента в открытом стволе", Name: "Скор. откр. ствол", Units: "м/с", Active: true},
	{Gid:1651,FullName: "Относительная конц.  C1", Name:"C1 отн.",Units:"%",Active:true},
	{Gid:1652,FullName:"Относительная конц.  C2",Name:"C2 отн.",Units:"%",Active:true},
	{Gid:1653,FullName:"Относительная конц.  C3",Name:"C3 отн.",Units:"%",Active:true},
	{Gid:1654,FullName:"Относительная конц.  C4",Name:"C4 отн.",Units:"%",Active:true},
	{Gid:1655,FullName:"Относительная конц.  C5",Name:"C5 отн.",Units:"%",Active:true},
	{Gid:1657,FullName:"Относительная конц.  iC4",Name:"iC4 отн.",Units:"%",Active:true},
	{Gid:1658,FullName:"Относительная конц.  iC5",Name:"iC5 отн.",Units:"%",Active:true},
	
}
