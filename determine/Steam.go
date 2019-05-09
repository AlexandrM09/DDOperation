package determine

import (
	"fmt"
)
type SteamRND struct{}
func (St *SteamRND) Read(d *DrillDataType) {
	fmt.Println("RND")
    return 
}