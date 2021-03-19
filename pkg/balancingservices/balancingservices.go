package balancingservices

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	bus "github.com/AlexandrM09/DDOperation/pkg/eventbussimple"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	steam "github.com/AlexandrM09/DDOperation/pkg/steamd"
)

type (
	Well = struct {
		id   int64
		name string
		path string
		Data chan nt.ScapeDataD
	}
	Wells    = []Well
	poolWell struct {
		wells Wells
		Log   *logrus.Logger
		Cfg   *nt.ConfigDt
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
	countwell              = 10
	countWellRepoSave      = 3
	countDetermiElementary = 3
	countDetermiSummary    = 3
)

func (pW *poolWell) Building(path string, durat int) error {
	fmt.Println("Start")
	pW.Log = createLog(logrus.DebugLevel)

	err := LoadConfigYaml(path, pW.Cfg)
	if err != nil {
		pW.Log.Fatal("Error loading the configuration file")
		return err
	}

	//Load Well
	pW.wells, err = LoadWell(countwell)
	if err != nil {
		pW.Log.Fatal("Error loading the wells information")
		return err
	}
	//EvetBus create
	evnt := bus.Neweventbussimple(countwell)
	//Make Steam array
	var steams [countwell]steam.SteamI2 //steam.SteamCsv
	buildSteam(steams, pW.wells, pW.Log, durat, countwell)
	for i, _ := range steams {
		evnt.AddWell(steams[i].(*steam.SteamCsv).Id, 50) // 50 - read buffer (chanel)
	}
	//StaemDataCh := make(chan nt.ScapeDataD, countwell)
	ErrSteam := make(chan error, countwell)
	DoneSteam := make(chan struct{}, countwell)
	//Start csv steam
	runSteam(steams, DoneSteam, ErrSteam, evnt)
	//
	var determineEls [countDetermiElementary]determineElementary.determineElI
	buildDetermineEl(determineEls, pW.wells, durat, countDetermiElementary)

	//Waiting end SteamCsv
	for i := 1; i <= countwell; i++ {
		steams[i].(*steam.SteamCsv).Wg.Wait()
	}
	robin := &Roundrobin{
		countclients: countwell,
		countworker:  countDetermiElementary,
	}
	robin.Init()
	//save repo skip
	//start determineElementary
	return nil
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
func runSteam(steams [countwell]steam.SteamI2,
	DoneSteam chan struct{}, ErrSteam chan error, evnt *bus.Eventbus) {
	for i := 1; i <= countwell; i++ {
		n := i
		go func(k int) {
			//steams[i].(*steam.SteamCsv).ScapeDataCh = steams[k].ReadCsv(DoneSteam, ErrSteam)
			for v := range steams[k].ReadSteam(DoneSteam, ErrSteam) {
				evnt.Send("Save", steams[i].(*steam.SteamCsv).Id, &v)
				evnt.Send("Determine", steams[i].(*steam.SteamCsv).Id, &v)
				//arwells[k].Data <- v
				//to-do check ErrSteam
			}
		}(n)
	}
}
func buildSteam(steams [countwell]steam.SteamI2, arwells Wells, l *logrus.Logger, durat int, countwell int) {
	for i := 1; i <= countwell; i++ {
		id1 := fmt.Sprintf("%d", arwells[i].id)
		steams[i] = &steam.SteamCsv{
			Id:       id1,
			FilePath: arwells[i].path,
			Dur:      time.Second * time.Duration(durat), // max time duration reading
			Log:      l,
		}
	}
}
func LoadWell(count int) ([]Well, error) {
	awells := make([]Well, count, count)
	for i := 1; i <= count; i++ {
		awells := append(awells,
			Well{
				int64(i),
				fmt.Sprintf("Well%d", i),
				"",
				make(chan nt.ScapeDataD),
			})
	}
	awells[1].path = ""
	return awells, nil
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
	plainFormatter.TimestampFormat = "2006-01-02 15:04:05"
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
