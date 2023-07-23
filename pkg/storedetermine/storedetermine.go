package storedetermine

import (
	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	// logrus "github.com/sirupsen/logrus"
)

type (
	//SaveDeteElementary

	Storedetermine struct {
		dElementary map[string]*nt.SaveDetElementary
		dSumm       map[string]*nt.SummaryResult
	}
)

func New() Storedetermine {
	return Storedetermine{
		dElementary: make(map[string]*nt.SaveDetElementary, 10),
		dSumm:       make(map[string]*nt.SummaryResult),
	}
}

// DElementaryGetAll..
func (d *Storedetermine) DElementaryGetAll() map[string]*nt.SaveDetElementary {
	return d.dElementary
}

// DElementaryGet..
func (d *Storedetermine) DElementaryGet(id string) (*nt.SaveDetElementary, bool) {
	v, ok := d.dElementary[id]
	return v, ok
}

// DElementarySet..
func (d *Storedetermine) DElementarySet(id string, v *nt.SaveDetElementary) {
	d.dElementary[id] = v
}

// DSummGet..
func (d *Storedetermine) DSummGet(id string) (*nt.SummaryResult, bool) {
	v, ok := d.dSumm[id]
	return v, ok
}

// DSummSet..
func (d *Storedetermine) DSummSet(id string, v *nt.SummaryResult) {
	d.dSumm[id] = v
}

// DSummGetAll
func (d *Storedetermine) DSummGetAll() map[string]*nt.SummaryResult {
	return d.dSumm
}
