package determine

import (
	"fmt"
)
type SteamRND struct{}
func (St *SteamRND) Read(ScapeDataCh chan ScapeDataD, DoneCh chan struct{}) {
	fmt.Println("RND")
    return 
}