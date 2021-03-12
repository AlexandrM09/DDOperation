package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	dtm "github.com/AlexandrM09/DDOperation/pkg/determine"
	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	steam "github.com/AlexandrM09/DDOperation/pkg/steamd"

	logrus "github.com/sirupsen/logrus"
	_ "gopkg.in/yaml.v2"
)

func main() {
	Cfg := nt.ConfigDt{}
	errf := dtm.LoadConfigYaml("config.yaml", &Cfg)
	if errf != nil {
		log.Fatal("not load config file")
	}
	sr := nt.DrillDataType{
		Log: createLog(logrus.DebugLevel),
		Cfg: &Cfg,
	}
	tm := dtm.NewDetermine(&sr, &steam.SteamCsv{
		FilePath:   "./source/source2.zip",
		SatartTime: "___2019-05-25 17:52:43",
		Log:        sr.Log,
	})
	dur, err := tm.Start(60)
	tempt, _ := time.Parse("15:04:01", "00:00:00")
	fmt.Printf("duration:%s,result err:%v\n", tempt.Add(dur).Format("15:04:00.000"), err)
	if err != nil {
		log.Fatal("error:time limit exceeded")
	}
	data2 := tm.GetSummarysheet()
	fmt.Printf("Start print Summarysheet len=%v \n", len(data2))
	for i := 0; i < len(data2); i++ {
		fmt.Println(dtm.FormatSheet(data2[i]))
		for j := 0; j < len(data2[i].Details); j++ {
			fmt.Println(dtm.FormatSheetDetails(data2[i].Details[j]))
		}
	}
	fmt.Printf("Start print Summarysheet short form len=%v \n", len(data2))
	fmt.Println(data2[0].Sheet.StartData.Time.Format("2006-01-02"))
	for i := 0; i < len(data2); i++ {
		fmt.Println(dtm.FormatSheet2(data2[i]))
	}
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
	plainFormatter.TimestampFormat = "2006-01-02 15:04:05"
	plainFormatter.LevelDesc = []string{"PANC", "FATL", "ERRO", "WARN", "INFO", "DEBG"}
	var log = logrus.New()
	log.SetReportCaller(true)
	log.SetFormatter(plainFormatter)
	//&logrus.TextFormatter{
	//	CallerPrettyfier: func(f *runtime.Frame) (string, string) {
	//			_, filename := path.Split(f.File)
	//			filename = fmt.Sprintf("%s:%d", filename, f.Line)
	//			return "", filename
	//		},

	//FieldMap: logrus.FieldMap{
	//	logrus.FieldKeyLevel: "[@level]",
	//	logrus.FieldKeyTime:  "@timestamp",
	//	logrus.FieldKeyFile:  "@filename",
	//	logrus.FieldKeyMsg:   "@message"},
	//DisableColors: false,

	//&easy.Formatter{
	//	TimestampFormat: "2006-01-02 15:04:05",
	//	LogFormat:       "[%lvl%]: %time% - %msg%",
	//	},
	//	FullTimestamp: true,
	//)

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
