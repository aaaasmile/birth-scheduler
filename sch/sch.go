package sch

import (
	"birthsch/conf"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	Appname = "birthday-sch"
	Buildnr = "00.001.20241208-00"
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

func doSchedule() error {
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
