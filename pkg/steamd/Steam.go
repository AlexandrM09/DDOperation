package steamd

import (
	"archive/zip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"

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
	//SteamI2 basic interface for operations recognition variant two
	SteamI2 interface {
		ReadSteam(Done chan struct{}, ErrCh chan error) chan nt.ScapeDataD
		ReadSteamTime(Done chan struct{}, ErrCh chan error) chan nt.ScapeDataD
	}

	//SteamRND test steam
	SteamRND struct{}
)

func (St *SteamRND) Read(ScapeDataCh chan nt.ScapeDataD, DoneCh, Done chan struct{}, ErrCh chan error) {
	fmt.Println("RND")
	return
}

//SteamCsv Steam for csv example files
type SteamCsv struct {
	FilePath  string
	StartTime string
	tm        time.Time
	bTime     bool
	Dur       time.Duration
	Log       *logrus.Logger
	Id        string
	Out       chan nt.ScapeDataD
	//	Wg        *sync.WaitGroup
	IdOut string
}

//ReadCsvTime steam SteamI2 for time
func (St *SteamCsv) ReadSteamTime(Done chan struct{}, ErrCh chan error) chan nt.ScapeDataD {
	Out := make(chan nt.ScapeDataD)
	ScapeDataChInside := make(chan nt.ScapeDataD)
	DoneInside := make(chan struct{})
	timer1 := time.NewTimer(St.Dur)
	go func() {
		defer func() {
			close(Out)
			close(ScapeDataChInside)
			close(DoneInside)
		}()
		// !!!!nead add new done chanel for exit St.Read
		go St.Read(ScapeDataChInside, DoneInside, Done, ErrCh)
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

//ReadCsv steam SteamI2
func (St *SteamCsv) ReadSteam(Done chan struct{}, ErrCh chan error) chan nt.ScapeDataD {
	Out := make(chan nt.ScapeDataD)
	DoneInside := make(chan struct{})
	go func() {
		defer func() {
			close(Out)
			close(DoneInside)
		}()
		// !!!!nead add new done chanel for exit St.Read
		go St.Read(Out, DoneInside, Done, ErrCh)
		//St.Wg.Add(1)
		for {
			select {
			//case <-ErrCh:
			//	return
			case <-DoneInside:
				{
					//	St.Wg.Done()
					return
				}
			default:
			}
		}
	}()
	return Out
}
func (St *SteamCsv) Read(ScapeDataCh chan nt.ScapeDataD, DoneCh chan struct{}, Done chan struct{}, ErrCh chan error) {
	defer func() {

		DoneCh <- struct{}{}
		St.Log.Info("Exit csv Steam")
	}()

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
				St.Log.Debugf("id=%s sending in chanel line %d, time:%s ", St.Id, n, ScapeData.Time.Format("2006-01-02 15:04:05"))
				ScapeDataCh <- ScapeData

			}
		} else {
			St.Log.Debugf("2variant id=%s sending in chanel line %d, time:%s ", St.Id, n, ScapeData.Time.Format("2006-01-02 15:04:05"))
			ScapeDataCh <- ScapeData

		}
		n++

		//time.Sleep(time.Millisecond * 100)
	}
	return
}

//
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

//return time from csv
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
