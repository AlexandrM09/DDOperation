package algoritmdetermine

import (
	"time"

	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	"github.com/sirupsen/logrus"
)

type (

	//SummarysheetT -type result list
	SummarysheetT struct {
		Sheet   OperationOne
		Details []OperationOne
		Log     *logrus.Logger
		Cfg     *nt.ConfigDt
		Temp    struct {
			LastToolDepht     float32
			LastTimeToolDepht time.Time
			StartDepht        float32
			LastStartData     ScapeDataD
			LastTripData      ScapeDataD
			FlagChangeTrip    int
		}
	}
)
