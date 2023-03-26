package balancingservices

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	// store "github.com/AlexandrM09/DDOperation/pkg/StoreMap"
	detElem "github.com/AlexandrM09/DDOperation/pkg/algoritmdetermine"
	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	steam "github.com/AlexandrM09/DDOperation/pkg/steamd"
	store "github.com/AlexandrM09/DDOperation/pkg/storedetermine"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// const (
//
//	topic1 = "Sensors data"
//	topic2 = "Determine_"
//
// )
const (
	countwell              = 3
	countWellRepoSave      = 3
	countDetermiElementary = 3
	countDetermiSummary    = 3
	countReadBufferChanel  = 1000
	timeleave              = 3
)

var topic = []string{"Sensors data",
	//   "Sensors data save",
	//   "Determine save",
	"Determine",
	"Summary",
}

type (
	//SteamI2 basic interface for operations recognition variant two
	steamI2 interface {
		ReadSteam(ErrCh chan error) chan nt.ScapeDataD
	}

	//DetermineElementary well
	determineElementaryI2 interface {
		Run(ErrCh chan error)
		WaitandGetReault() map[string]*nt.SaveDetElementary
	}
	determineSummaryI interface {
		Run(ErrCh chan error)
		WaitandGetReault() map[string]*nt.SummaryResult
	}
	steams struct {
		Steams [countwell]steamI2
		Wg     *sync.WaitGroup
	}
	serviceCh struct {
		In  []chan interface{}
		Out []chan interface{}
	}
	Well = struct {
		Id         string
		name       string
		pathsource string
	}
	Wells    = []Well
	PoolWell struct {
		wells  *Wells
		Log    *logrus.Logger
		Cfg    *nt.ConfigDt
		Steams steams //steam.SteamCsv
		// Store            *store.Brocker
		store            store.Storedetermine
		detElementary    determineElementaryI2
		DetermineSummary determineSummaryI
		ServicesCh       map[string]serviceCh
		ctx              context.Context
		Cancel           context.CancelFunc
	}
	rrobin = struct {
		n int
		// name string
	}
	Roundrobin struct {
		countclients int
		countworker  int
		wrk          []rrobin
	}
)

