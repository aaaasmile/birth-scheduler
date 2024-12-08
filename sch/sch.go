package sch

import (
	"birthsch/conf"
	"birthsch/idl"
	"birthsch/mail"
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
	nextBirthday    []*idl.SchedNextItem
	nextAnniversary []*idl.SchedNextItem
	mail_simulation bool
}

func RunService(configfile string) error {

	if _, err := conf.ReadConfig(configfile); err != nil {
		return err
	}

	chShutdown := make(chan struct{}, 1)
	go func(chs chan struct{}) {
		sch := Scheduler{datafileName: conf.Current.DataFileName,
			mail_simulation: conf.Current.SimulateMail,
		}
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
	last_month := time.Month(1)
	last_year := 0
	for {
		time.Sleep(60 * time.Second)

		now := time.Now()
		if now.Year() > last_year {
			log.Println("year change")
			last_year = now.Year()
			last_month = time.Month(1)
			last_day = 0
		}
		if now.Month() > last_month {
			log.Println("month change")
			last_month = now.Month()
			last_day = 0
		}
		if now.Day() > last_day {
			log.Println("day change")
			last_day = now.Day()
			if err := sch.reschedule(); err != nil {
				return err
			}
		}
		if sch.hasItems() && now.Hour() > 9 {
			log.Println("time to send an alarm ", now)
			if err := sch.sendItemsAlarm(); err != nil {
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
	sch.nextBirthday = make([]*idl.SchedNextItem, 0)
	sch.nextAnniversary = make([]*idl.SchedNextItem, 0)
	now := time.Now()
	log.Println("Schedule next for ", now)

	yy := now.Year()
	for _, item := range schList.List {
		tmp_arr := strings.Split(item.MonthDay, "-")
		if len(tmp_arr) != 2 {
			return fmt.Errorf("expect month-day format, but get %s", item.MonthDay)
		}
		mm, err := monthToMonth(tmp_arr[0])
		if err != nil {
			return err
		}
		dd, err := strconv.Atoi(tmp_arr[1])
		if err != nil {
			return err
		}
		time_item := time.Date(yy, mm, dd, 23, 59, 0, 0, time.Local)
		if time_item.Unix() > now.Unix() {
			nextItem := idl.SchedNextItem{Name: item.Name, Note: item.Note, Time: time_item}
			err = nextItem.SetEventType(item.Type)
			//log.Println("check ", nextItem)
			if err != nil {
				return err
			}
			if now.Day() == time_item.Day() {
				if nextItem.EventType == idl.Birthday {
					sch.nextBirthday = append(sch.nextBirthday, &nextItem)

				}
				if nextItem.EventType == idl.Anniversary {
					sch.nextAnniversary = append(sch.nextAnniversary, &nextItem)
				}
			}
		}
	}
	found := false
	if len(sch.nextBirthday) > 0 {
		log.Println("Next birthday ", sch.nextBirthday[0])
		found = true
	}
	if len(sch.nextAnniversary) > 0 {
		log.Println("Next anniversary ", sch.nextAnniversary)
		found = true
	}
	if !found {
		log.Println("Nothing found for today ", now)
	}
	return nil
}

func monthToMonth(s string) (time.Month, error) {
	switch s {
	case "Gen":
		return time.Month(1), nil
	case "Feb":
		return time.Month(2), nil
	case "Mar":
		return time.Month(3), nil
	case "Apr":
		return time.Month(4), nil
	case "Mag":
		return time.Month(5), nil
	case "Giu":
		return time.Month(6), nil
	case "Lug":
		return time.Month(7), nil
	case "Ago":
		return time.Month(8), nil
	case "Set":
		return time.Month(9), nil
	case "Ott":
		return time.Month(10), nil
	case "Nov":
		return time.Month(11), nil
	case "Dic":
		return time.Month(12), nil

	}
	return 0, fmt.Errorf("month not recongnized %s", s)
}

func (sch *Scheduler) hasItems() bool {
	if len(sch.nextAnniversary) > 0 {
		return true
	}
	if len(sch.nextBirthday) > 0 {
		return true
	}
	return false
}

func (sch *Scheduler) sendItemsAlarm() error {
	if len(sch.nextBirthday) > 0 {
		if err := sch.sendBirthdayAlarm(); err != nil {
			return err
		}
	}
	if len(sch.nextAnniversary) > 0 {
		if err := sch.sendAnniversaryAlarm(); err != nil {
			return err
		}
	}
	return nil
}

func (sch *Scheduler) sendBirthdayAlarm() error {
	mail := mail.MailSender{}
	mail.FillConf(sch.mail_simulation)
	templFileName := "templates/birthday-mail.html"
	if err := mail.BuildEmailMsg(templFileName, sch.nextBirthday); err != nil {
		return err
	}
	if err := mail.SendEmailViaRelay(); err != nil {
		return err
	}
	sch.nextBirthday = make([]*idl.SchedNextItem, 0)

	return nil
}

func (sch *Scheduler) sendAnniversaryAlarm() error {
	mail := mail.MailSender{}
	mail.FillConf(sch.mail_simulation)
	templFileName := "templates/anniversary-mail.html"
	if err := mail.BuildEmailMsg(templFileName, sch.nextAnniversary); err != nil {
		return err
	}
	if err := mail.SendEmailViaRelay(); err != nil {
		return err
	}

	sch.nextAnniversary = make([]*idl.SchedNextItem, 0)
	return nil
}
