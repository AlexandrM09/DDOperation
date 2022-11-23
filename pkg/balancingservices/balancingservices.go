package balancingservices

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	store "github.com/AlexandrM09/DDOperation/pkg/StoreMap"
	detElem "github.com/AlexandrM09/DDOperation/pkg/algoritmdetermine"
	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	steam "github.com/AlexandrM09/DDOperation/pkg/steamd"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type (
	//SteamI2 basic interface for operations recognition variant two
	SteamI2 interface {
		ReadSteam(ErrCh chan error) chan nt.ScapeDataD
		ReadSteamTime(ErrCh chan error) chan nt.ScapeDataD
		Stop()
	}
	arStmI = [countwell]SteamI2
	//DetermineElementary well
	DetermineElementaryI2 interface {
		Run(ErrCh chan error)
		//ReadTime(Done chan struct{}, ErrCh chan error)
		Stop() map[string]detElem.SaveDetElementary
	}
	// arDeEl = [countDetermiElementary]DetermineElementaryI2
	Well   = struct {
		Id         string
		name       string
		pathsource string
		// Data chan nt.ScapeDataD
	}
	Wells    = []Well
	PoolWell struct {
		wells    *Wells
		Log      *logrus.Logger
		Cfg      *nt.ConfigDt
		Steams   arStmI //steam.SteamCsv
		Store    *store.Brocker
		detElmtr DetermineElementaryI2
		WgSteam  *sync.WaitGroup
	}
	rrobin = struct {
		n    int
		name string
	}
	Roundrobin struct {
		countclients int
		countworker  int
		wrk          []rrobin
	}
)

const (
	countwell              = 3
	countWellRepoSave      = 3
	countDetermiElementary = 3
	countDetermiSummary    = 3
	countReadBufferChanel  = 1000
	timeleave              = 3
)

func (pW *PoolWell) Building(path string, durat int) error {
	fmt.Println("Start")
	pW.Log = createLog(logrus.DebugLevel)
	err := LoadConfigYaml(path, pW.Cfg)
	if err != nil {
		pW.Log.Fatal("Error loading the configuration file")
		return err
	}
	//Load Well
	pW.wells, err = LoadWell(countwell)
	for i, _ := range *pW.wells {
		pW.Log.Infof("well i:%d,id:%s\n", i, (*pW.wells)[i].Id)
	}
	if err != nil {
		pW.Log.Fatal("Error loading the wells information")
		return err
	}
	//EvetBus create
	// pW.Evnt = bus.Neweventbussimple(countwell, pW.Log)
	pW.Store = store.New(pW.Log)
	//Make Steam array
	buildSteam(&pW.Steams, pW.wells, pW.Log, durat, countwell)
	pW.WgSteam = &sync.WaitGroup{}

	pW.Log.Infof("after build Steam: %v\n", pW.Steams)
	e := detElem.DetermineElementary{
		Log:       pW.Log,
		Wg:        &sync.WaitGroup{},
		Cfg:       pW.Cfg,
		In:        "ScapeData",
		Out:       []string{"Determine"},
		DataMapId: make(map[string]detElem.SaveDetElementary),
		Id:        make(map[string]int),
		Store:     pW.Store,
	}
	for _, v := range *pW.wells {
		pW.Log.Infof("before e.AddWell(v.Id),id:%s\n", v.Id)
		e.AddWell(v.Id)
	}
	pW.detElmtr = &e
	fmt.Printf("Exit Building \n")
	return nil
}
func (pW *PoolWell) Run() error {

	//StaemDataCh := make(chan nt.ScapeDataD, countwell)
	ErrSteam := make(chan error, countwell*2)
	//Start csv steam
	runSteam(&pW.Steams, ErrSteam, pW.Store, countwell, pW.WgSteam)
	defer func() {

		// close(ErrSteam)
		// close(DoneSteam)
		// close(DoneEl)
		// close(ErrEl)
		for k, _ := range pW.Steams {
			pW.Steams[k].Stop()
		}
		fmt.Printf("after Run defer func() 1 \n")
		// pW.Steams[0].Stop()

		pW.detElmtr.Stop()
		fmt.Printf("after Run defer func() 2 \n")
		pW.Store.CloseBrockerChanel()
		fmt.Printf("after Run defer func() \n")
		pW.Log.Info("after Run defer func()")
	}()

	go pW.detElmtr.Run(ErrSteam)
	to := time.After(timeleave * time.Second)
	doneAllSteam := make(chan struct{}, 1)
	fmt.Printf("Start run \n")
	// var aCount [countwell]int
	//waiting for the reading to finish steam
	go func() {
		pW.WgSteam.Wait()
		doneAllSteam <- struct{}{}
		fmt.Printf("cahnel doneAllSteam sending \n")
		pW.Log.Info("cahnel doneAllSteam sending")
	}()

	//time.Sleep(2 * time.Millisecond)
	for {
		select {
		case <-to:
			{
				fmt.Printf("the time limit is exhausted \n")
				pW.Log.Info("the time limit is exhausted")
				// break
				fmt.Printf("the program has been successfully completed \n")
				pW.Log.Info("the program has been successfully completed")

				return nil
			}
		case <-doneAllSteam:
			{
				fmt.Printf("cahnel doneAllSteam reding \n")
				pW.Log.Info("cahnel doneAllSteam reding")
				// break
				fmt.Printf("the program has been successfully completed \n")
				pW.Log.Info("the program has been successfully completed")

				return nil
			}
		case err := <-ErrSteam:
			pW.Log.Errorf("Error", err)

		}
	}

	/*	robin := &Roundrobin{
			countclients: countwell,
			countworker:  countDetermiElementary,
		}
		robin.Init()
		//save repo skip
		//start determineElementary
	*/
	return nil
}
func FormatSheet(Op nt.OperationOne) string {
	//	tempt, _ := time.Parse("15:04:01", "00:00:00")
	//	dur := sh.Sheet.StopData.Time.Sub(sh.Sheet.StartData.Time)
	return fmt.Sprintf("%s | %s - %s\n",
		Op.StartData.Time.Format("15:04"),
		Op.Operaton,
		Op.Params)
}

