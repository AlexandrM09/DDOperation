package steamd

import (
	"archive/zip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sync"

	"strconv"
	"strings"
	"time"

	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	"github.com/sirupsen/logrus"
)

type (
	//SteamI basic interface for operations recognition
	SteamI interface {
		Read(ScapeDataCh chan nt.ScapeDataD, DoneCh, Done chan struct{}, ErrCh chan error)
	}

	//SteamRND test steam
	SteamRND struct{}
)

func (St *SteamRND) Read(ScapeDataCh chan nt.ScapeDataD, DoneCh, Done chan struct{}, ErrCh chan error) {
	fmt.Println("RND")
	
}

// SteamCsv Steam for csv example files
type SteamCsv struct {
	FilePath    string
	StartTime   string
	tm          time.Time
	bTime       bool
	Dur         time.Duration
	Log         *logrus.Logger
	Id          string
	Out         chan nt.ScapeDataD
	Wg          *sync.WaitGroup
	IdOut       string
	done        chan struct{}
	start       time.Time
	alldataread bool
}

// Stop..
func (St *SteamCsv) Stop() {
	defer close(St.done)
	St.Log.Debug("send  done chanel ", St.Id)
	if St.alldataread {
		return
	}
	St.done <- struct{}{}
	St.Wg.Wait()

}
func New(id string, filePath string,
	dur time.Duration, // max time duration reading
	l *logrus.Logger,
	Wg *sync.WaitGroup) *SteamCsv {

	return &SteamCsv{
		Id:          id,
		FilePath:    filePath,
		Dur:         dur, // max time duration reading
		Log:         l,
		Wg:          Wg,
		Out:         make(chan nt.ScapeDataD),
		done:        make(chan struct{}),
		alldataread: false,
	}
}

// ReadCsvTime steam SteamI2 for time
func (St *SteamCsv) ReadSteamTime(ErrCh chan error) chan nt.ScapeDataD {
	Out := make(chan nt.ScapeDataD)
	ScapeDataChInside := make(chan nt.ScapeDataD)
	DoneInside := make(chan struct{})
	timer1 := time.NewTimer(St.Dur)
	St.Wg.Add(1)
	go func() {
		defer func() {
			close(Out)
			close(ScapeDataChInside)
			close(DoneInside)
			St.Wg.Done()
			St.Log.Debug("WTF Before close done chanel ", St.Id)

		}()
		// !!!!nead add new done chanel for exit St.Read
		go St.Read(ScapeDataChInside, DoneInside, St.done, ErrCh)
		//St.Wg.Add(1)
		for {
			select {
			case <-timer1.C:
				{
					select {
					case Out <- <-ScapeDataChInside: //St.Log.Debugf("SendScapeData")

					default:
					}
				}
			//case <-ErrCh:
			//	{
			//		timer1.Stop()
			//		return
			//	}
			case <-DoneInside:
				{
					timer1.Stop()
					//St.Wg.Done()
					return
				}
			default:
			}
		}
	}()
	return Out
}

