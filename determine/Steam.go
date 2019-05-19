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

type SteamRND struct{}

func (St *SteamRND) Read(ScapeDataCh chan ScapeDataD, DoneCh chan struct{}) {
	fmt.Println("RND")
	return
}

// Steam for csv example files
type SteamCsv struct{}

func (St *SteamCsv) Read(ScapeDataCh chan ScapeDataD, DoneCh chan struct{}) {
	defer func() {
		DoneCh <- struct{}{}

	}()
	//nothing
	//v1 := [20]float32{0, 0, 100, 90, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	var ScapeData ScapeDataD
	sH := scapeHeader{}

	r, err := zip.OpenReader("../source/source.zip")
	if err != nil {
		log.Fatal("unpacking zip file", err)
	}
	defer r.Close()

	csvFile := r.File[0]
	fmt.Printf("open csv \n")
	//csvFile, errf := os.Open("../source/03261652_small.csv")
	rc, errf := csvFile.Open()
	defer rc.Close()
	if errf != nil {
		log.Fatal("not open, ", errf)
	}
	fmt.Printf("create new reader \n")
	//reader := csv.NewReader(bufio.NewReader(csvFile))
	reader := csv.NewReader(rc)
	reader.Comma = ';'
	n := 0
	fmt.Printf("start for \n")
	for {
		fmt.Printf("n= %v \n", n)
		line, error := reader.Read()
		if error == io.EOF { //|| (n > 5)
			close(ScapeDataCh)
			fmt.Printf("close(ScapeDataCh) \n")
			break
		} else if error != nil {

			log.Fatal("read line, ", error)
		}
		if n == 0 {
			n = 1
			fmt.Printf("start parse csv header \n")
			//spl:=strings.Split(line[0], ";")
			err := parseheader(line, &sH)
			//fmt.Printf("line %v \n", line)

			fmt.Printf("sH %v \n", sH)
			if err != nil {

				log.Fatal("parse header, ", err)
			}
			continue
		}
		/*  people = append(people, Person{
		            Firstname: line[0],
		            Lastname:  line[1],
		            Address: &Address{
		                City:  line[2],
		                State: line[3],
		            },
				})
		*/
		fmt.Printf("read line \n")
		ScapeData.Time, _ = getTime(line)
		//fmt.Printf("Time= %v \n", ScapeData.Time)
		len := len(line)
		fmt.Printf("line len= %d , %v\n", len, line)
		for i := 2; i < 20; i++ {

			if (sH[i] > 2) && (len > sH[i]) {
				fmt.Printf("index %d", sH[i])
				fmt.Printf("read float value= %v \n", line[sH[i]])
				if f1, err := strconv.ParseFloat(line[sH[i]], 32); err == nil {
					ScapeData.Values[i] = float32(f1)
				}
			}
		}
		ScapeDataCh <- ScapeData
		n++

	}

	return
}

//
type scapeHeader [20]int

// parse csv header

func parseheader(record []string, sH *scapeHeader) error {
	fmt.Printf("parse header %v, len= %d\n", record, len(record))
	h := [20]string{"", "", "S101", "S115", "S300", "S1001", "S202", "S200", "S1300", "S1200", "S111", "S110", "", "", "", "", "", "", "", ""}
	for i := 2; i < 20; i++ {
		if h[i] != "" {
			n := findColumn(h[i], record)
			fmt.Printf("find column result i= %d,n= %v, \n", i, n)
			sH[i] = n
			//if sH[2] == 0 {
			//	return errors.New("not found column")
			//}
		}
	}
	return nil
}

func findColumn(sourceString string, record []string) int {
	//	fmt.Printf("find column len= %d \n", len(record))
	for i := 2; i < len(record); i++ {
		//	fmt.Printf("find column i= %d,rec= %s, source %s \n", i, record[i], sourceString)
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
	month_int, err2 := strconv.Atoi(parseDate[1])
	if err2 != nil {
		return time.Now(), errors.New("bad format csv")
	}
	month := time.Month(month_int)
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
	//fmt.Printf("type %T value %d \n",h,h)
	min := n / (1000 * 60)
	n = n - min*(1000*60)
	//fmt.Printf("type %T value %d \n",min,min)
	sec := n / 1000
	n = n - sec*1000
	//fmt.Printf("type %T value %d \n",sec,sec)
	//nsec := n
	//fmt.Printf("type %T value %d \n",msec,msec)
	//hour:=
	return time.Date(year, month, day, h, min, sec, n, time.UTC), nil
}
