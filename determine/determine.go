package determine

/*func (d *drillData) read() error{
    return nil
}
*/

type (
	listoperation []string
	determineOne  interface {
		check(d *DrillDataType) string
		getname() string
	}
	SteamI interface {
		Read(d *DrillDataType)
	}
	Determine struct {
		Data        *DrillDataType
		Steam       SteamI
		ListCheck   []determineOne
		activecheck determineOne
	}
)

func (dt *Determine) Start() error {
	go dt.Steam.Read(dt.Data)
	//var err error
	go func() { _ = dt.Run(dt.Data) }()
	return nil
}

// Main dispath function in list
func (dt *Determine) Run(d *DrillDataType) error {
	var resSt string
	var n int
	for {
		resSt = ""
		select {
		case <-d.DoneCh:
			{ //close all
				return nil
			}
		case err := <-d.ErrCh:
			return err
		default:
			for i := 0; i < len(dt.ListCheck) && (resSt == ""); i++ {
				resSt = dt.ListCheck[i].check(dt.Data)
			}
			n = dt.findbyName(resSt)
			if n >= 0 {
				resSt = dt.ListCheck[n].check(dt.Data)
			}
		}
	}
}

//find by name check
func (dt *Determine) findbyName(s string) int {

	for i := 1; i < len(dt.ListCheck); i++ {
		if s == dt.ListCheck[i].getname() {
			return i
		}
	}
	return -1
}

// Create new List determine
func NewDetermine(ds *DrillDataType, sm SteamI) *Determine {
	return &Determine{Data: ds,
		Steam:     sm,
		ListCheck: []determineOne{}}
}

func GetList() listoperation {
	return []string{"First operation"}
}