// ReadCsv steam SteamI2
func (St *SteamCsv) ReadSteam(ErrCh chan error) chan nt.ScapeDataD {
	Out := make(chan nt.ScapeDataD)
	DoneInside := make(chan struct{})
	St.Log.Debug("After Steam make done chanel ", St.Id)
	St.Wg.Add(1)
	go func() {
		defer func() {
			close(Out)
			close(DoneInside)
			St.Log.Debug("defer main steam  func", St.Id)

			St.Wg.Done()

		}()
		// !!!!nead add new done chanel for exit St.Read
		go St.Read(Out, DoneInside, St.done, ErrCh)
		//St.Wg.Add(1)

		 <-DoneInside
			{
				//	St.Wg.Done()
				St.Log.Debug("After DoneInside reading ", St.Id)
				St.Log.Debug("exit main steam  func ", St.Id)
				return
			}
		

	}()
	return Out
}
func (St *SteamCsv) Read(ScapeDataCh chan nt.ScapeDataD, DoneCh chan struct{}, Done chan struct{}, ErrCh chan error) {
	defer func() {
		St.Log.Info("Exit csv Steam ", St.Id, ", total working time(s):", time.Since(St.start).Seconds())
		DoneCh <- struct{}{}

	}()
	St.start = time.Now()
	St.Log.Infof("Start steam from csv file path:%s", St.FilePath)
	//time.Sleep(time.Second * 4)
	var err error
	St.tm, err = time.Parse("2006-01-02 15:04:05", St.StartTime)
	St.bTime = err == nil
	//fmt.Printf("parse time=%v \n", St.bTime)
	var ScapeData nt.ScapeDataD
	ScapeData.Id = St.Id
	sH := scapeHeader{}
	r, err := zip.OpenReader(St.FilePath) //"../source/source.zip"
	if err != nil {
		St.Log.Fatal("id=%s unpacking zip file path", St.Id, err)
		ErrCh <- err

		return
	}
	defer r.Close()
	csvFile := r.File[0]
	rc, errf := csvFile.Open()
	defer rc.Close()
	if errf != nil {
		St.Log.Fatal("id=%s not open file, ", St.Id, errf)
		ErrCh <- err
		return
	}
	St.Log.Infof("id=%s create new reader zip file:%s\n", St.Id, St.FilePath)
	reader := csv.NewReader(rc)
	reader.Comma = ';'
	n := 0
	for {
		//====for exit
		select {
		case <-Done:
			{
				St.Log.Info("id=%s on-demand output csv Steam", St.Id)
				return
			}
		default:
			{
			}
		}
		//===========
		line, error := reader.Read()
		if error == io.EOF { //|| (n > 5)
			break
		} else if error != nil {
			St.Log.Fatal("read line, ", error)
			ErrCh <- err
			return
		}
		if n == 0 {
			n = 1
			err := parseheader(line, &sH)
			if err != nil {
				St.Log.Fatal("parse header, ", err)
				ErrCh <- err
				return
			}
			continue
		}
		err = nil
		ScapeData.Time, err = GetTime(line)
		if !(err == nil) {
			continue
		}
		len := len(line)
		for i := 2; i < 20; i++ {
			if (sH[i] > 2) && (len > sH[i]) {
				if f1, err := strconv.ParseFloat(line[sH[i]], 32); err == nil {
					ScapeData.Values[i] = float32(f1)
				}
			}
		}
		ScapeData.Count = n
		if St.bTime {
			if ScapeData.Time.Sub(St.tm) > 0 {
				// St.Log.Debugf("id=%s sending in chanel line %d, time:%s ", St.Id, n, ScapeData.Time.Format("2006-01-02 15:04:05"))
				ScapeDataCh <- ScapeData

			}
		} else {
			// St.Log.Debugf("2variant id=%s sending in chanel line %d, time:%s ", St.Id, n, ScapeData.Time.Format("2006-01-02 15:04:05"))
			ScapeDataCh <- ScapeData

		}
		n++

		//time.Sleep(time.Millisecond * 100)
	}
	St.alldataread = true
	St.Log.Info("All Data read ", St.Id, ", total working time(s):", time.Since(St.start).Seconds())
	return
}

type scapeHeader [20]int

// parse csv header
func parseheader(record []string, sH *scapeHeader) error {
	h := [20]string{"", "", "S101", "S115", "S300", "S1001", "S202", "S200", "S1300", "S1200", "S111", "S110", "", "", "", "", "", "", "", ""}
	for i := 2; i < 20; i++ {
		if h[i] != "" {
			n := findColumn(h[i], record)
			sH[i] = n

		}
	}
	return nil
}

func findColumn(sourceString string, record []string) int {
	for i := 2; i < len(record); i++ {
		if record[i] == sourceString {
			return i
		}
	}
	return 0
}

// return time from csv
func GetTime(record []string) (time.Time, error) {
	//func Date(year int, month Month, day, hour, min, sec, nsec int, loc *Location) Time
	if len(record) < 3 {
		return time.Now(), errors.New("bad format csv")
	}
	//30.09.2017
	parseDate := strings.Split(record[0], ".")
	year, err1 := strconv.Atoi(parseDate[2])
	if err1 != nil {
		return time.Now(), errors.New("bad format csv")
	}
	monthint, err2 := strconv.Atoi(parseDate[1])
	if err2 != nil {
		return time.Now(), errors.New("bad format csv")
	}
	month := time.Month(monthint)
	day, err3 := strconv.Atoi(parseDate[0])
	if err3 != nil {
		return time.Now(), errors.New("bad format csv")
	}
	if len(parseDate) != 3 {
		return time.Now(), errors.New("bad format csv")
	}
	n, err := strconv.Atoi(record[1])
	if err != nil {
		return time.Now(), errors.New("bad format csv")
	}

	h := n / (1000 * 3600)
	n = n - h*(1000*3600)
	min := n / (1000 * 60)
	n = n - min*(1000*60)
	sec := n / 1000
	n = n - sec*1000

	return time.Date(year, month, day, h, min, sec, n, time.UTC), nil
}
