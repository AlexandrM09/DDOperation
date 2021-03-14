package balancingservices

import (
	"fmt"
	"time"

	nt "github.com/AlexandrM09/DDOperation/pkg/sharetype"
	steam "github.com/AlexandrM09/DDOperation/pkg/steamd"
)

type (
	well = struct {
		id   int64
		name string
		path string
		Data chan nt.ScapeDataD
	}
	wells  = []well
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

func Building() {
	fmt.Println("Start")
	//Load
	arwells := LoadWell(countwell)
	//Make
	var steams [countwell]steam.SteamI2 //steam.SteamCsv
	for i := 1; i <= countwell; i++ {
		id1 := fmt.Sprintf("%d", arwells[i].id)
		steams[i] = &steam.SteamCsv{
			Id:       id1,
			FilePath: arwells[i].path,
			Dur:      time.Second * 300, // max time duration reading
		}
	}
	//StaemDataCh := make(chan nt.ScapeDataD, countwell)
	ErrSteam := make(chan error, countwell)
	DoneSteam := make(chan struct{}, countwell)
	//Start csv steam
	for i := 1; i <= countwell; i++ {
		n := i
		go func(k int) {
			//steams[i].(*steam.SteamCsv).ScapeDataCh = steams[k].ReadCsv(DoneSteam, ErrSteam)
			for v := range steams[k].ReadCsv(DoneSteam, ErrSteam) {
				arwells[k].Data <- v
			}
		}(n)
	}
	//
	robin := &Roundrobin{
		countclients: countwell,
		countworker:  countDetermiElementary,
	}
	robin.Init()
	//save repo skip
	//start determineElementary

}
func LoadWell(count int) wells {
	awells := make(wells, count, count)
	for i := 1; i <= count; i++ {
		awells := append(awells,
			well{
				int64(i),
				fmt.Sprintf("Well%d", i),
				"",
				make(chan nt.ScapeDataD)})
	}
	awells[1].path = ""
	return awells
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
