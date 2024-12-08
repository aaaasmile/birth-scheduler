package depl

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
)

func CreateDeployZip(rootDirRel string, pathItems []string, outFn string, fnPathRes func(string) string) {
	// root path
	rootPath := path.Join(path.Dir("."), rootDirRel)
	rootPath, _ = filepath.Abs(rootPath)
	//fmt.Println("Root path is: ", rootPath)

	onlyFiles := getAllFiles(pathItems, rootPath)
	//fmt.Println(onlyFiles)

	// Add files to zip
	if err := zipFiles(rootPath, onlyFiles, outFn, fnPathRes); err != nil {
		log.Fatal(err)
	}
	log.Println("Zipped file: ", outFn)
}

func getAllFiles(pathItems []string, rootPath string) []string {
	onlyFiles := []string{}
	for _, pathItem := range pathItems {
		log.Printf("Process item %s", pathItem)
		itemAbs := path.Join(rootPath, pathItem)
		if info, err := os.Stat(itemAbs); err == nil && info.IsDir() {
			arr := []string{}
			arr = getFilesinDir(itemAbs, pathItem, arr)
			//fmt.Println("Dir process result: ", arr)
			for _, ele := range arr {
				onlyFiles = append(onlyFiles, ele)
			}
		} else {
			onlyFiles = append(onlyFiles, pathItem)
		}
	}
	return onlyFiles
}

func getFilesinDir(dirAbs string, dirRel string, ini []string) []string {
	r := ini
	log.Println("Scan dir ", dirAbs)
	files, err := os.ReadDir(dirAbs)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		itemAbs := path.Join(dirAbs, f.Name())
		if info, err := os.Stat(itemAbs); err == nil && info.IsDir() {
			//fmt.Println("** Sub dir found ", f.Name())
			r = getFilesinDir(itemAbs, path.Join(dirRel, f.Name()), r)
		} else {
			//fmt.Println("** file is ", f.Name())
			r = append(r, path.Join(dirRel, f.Name()))
		}
	}
	return r
}

func zipFiles(rootDirAbs string, relFiles []string, outFilename string, fnPathRes func(string) string) error {

	newZipFile, err := os.Create(outFilename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, pathItem := range relFiles {
		itemAbs := path.Join(rootDirAbs, pathItem)

		zipfile, err := os.Open(itemAbs)
		if err != nil {
			return err
		}
		defer zipfile.Close()

		// Get the file information
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Using FileInfoHeader() above only uses the basename of the file. If we want
		// to preserve the folder structure we can overwrite this with the full path.
		header.Name = fnPathRes(pathItem) //pathItem

		// Change to deflate to gain better compression
		// see http://golang.org/pkg/archive/zip/#pkg-constants
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err = io.Copy(writer, zipfile); err != nil {
			return err
		}
	}
	return nil
}
