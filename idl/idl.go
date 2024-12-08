package idl

import (
	"fmt"
	"time"
)

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

type EventType int

const (
	Birthday EventType = iota
	Anniversary
)

type SchedNextItem struct {
	Name      string
	Time      time.Time
	EventType EventType
	Note      string
}

func (sni *SchedNextItem) SetEventType(tt string) error {
	switch tt {
	case "Compl":
		sni.EventType = Birthday
		return nil
	case "Anniv":
		sni.EventType = Anniversary
		return nil
	}
	return fmt.Errorf("type %s not recognized", tt)
}
