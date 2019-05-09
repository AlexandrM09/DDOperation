package determine

/*func (d *drillData) read() error{
    return nil
}
*/

type (
 listoperation []string
 determineOne interface {
    check(d *DrillDataType) string
}
SteamI interface{
    Read(d *DrillDataType) error
}
 Determine struct{
     Data *DrillDataType
     Steam SteamI 
     ListCheck []determineOne
     activecheck determineOne
 }
 
)
func (dt *Determine) Read() error{
     err:=dt.Steam.Read(dt.Data)
     return err}
     
// Create new List determine
func NewDetermine(ds *DrillDataType,sm SteamI) *Determine{
    return &Determine{Data:ds,
        Steam:sm,
        ListCheck:make([]determineOne,1)}
}

func GetList() listoperation{
    return []string{"First operation"} 
}