func (pW *PoolWell) Building(path string, durat int) error {
	var err error
	pW.Log = createLog(logrus.DebugLevel)
	pW.Cfg, err = LoadConfigYaml(path)
	if err != nil {
		pW.Log.Fatal("Error loading the configuration file")
		return err
	}
	fmt.Printf("pW.Cfg=%v\n", pW.Cfg)
	//Load Well
	pW.wells, err = LoadWell(countwell)
	for i := range *pW.wells {
		pW.Log.Debugf("after load well i:%d,id:%s\n", i, (*pW.wells)[i].Id)
	}
	if err != nil {
		pW.Log.Fatal("Error loading the wells information")
		return err
	}
	//EvetBus create
	pW.store = store.New()
	//Make Steam array

	pW.ctx, pW.Cancel = context.WithTimeout(context.Background(), time.Duration(durat)*time.Second)
	buildSteam(pW.ctx, &pW.Steams, pW.wells, pW.Log)
	pW.Log.Debugf("after build Steam: %v\n", pW.Steams)
	//Make DetElementary
	pW.ServicesCh = map[string]serviceCh{
		topic[1]: {
			In:  []chan interface{}{make(chan interface{}, 50)},
			Out: []chan interface{}{make(chan interface{}, 50)},
		},
		topic[2]: {
			In:  []chan interface{}{make(chan interface{}, 50)},
			Out: []chan interface{}{make(chan interface{}, 50)},
		},
	}
	pW.detElementary = detElem.NewDetElementary(pW.ctx,
		pW.ServicesCh[topic[1]].In[0], pW.ServicesCh[topic[1]].Out[0], pW.Log, pW.Cfg, toArrayWellId(*pW.wells), &pW.store)
	pW.DetermineSummary = detElem.New(pW.ctx,
		pW.ServicesCh[topic[2]].In[0], pW.ServicesCh[topic[2]].Out[0], pW.Log, pW.Cfg, toArrayWellId(*pW.wells), &pW.store)
	fmt.Printf("Exit Building \n")
	return nil
}
func toArrayWellId(w []Well) []string {
	a := make([]string, len(w))
	for i, v := range w {
		a[i] = v.Id
	}
	return a
}
func (pW *PoolWell) Run() error {

	defer func() {

		fmt.Printf("after Run defer func() 1 \n")
		pW.Cancel()
		fmt.Printf("after Run defer func() 2 \n")
		//pW.Store.Close()
		fmt.Printf("after Run defer func() \n")
		pW.Log.Info("after Run defer func()")
		// for key := range pW.ServicesCh {
		// 	// for j := range pW.ServicesCh[key].In {
		// 	// 	close(pW.ServicesCh[key].In[j])
		// 	// }
		// 	// for j := range pW.ServicesCh[key].Out {
		// 	// 	close(pW.ServicesCh[key].Out[j])
		// 	// }
		// }
	}()

	//Запускаем распознавание элементарных операций
	ErrSteam := make(chan error, countwell*2)
	pW.detElementary.Run(ErrSteam)
	//Запускаем DetermineSummary
	pW.DetermineSummary.Run(ErrSteam)

	//Читаем из detElementary и пишем на вход determineSummary
	go func() {
		for v := range pW.ServicesCh[topic[1]].Out[0] {
			pW.ServicesCh[topic[2]].In[0] <- v

		}
		pW.Log.Debugf("close(pW.ServicesCh[topic[2]].In[0])")
		close(pW.ServicesCh[topic[2]].In[0])
	}()
	//Читаем из DetermineSummary
	go func() {
		for v := range pW.ServicesCh[topic[2]].Out[0] {
			_ = v

		}
		// close(pW.ServicesCh[topic[2]].In[0])
	}()
	tick1 := time.NewTicker(time.Duration(500 * time.Millisecond))
	defer tick1.Stop()
	//Пишем ошибки в лог
	go func() {
		count := 0
		for {

			select {
			case <-tick1.C:
				{
					count += 500
					pW.Log.Infof("duration is %dms", count)
				}
			case <-pW.ctx.Done():
				{
					fmt.Printf("exit to error <-ctx.Done() \n")
					return
				}
			case err := <-ErrSteam:
				pW.Log.Errorf("Error", err)

			}
		}
	}()

	//Запускаем чтение данных из csv файлов
	//Ждем окончания данных
	go runSteam(&pW.Steams, pW.ServicesCh[topic[1]].In[0], ErrSteam, countwell, pW.Log)
	pW.Log.Debugf("after close(pW.ServicesCh[topic[1]].In[0]) ")
	//Ждем окончания первичного распознавания операций
	resElementary := pW.detElementary.WaitandGetReault()
	pW.Log.Debugf("after pW.detElementary.WaitandGetReault() ")
	_ = resElementary
	//Ждем окончания свернутого списка операций
	resSummary := pW.DetermineSummary.WaitandGetReault()
	_ = resSummary
	//
	//Печатаем результат
	for j := range *pW.wells {
		id := (*pW.wells)[j].Id
		data2 := resSummary[id].Summarysheet
		fmt.Printf("Start print Summarysheet(idWell=%s)  short form len=%v \n", id, len(data2))
		//Подробный
		//fmt.Printf("Start print Summarysheet len=%v \n", len(data2))
		for i := 0; i < len(data2); i++ {
			fmt.Println(FormatSheet(data2[i]))
			for k := 0; k < len(data2[i].Details); k++ {
				fmt.Println(FormatSheetDetails(data2[i].Details[k]))
			}
		}
		//Короткий
		if len(data2) > 0 {
			fmt.Println(data2[0].Sheet.StartData.Time.Format("2006-01-02"))
		}
		for i := 0; i < len(data2); i++ {
			fmt.Println(FormatSheet2(data2[i]))
		}
	}

	pW.Log.Debugf("cahnel doneAllSteam reding")

	return nil
}

// func FormatSheet(Op nt.OperationOne) string {

// 	return fmt.Sprintf("%s | %s - %s\n",
// 		Op.StartData.Time.Format("15:04"),
// 		Op.Operaton,
// 		Op.Params)
// }

