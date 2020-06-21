package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

func findFilesWithExt(dirPath string, exts ...string) <-chan string {
	extMap := map[string]bool{}
	for _, e := range exts {
		extMap[e] = true
	}

	out := make(chan string)
	go func() {
		err := filepath.Walk(dirPath,
			func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if path.Ext(p) != "" && extMap[path.Ext(p)[1:]] {
					out <- p
				}
				return nil
			})
		if err != nil {
			log.Fatal(err)
		}
		close(out)
	}()
	return out
}

func appendStringToFile(p string) (chan<- string, chan<- error) {
	f, err := os.OpenFile(p,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	contentsCH := make(chan string)
	errorCh := make(chan error)
	go func() {
		for {
			select {
			case c := <-contentsCH:
				if _, err := f.WriteString(c); err != nil {
					log.Fatal(err)
				}
			case err := <-errorCh:
				log.Println(err.Error())
				close(contentsCH)
				close(errorCh)
				err = f.Close()
				log.Fatal(err)
			}
		}
	}()

	return contentsCH, errorCh
}

// "/home/me/Templates/Note/app"
func main() {
	contentsCH, errorCh := appendStringToFile("./out.txt")

	for filePath := range findFilesWithExt("/home/me/Templates/Note/app", "java", "xml") {
		log.Println(filePath)
		dat, err := ioutil.ReadFile(filePath)
		if err != nil {
			panic(err)
		}
		contentsCH <- string(dat)
	}
	errorCh <- errors.New("ENDing")
	fmt.Println("nice!")
}
