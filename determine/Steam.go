package determine

import (
	"fmt"
)
type SteamRND struct{}
func (St *SteamRND) Read(d *DrillDataType) error{
	fmt.Println("RND")
    return nil
}