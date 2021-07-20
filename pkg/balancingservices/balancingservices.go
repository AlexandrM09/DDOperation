package balancingservices

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	detElem "github.com/AlexandrM09/DDOperation/pkg/algoritmdetermine"
	bus "github.com/AlexandrM09/DDOperation/pkg/eventbussimple"
	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	steam "github.com/AlexandrM09/DDOperation/pkg/steamd"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type (
	arStmI = [countwell]steam.SteamI2
	arDeEl = [countDetermiElementary]detElem.DetermineElementaryI2
	Well   = struct {
		id   int64
		name string
		path string
		Data chan nt.ScapeDataD
	}
	Wells    = []Well
	PoolWell struct {
		wells     *Wells
		Log       *logrus.Logger
		Cfg       *nt.ConfigDt
		Steams    arStmI //steam.SteamCsv
		Evnt      *bus.Eventbus
		detElmtrs arDeEl
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
	countwell              = 1
	countWellRepoSave      = 3
	countDetermiElementary = 3
	countDetermiSummary    = 3
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
		pW.Log.Infof("well i:%d,id:%d\n", i, (*pW.wells)[i].id)
	}
	if err != nil {
		pW.Log.Fatal("Error loading the wells information")
		return err
	}
	//EvetBus create
	pW.Evnt = bus.Neweventbussimple(countwell, pW.Log)

	//Make Steam array
	buildSteam(&pW.Steams, pW.wells, pW.Log, durat, countwell)
	pW.Log.Infof("after build Steam: %v\n", pW.Steams)
	for i, _ := range pW.Steams {
		p := pW.Steams[i]
		pW.Log.Infof("evnt.AddWell:%d,%v\n", i, p)
		pW.Evnt.AddWell(p.(*steam.SteamCsv).Id, 100) // 50 - read buffer (chanel)
	}

	buildDetermineEl(&pW.detElmtrs, pW.Log, pW.Cfg, pW.Evnt, durat, countDetermiElementary)
	fmt.Printf("Exit Building \n")
	return nil
}
func (pW *PoolWell) Run() error {
	//StaemDataCh := make(chan nt.ScapeDataD, countwell)
	ErrSteam := make(chan error, countwell)
	DoneSteam := make(chan struct{}, countwell)
	//Start csv steam
	runSteam(&pW.Steams, DoneSteam, ErrSteam, pW.Evnt, countwell)
	to := time.After(3 * time.Second)
	done := make(chan bool, 1)
	fmt.Printf("Start run \n")

	go func() {

		for {
			//time.Sleep(2 * time.Millisecond)
			select {
			case <-to:
				DoneSteam <- struct{}{}
				done <- true
				return
			case err := <-ErrSteam:
				pW.Log.Errorf("Error", err)
			default:
				for i := 0; i < countwell; i++ {
					id := pW.Steams[i].(*steam.SteamCsv).Id

					if t, ok2 := pW.Evnt.Receive("ScapeData", id); ok2 {

						d, ok3 := t.(*nt.ScapeDataD)
						if ok3 {
							pW.Log.Debugf("Run:Read ScapeData id=%s,count=%d,t=%s", id, d.Count, d.Time.Format("2006-01-02 15:04:05"), d.Values[3])

							fmt.Printf("Id=%s,count=%d, read time=%s,val=%.3f \n", id, d.Count, d.Time.Format("2006-01-02 15:04:05"), d.Values[3])

						}
						//id = ""
						//d.Count = 0
						//ok2 = false
						//ok3 = false
						//t = nil
					}
				}
			}
		}
	}()
	fmt.Printf("expectation gorooting \n")
	<-done
	fmt.Printf("the program has been successfully completed \n")
	/*
		ErrEl := make(chan error, countDetermiElementary)
		DoneEl := make(chan struct{}, countDetermiElementary)
		runDetermineEl(pW.detElmtrs, DoneEl, ErrEl, pW.Evnt, countDetermiElementary)
		//Waiting end SteamCsv
		//	for i := 0; i < countwell; i++ {
		//		pW.Steams[i].(*steam.SteamCsv).Wg.Wait()
		//	}
		for i := 0; i < countDetermiElementary; i++ {
			pW.detElmtrs[i].(*detElem.DetermineElementary).Wg.Wait()
		}
		fmt.Scanln()
		pW.Log.Infof("After wait ")
		for i := 0; i < countwell; i++ {
			id := pW.Steams[i].(*steam.SteamCsv).Id
			pW.Log.Infof("result Well:%s:\n", id)
			b := true
			for b {
				t := pW.Evnt.Receive("Determine", id)
				if t == nil {
					b = false
					continue
				}
				d, ok := t.(nt.OperationOne)
				if !ok {
					pW.Log.Errorf("not OperationOne interface:%v", t)
				}
				pW.Log.Infof("Result:", FormatSheet(d))
			}
		}
	*/
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

//LoadConfigYaml - load config file yaml
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
		id1 := fmt.Sprintf("%d", (*arwells)[i].id)
		fmt.Printf("buildSteam id1:%s\n", id1)
		steams[i] = &steam.SteamCsv{
			Id:       id1,
			FilePath: (*arwells)[i].path,
			Dur:      time.Millisecond * time.Duration(durat), // max time duration reading
			Log:      l,
			//Wg:       &sync.WaitGroup{},
		}
	}
	fmt.Printf("buildSteam: %v\n", *steams)
	//return
}
func runSteam(steams *arStmI,
	DoneSteam chan struct{}, ErrSteam chan error, evnt *bus.Eventbus, count int) {
	for i := 0; i < count; i++ {
		//n := i
		go func(k int) {
			steams[k].(*steam.SteamCsv).Log.Infof("Start Steam id=%s", steams[k].(*steam.SteamCsv).Id)
			//steams[i].(*steam.SteamCsv).ScapeDataCh = steams[k].ReadCsv(DoneSteam, ErrSteam)
			for v := range steams[k].ReadSteam(DoneSteam, ErrSteam) {
				//St.Log.Debugf("id=%s sending in chanel line %d, time:%s ", St.Id, n, ScapeData.Time.Format("2006-01-02 15:04:05"))
				value := v
				for !evnt.Send("ScapeData", steams[k].(*steam.SteamCsv).Id, &value) {
					//time.Sleep(100 * time.Microsecond)
					steams[k].(*steam.SteamCsv).Log.Debugf("runSteam lock: id=%s sending in steem line %d, time:%s", steams[k].(*steam.SteamCsv).Id, v.Count, v.Time)
				}
				steams[k].(*steam.SteamCsv).Log.Debugf("runSteam: id=%s sending in steem line %d, time:%s", steams[k].(*steam.SteamCsv).Id, v.Count, v.Time)
			}
		}(i)
	}
}
func buildDetermineEl(detElmtrs *arDeEl,
	l *logrus.Logger, Cf *nt.ConfigDt, evnt *bus.Eventbus, durat int, count int) {

	for i := 0; i < count; i++ {

		detElmtrs[i] = &detElem.DetermineElementary{
			Log: l,
			Wg:  &sync.WaitGroup{},
			Cfg: Cf,
			/*				Pmin:                     Cf.Pmin,
				Flowmin:                  Cf.Flowmin,
				Rotationmin:              Cf.Rotationmin,
				PresFlowCheck:            Cf.PresFlowCheck,
				DephtTool:                Cf.DephtTool,
				RotorSl:                  Cf.RotorSl,
				DirectionalCheck:         Cf.DirectionalCheck,
				BeforeDrillString:        Cf.BeforeDrillString,
				ShowParamRotSl:           Cf.ShowParamRotSl,
				ShowParamCircl:           Cf.ShowParamCircl,
				ShowParamWiper:           Cf.ShowParamWiper,
				ChangeCircWiperfromDrill: Cf.ChangeCircWiperfromDrill,
				Avgstand:                 Cf.Avgstand,
				Wbitmax:                  Cf.Wbitmax,
				Pressmax:                 Cf.Pressmax,
				TimeIntervalAll:          Cf.TimeIntervalAll,
				TimeIntervalMkTrip:       Cf.TimeIntervalMkTrip,
				TimeIntervalMaxMkconn:    Cf.TimeIntervalMaxMkconn,
				TimeIntervalKNBK:         Cf.TimeIntervalKNBK,
				MinLenforTrip:            Cf.MinLenforTrip,
				ScapeParam:               Cf.ScapeParam,
				Operationtype:            Cf.Operationtype,
			},
			*/
			IdIn:  "ScapeData",
			IdOut: "Determine",
			Id:    fmt.Sprintf("%d", i+11),
			Evnt:  evnt,
		}
		//	detElmtrs[i].(*detElem.DetermineElementary).Cfg = *Cf
	}
}
func runDetermineEl(El [countDetermiElementary]detElem.DetermineElementaryI2,
	DoneEl chan struct{}, ErrEl chan error, evnt *bus.Eventbus, count int) {
	for i := 0; i < count; i++ {
		n := i
		go func(k int) {
			//steams[i].(*steam.SteamCsv).ScapeDataCh = steams[k].ReadCsv(DoneSteam, ErrSteam)
			El[k].Run(DoneEl, ErrEl)
		}(n)
	}
}
func LoadWell(count int) (awells *Wells, er error) {
	fmt.Printf("LoadWell count=%d\n", count)
	t := make(Wells, 0, count)

	for i := 0; i < count; i++ {
		fmt.Printf("LoadWell i=%d\n", i)
		t = append(t,
			Well{
				int64(i + 11),
				fmt.Sprintf("Well%d", i),
				"",
				make(chan nt.ScapeDataD),
			})
	}
	//awells = &t

	t[0].path = "source/source.zip"
	t[1].path = "source/source1.zip"
	t[2].path = "source/source2.zip"
	for i, _ := range t {
		fmt.Printf("LoadWell well i:%d,id:%d\n", i, t[i].id)
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

//
type plainFormatter struct {
	TimestampFormat string
	LevelDesc       []string
}

//
func (f *plainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := fmt.Sprintf(entry.Time.Format(f.TimestampFormat))
	return []byte(fmt.Sprintf("[%s] %s %s:%d  %s \n", f.LevelDesc[entry.Level], timestamp,
		filepath.Base(entry.Caller.File), entry.Caller.Line, entry.Message)), nil
}

//
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
