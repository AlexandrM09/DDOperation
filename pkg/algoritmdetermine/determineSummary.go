package algoritmdetermine

import (
	"time"

	nt1 "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	"github.com/sirupsen/logrus"
)

type (

	//SummarysheetT -type result list
	SummarysheetT2 struct {
		Sheet   nt1.OperationOne
		Details []nt1.OperationOne
		Log     *logrus.Logger
		Cfg     *nt1.ConfigDt
		Temp    struct {
			LastToolDepht     float32
			LastTimeToolDepht time.Time
			StartDepht        float32
			LastStartData     nt1.ScapeDataD
			LastTripData      nt1.ScapeDataD
			FlagChangeTrip    int
		}
	}
)