// LoadConfigYaml - load config file yaml
func LoadConfigYaml(path string) (*nt.ConfigDt, error) {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c nt.ConfigDt
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

//

func buildSteam(ctx context.Context, steams *steams, arwells *Wells, l *logrus.Logger) {

	for i := range steams.Steams {
		id1 := (*arwells)[i].Id
		fmt.Printf("buildSteam i:%d,id:%s,path:%s\n", i, id1, (*arwells)[i].pathsource)
		steams.Steams[i] = steam.New(ctx, id1, (*arwells)[i].pathsource, l)

	}
	steams.Wg = &sync.WaitGroup{}
	//return
	fmt.Printf("buildSteam: %v\n", *steams)
}
func runSteam(steams *steams, Out chan interface{}, ErrSteam chan error, count int, l *logrus.Logger) {
	for i := range steams.Steams {
		steams.Wg.Add(1)
		go func(k int) {
			n := k
			l.Debugf("start steams %d", n)

			for v := range steams.Steams[n].ReadSteam(ErrSteam) {
				value := v
				l.Debugf("steams %d ,id = %s, value=%v", n, steams.Steams[k].(*steam.SteamCsv).Id, v.Values[3])
				//
				Out <- value
			}
			Out <- nt.ScapeDataD{Id: steams.Steams[k].(*steam.SteamCsv).Id, StatusLastData: true}
			steams.Wg.Done()
		}(i)
	}
	steams.Wg.Wait()
	l.Debugf("after pW.Steams.Wg.Wait() ")
	close(Out)
}

func LoadWell(count int) (awells *Wells, er error) {
	fmt.Printf("LoadWell count=%d\n", count)
	t := make(Wells, 0, count)

	for i := 0; i < count; i++ {
		fmt.Printf("LoadWell i=%d\n", i)
		t = append(t,
			Well{
				fmt.Sprintf("%d", int64(i+100)),
				fmt.Sprintf("Well%d", i),
				"",
			})
	}
	t[0].pathsource = "source/source.zip"
	if len(t) > 1 {
		t[1].pathsource = "source/source1.zip"
	}
	if len(t) > 2 {
		t[2].pathsource = "source/source2.zip"
	}
	for i := range t {
		fmt.Printf("LoadWell well i:%d,id:%s\n", i, t[i].Id)
	}
	return &t, nil
}
func (r *Roundrobin) add(n int) int {
	if n > r.countclients {
		res := r.add(n - r.countclients)
		return res
	}
	return n
}
func (r *Roundrobin) Next() {
	for i := 1; i <= r.countworker; i++ {
		r.wrk[i].n = r.add(r.wrk[i].n + r.countworker)
	}
}
func (r *Roundrobin) Init() {
	// r.wrk = make([]rrobin, r.countworker, r.countworker)
	// for i := 1; i <= r.countworker; i++ {
	// 	r.wrk = append(r.wrk, rrobin{
	// 		n:    i,
	// 		name: "",
	// 	})
	// }
}
func (r *Roundrobin) Get(n int) int {
	if (n < 1) || (n > r.countworker) {
		return -1
	}
	return r.wrk[n].n
}

type plainFormatter struct {
	TimestampFormat string
	LevelDesc       []string
}

func (f *plainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := fmt.Sprint(entry.Time.Format(f.TimestampFormat))
	return []byte(fmt.Sprintf("[%s] %s %s:%d  %s \n", f.LevelDesc[entry.Level], timestamp,
		filepath.Base(entry.Caller.File), entry.Caller.Line, entry.Message)), nil
}

func createLog(ll logrus.Level) *logrus.Logger {
	plainFormatter := new(plainFormatter)
	plainFormatter.TimestampFormat = "2006-01-02 15:04:05.000"
	plainFormatter.LevelDesc = []string{"PANC", "FATL", "ERRO", "WARN", "INFO", "DEBG"}
	var log = logrus.New()
	log.SetReportCaller(true)
	log.SetFormatter(plainFormatter)
	log.SetLevel(ll)
	log.Out = os.Stdout
	file, err := os.OpenFile("logrus.log", os.O_CREATE, 0666) //|os.O_WRONLY
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
	return log

}
func FormatSheet2(sh nt.SummarysheetT) string {
	tempt, _ := time.Parse("15:04:01", "00:00:00")
	dur := sh.Sheet.StopData.Time.Sub(sh.Sheet.StartData.Time)
	return fmt.Sprintf("%s | %s |%s |%s %s",
		sh.Sheet.StartData.Time.Format("15:04"),
		sh.Sheet.StopData.Time.Format("15:04"),
		tempt.Add(dur).Format("15:04"),
		sh.Sheet.Operaton,
		sh.Sheet.Params)
}

// FormatSheet format string function
func FormatSheet(sh nt.SummarysheetT) string {
	return fmt.Sprintf("%s | %s |%s%s",
		sh.Sheet.StartData.Time.Format("2006-01-02 15:04:05"),
		sh.Sheet.StopData.Time.Format("15:04:05"),
		sh.Sheet.Operaton,
		sh.Sheet.Params)
}

// FormatSheetDetails format string function
func FormatSheetDetails(Det nt.OperationOne) string {
	return fmt.Sprintf("____%s | %s |%s %s",
		Det.StartData.Time.Format("15:04:05"),
		Det.StopData.Time.Format("15:04:05"),
		Det.Operaton, Det.Params)
}
