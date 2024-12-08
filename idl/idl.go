package idl

var (
	Appname = "birthday-sch"
	Buildnr = "00.001.20241208-00"
)

type SchedItem struct {
	Name     string
	MonthDay string
	Type     string
	Note     string
}

type SchedList struct {
	List []SchedItem
}
