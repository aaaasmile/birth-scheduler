package sch

import (
	"birthsch/conf"
	"birthsch/idl"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

func RunService(configfile string) error {

	if _, err := conf.ReadConfig(configfile); err != nil {
		return err
	}
	log.Println("Configuration is read")

	chShutdown := make(chan struct{}, 1)
	go func(chs chan struct{}) {
		if err := doSchedule(); err != nil {
			log.Println("Server is not scheduling anymore: ", err)
			chs <- struct{}{}
		}
	}(chShutdown)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	log.Println("Enter in server loop")

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
	next idl.SchedItem
}

func doSchedule() error {
	sch := Scheduler{}
	if err := sch.readDataJsonFile(conf.Current.DataFileName); err != nil {
		return err
	}

	log.Println("Enter into an infinite loop")
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

	return nil
}
