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

type Scheduler struct {
	datafileName    string
	nextBirthday    []idl.SchedNextItem
	nextAnniversary []idl.SchedNextItem
}

func RunService(configfile string) error {

	if _, err := conf.ReadConfig(configfile); err != nil {
		return err
	}

	chShutdown := make(chan struct{}, 1)
	go func(chs chan struct{}) {
		sch := Scheduler{datafileName: conf.Current.DataFileName}
		if err := sch.doSchedule(); err != nil {
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

func (sch *Scheduler) doSchedule() error {
	log.Println("Infinite scheduler loop")
	last_day := 0
	for {
		now := time.Now()
		time.Sleep(1 * time.Second)
		if now.Day() > last_day {
			log.Println("day change")
			last_day = now.Day()
			if err := sch.reschedule(); err != nil {
				return err
			}
		}
	}
}

func (sch *Scheduler) reschedule() error {
	schList, err := sch.readDataJsonFile()
	if err != nil {
		return err
	}
	return sch.scheduleNext(schList)
}

func (sch *Scheduler) readDataJsonFile() (*idl.SchedList, error) {
	fname := sch.datafileName
	log.Println("load scheduler json data ", fname)
	if fname == "" {
		return nil, fmt.Errorf("data file is empty")
	}
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	schList := idl.SchedList{}

	err = json.NewDecoder(f).Decode(&schList)
	if err != nil {
		return nil, err
	}
	log.Println("Loaded scheduler from file ", fname, schList)
	return &schList, nil
}

func (sch *Scheduler) scheduleNext(schList *idl.SchedList) error {
	sch.nextBirthday = make([]idl.SchedNextItem, 0)
	sch.nextAnniversary = make([]idl.SchedNextItem, 0)

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
			if nextItem.EventType == idl.Birthday {
				sch.nextBirthday = append(sch.nextBirthday, nextItem)
			}
			if nextItem.EventType == idl.Anniversary {
				sch.nextAnniversary = append(sch.nextAnniversary, nextItem)
			}
		}
	}
	if len(sch.nextBirthday) > 0 {
		log.Println("Next birthday ", sch.nextBirthday)
	}
	if len(sch.nextAnniversary) > 0 {
		log.Println("Next anniversary ", sch.nextAnniversary)
	}
	return nil
}
