package main

import (
	"birthsch/deploy/depl"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"
)

var (
	defOutDir = "~/app/go/birthday-scheduler/zips/"
)

func main() {
	const (
		service = "service"
	)
	var outdir = flag.String("outdir", "",
		fmt.Sprintf("Output zip directory. If empty use the hardcoded one: %s\n", defOutDir))

	var target = flag.String("target", "",
		fmt.Sprintf("Target of deployment: %s", service))

	flag.Parse()

	rootDirRel := ".."
	pathItems := []string{"birthday-scheduler.bin", "templates"}
	switch *target {
	case service:
		pathItems = append(pathItems, "deploy/config_files/service_config.toml")
		pathItems[0] = "birthday-scheduler.bin"
	default:
		log.Fatalf("Deployment target %s is not recognized or not specified", *target)
	}
	log.Printf("Create the zip package for target %s", *target)

	outFn := getOutFileName(*outdir, *target)
	depl.CreateDeployZip(rootDirRel, pathItems, outFn, func(pathItem string) string {
		if strings.HasPrefix(pathItem, "deploy/config_files") {
			return "config.toml"
		}
		return pathItem
	})
}

func getOutFileName(outdir string, tgt string) string {
	if outdir == "" {
		outdir = defOutDir
	}
	vn := depl.GetVersionNrFromFile("../idl/idl.go", "")
	log.Println("Version is ", vn)

	currentTime := time.Now()
	s := fmt.Sprintf("birthday-scheduler_%s_%s_%s.zip", strings.Replace(vn, ".", "-", -1), currentTime.Format("02012006-150405"), tgt) // current date-time stamp using 2006 date time format template
	s = filepath.Join(outdir, s)
	return s
}
