package sch

import (
	"birthsch/conf"
	"birthsch/idl"
	"birthsch/mail"
	"birthsch/telegram"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

type Scheduler struct {
	datafileName    string
	nextBirthday    []*idl.SchedNextItem
	nextAnniversary []*idl.SchedNextItem
	monitoredURL    string
	simulation      bool
	debug           bool
}

func RunService(configfile string, simulate bool) error {

	if _, err := conf.ReadConfig(configfile); err != nil {
		return err
	}

	chShutdown := make(chan struct{}, 1)
	go func(chs chan struct{}) {
		sch := Scheduler{datafileName: conf.Current.DataFileName,
			simulation: (conf.Current.SimulateAlarm || simulate),
			debug:      conf.Current.Debug,
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

func (sch *Scheduler) checkSite() error {
	URL := sch.monitoredURL
	if URL == "" {
		return nil
	}
	log.Println("Check URL for", URL)
	c := colly.NewCollector()
	c.OnHTML("body > main > section.event-hero.bg-mono-darkest.color-brand-primary > div.event-hero__content > div > div > div:nth-child(1) > div > div.event-hero__buttons.mt-5 > p", func(e *colly.HTMLElement) {
		pp := e.Text
		log.Println("Site checked: ", e)
		if !strings.Contains(pp, "Check back soon for entry details on this race") {
			log.Println("Site has changed to: ", pp)
			if err := sch.sendWebChangedAlarm(URL); err != nil {
				log.Println("[OnHTML] error ", err)
			}
		}
	})
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})
	c.OnError(func(e *colly.Response, err error) {
		log.Println("Error on scrap", err)
	})
	c.Visit(URL)

	log.Println("check site done")
	return nil
}

func (sch *Scheduler) doSchedule() error {
	sch.monitoredURL = conf.Current.UrlToCheck

	log.Println("Infinite scheduler loop")
	if sch.monitoredURL != "" {
		log.Println("Url to check is set to ", sch.monitoredURL)
	}
	last_day := 0
	last_month := time.Month(1)
	last_year := 0
	sleeped_time := -1
	for {
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
		if sch.hasItems() && now.Hour() >= 9 {
			log.Println("time to send an alarm ", now)
			if err := sch.sendItemsAlarm(); err != nil {
				return err
			}
		}
		if sleeped_time == -1 || sleeped_time > 3600*6 {
			if err := sch.checkSite(); err != nil {
				return err
			}
			sleeped_time = 0
		}
		time.Sleep(60 * time.Second)
		sleeped_time += 60
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
			if (now.Day() == time_item.Day()) &&
				(now.Month() == time_item.Month()) {
				log.Println("candidate for today alarm", nextItem)
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
	templ := "templates/birthday-mail.html"
	if err := sendEmail(templ, sch.simulation, sch.nextBirthday); err != nil {
		return err
	}
	if err := sendTelegram(templ, sch.simulation, sch.nextBirthday, sch.debug); err != nil {
		return err
	}

	sch.nextBirthday = make([]*idl.SchedNextItem, 0)
	return nil
}

func (sch *Scheduler) sendWebChangedAlarm(URL string) error {
	templ := "templates/webchanged-mail.html"
	if err := sendEmailForWeb(templ, sch.simulation, URL); err != nil {
		return err
	}
	if err := sendTelegramForWeb(templ, sch.simulation, URL, sch.debug); err != nil {
		return err
	}
	sch.monitoredURL = ""

	return nil
}

func (sch *Scheduler) sendAnniversaryAlarm() error {
	templ := "templates/anniversary-mail.html"
	if err := sendEmail(templ, sch.simulation, sch.nextAnniversary); err != nil {
		return err
	}
	if err := sendTelegram(templ, sch.simulation, sch.nextAnniversary, sch.debug); err != nil {
		return err
	}
	sch.nextAnniversary = make([]*idl.SchedNextItem, 0)
	return nil
}

func sendEmail(templFileName string, simulation bool, schItems []*idl.SchedNextItem) error {
	mail := mail.MailSender{}
	mail.FillConf(simulation)
	if err := mail.BuildEmailMsg(templFileName, schItems); err != nil {
		return err
	}
	if err := mail.SendEmailViaRelay(); err != nil {
		return err
	}
	return nil
}

func sendEmailForWeb(templFileName string, simulation bool, URL string) error {
	mail := mail.MailSender{}
	mail.FillConf(simulation)
	if err := mail.BuildEmailMsgWithURL(templFileName, URL); err != nil {
		return err
	}
	if err := mail.SendEmailViaRelay(); err != nil {
		return err
	}
	return nil
}

func sendTelegram(templFileName string, simulation bool, schItems []*idl.SchedNextItem, debug bool) error {
	ts := telegram.TelegramSender{}
	ts.FillConf(simulation, debug)

	if err := ts.BuildMsg(templFileName, schItems); err != nil {
		return err
	}
	if err := ts.Send(); err != nil {
		return err
	}
	return nil
}

func sendTelegramForWeb(templFileName string, simulation bool, URL string, debug bool) error {
	ts := telegram.TelegramSender{}
	ts.FillConf(simulation, debug)

	if err := ts.BuildMsgWithURL(templFileName, URL); err != nil {
		return err
	}
	if err := ts.Send(); err != nil {
		return err
	}
	return nil
}
