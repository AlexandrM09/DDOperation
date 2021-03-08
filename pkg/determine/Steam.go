package determine

import (
	"archive/zip"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

//SteamRND test steam
type SteamRND struct{}

func (St *SteamRND) Read(ScapeDataCh chan ScapeDataD, DoneCh chan struct{}) {
	fmt.Println("RND")
	return
}

//SteamCsv Steam for csv example files
type SteamCsv struct {
	FilePath   string
	SatartTime string
	tm         time.Time
	bTime      bool
}

func (St *SteamCsv) Read(ScapeDataCh chan ScapeDataD, DoneCh chan struct{}) {
	defer func() {

		DoneCh <- struct{}{}

	}()

	var err error
	St.tm, err = time.Parse("2006-01-02 15:04:05", St.SatartTime)
	St.bTime = err == nil
	//fmt.Printf("parse time=%v \n", St.bTime)
	var ScapeData ScapeDataD
	sH := scapeHeader{}
	r, err := zip.OpenReader(St.FilePath) //"../source/source.zip"
	if err != nil {
		log.Fatal("unpacking zip file", err)
	}
	defer r.Close()
	csvFile := r.File[0]
	rc, errf := csvFile.Open()
	defer rc.Close()
	if errf != nil {
		log.Fatal("not open, ", errf)
	}
	fmt.Printf("create new reader zip file:%s\n", St.FilePath)
	reader := csv.NewReader(rc)
	reader.Comma = ';'
	n := 0
	for {
		if (n == 0) || (n%1000 == 0) {
		}
		line, error := reader.Read()
		if error == io.EOF { //|| (n > 5)
			break
		} else if error != nil {

			log.Fatal("read line, ", error)
		}
		if n == 0 {
			n = 1
			err := parseheader(line, &sH)
			if err != nil {

				log.Fatal("parse header, ", err)
			}
			continue
		}
		ScapeData.Time, _ = getTime(line)
		len := len(line)
		for i := 2; i < 20; i++ {
			if (sH[i] > 2) && (len > sH[i]) {
				if f1, err := strconv.ParseFloat(line[sH[i]], 32); err == nil {
					ScapeData.Values[i] = float32(f1)
				}
			}
		}
		if St.bTime {
			if ScapeData.Time.Sub(St.tm) > 0 {
				ScapeDataCh <- ScapeData
			}
		} else {
			ScapeDataCh <- ScapeData
		}
		n++
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
func getTime(record []string) (time.Time, error) {
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