// LoadConfigYaml - load config file yaml
func LoadConfigYaml(path string, cf *nt.ConfigDt) error {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &cf)
	//json.Unmarshal()
	if err != nil {
		return err
	}
	return nil
}

//

func buildSteam(steams *arStmI, arwells *Wells, l *logrus.Logger, durat int, count int) {
	//st:=*steams
	//	steams := make([]steam.SteamI2, count)
	for i := 0; i < count; i++ {
		id1 := (*arwells)[i].Id
		fmt.Printf("buildSteam id1:%s\n", id1)
		steams[i] = steam.New(
			id1,
			(*arwells)[i].pathsource,
			time.Second*time.Duration(durat), // max time duration reading
			l,
			&sync.WaitGroup{})
	}
	//return
	fmt.Printf("buildSteam: %v\n", *steams)
}
func runSteam(steams *arStmI, ErrSteam chan error, Store *store.Brocker, count int, wg *sync.WaitGroup) {
	for i := 0; i < count; i++ {
		//n := i
		wg.Add(1)
		go func(k int) {
			// steams[k].(*steam.SteamCsv).Log.Infof("Start Steam id=%s, number=%d", steams[k].(*steam.SteamCsv).Id, k)

			for v := range steams[k].ReadSteam(ErrSteam) {

				value := v
				Store.Send("ScapeData", steams[k].(*steam.SteamCsv).Id, &value)
				// steams[k].(*steam.SteamCsv).Log.Debugf("runSteam: id=%s sending in steem line %d, time:%s", steams[k].(*steam.SteamCsv).Id, v.Count, v.Time)
			}
			wg.Done()
		}(i)
	}
}

func LoadWell(count int) (awells *Wells, er error) {
	fmt.Printf("LoadWell count=%d\n", count)
	t := make(Wells, 0, count)

	for i := 0; i < count; i++ {
		fmt.Printf("LoadWell i=%d\n", i)
		t = append(t,
			Well{
				fmt.Sprintf("%d", int64(i+11)),
				fmt.Sprintf("Well%d", i),
				"",
				// make(chan nt.ScapeDataD),
			})
	}
	//awells = &t

	t[0].pathsource = "source/source.zip"
	t[1].pathsource = "source/source1.zip"
	t[2].pathsource = "source/source2.zip"
	for i, _ := range t {
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
	r.wrk = make([]rrobin, r.countworker, r.countworker)
	for i := 1; i <= r.countworker; i++ {
		r.wrk = append(r.wrk, rrobin{
			n:    i,
			name: "",
		})
	}
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
	timestamp := fmt.Sprintf(entry.Time.Format(f.TimestampFormat))
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
