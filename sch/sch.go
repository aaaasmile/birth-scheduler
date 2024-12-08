package sch

import (
	"birthsch/conf"
	"birthsch/idl"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

func RunService(configfile string) error {

	if _, err := conf.ReadConfig(configfile); err != nil {
		return err
	}

	chShutdown := make(chan struct{}, 1)
	go func(chs chan struct{}) {
		if err := doSchedule(); err != nil {
			log.Println("Server is not scheduling anymore: ", err)
			chs <- struct{}{}
		}
	}(chShutdown)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	log.Println("Enter in server blocking loop")

loop:
	for {
		select {
		case <-sig:
			log.Println("stop because interrupt")
			break loop
		case <-chShutdown:
			log.Println("stop because service shutdown on scheduling")
			log.Fatal("Force with an error to restart the service")
		}
	}

	log.Println("Bye, service")
	return nil
}

type Scheduler struct {
	nextEvents []idl.SchedNextItem
}

func doSchedule() error {
	sch := Scheduler{}
	if err := sch.readDataJsonFile(conf.Current.DataFileName); err != nil {
		return err
	}

	log.Println("Infinite scheduler loop")
	a := 0
	for {
		a++
		time.Sleep(1 * time.Second)
		if a > 10 {
			return fmt.Errorf("scheduler crash")
		}
	}
}

func (sch *Scheduler) readDataJsonFile(fname string) error {
	log.Println("load scheduler json data ", fname)
	if fname == "" {
		return fmt.Errorf("data file is empty")
	}
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	schList := idl.SchedList{}

	err = json.NewDecoder(f).Decode(&schList)
	if err != nil {
		return err
	}
	log.Println("Loaded scheduler from file ", fname, schList)
	return sch.scheduleNext(&schList)
}

func (sch *Scheduler) scheduleNext(schList *idl.SchedList) error {
	sch.nextEvents = make([]idl.SchedNextItem, 0)
	now := time.Now()
	yy := now.Year()
	for _, item := range schList.List {
		tmp_arr := strings.Split(item.MonthDay, "-")
		if len(tmp_arr) != 2 {
			return fmt.Errorf("expect month-day format, but get %s", item.MonthDay)
		}
		mmi, err := strconv.Atoi(tmp_arr[0])
		if err != nil {
			return err
		}
		dd, err := strconv.Atoi(tmp_arr[1])
		if err != nil {
			return err
		}
		mm := time.Month(mmi)
		time_item := time.Date(yy, mm, dd, 23, 59, 0, 0, time.Local)
		if time_item.Unix() > now.Unix() {
			nextItem := idl.SchedNextItem{Name: item.Name, Note: item.Note, Time: time_item}
			err = nextItem.SetEventType(item.Type)
			if err != nil {
				return err
			}
			sch.nextEvents = append(sch.nextEvents, nextItem)
		}
	}
	log.Println("Next events ", sch.nextEvents)
	return nil
}
