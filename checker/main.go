package main

import (
	"bufio"
	"flag"
	"github.com/mono83/slf/recievers/ansi"
	"github.com/mono83/slf/wd"
	"os"
	"strings"
	"time"
	"github.com/2at2/proxyChecker/checker/module"
)

func main() {
	var target string
	var filePath string
	var resultPath string
	flag.StringVar(&target, "target", "", "")
	flag.StringVar(&filePath, "file", "", "")
	flag.StringVar(&resultPath, "result", "", "")
	flag.Parse()

	started := time.Now()

	if len(target) == 0 {
		panic("Empty target url")
	}
	if len(filePath) == 0 {
		panic("Empty file path")
	}

	wd.AddReceiver(ansi.New(true, true, false))

	log := wd.New("checker", "checker.")

	log.Debug("Application starting")

	log.Info("File path :path", wd.StringParam("path", filePath))

	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Unable open file - :err", wd.ErrParam(err))
		panic(err)
	}
	defer file.Close()

	resultFile, err := os.Create(resultPath)

	if err != nil {
		log.Error("Unable create file - :err", wd.ErrParam(err))
		panic(err)
	}
	defer resultFile.Close()

	list := []string{}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lineStr := scanner.Text()

		if len(lineStr) != 0 {
			if !strings.HasPrefix(lineStr, "http://") && !strings.HasPrefix(lineStr, "https://") {
				lineStr = "http://" + lineStr
			}

			list = append(list, lineStr)
		}
	}

	log.Info("Loaded :count proxies", wd.IntParam("count", len(list)))

	// Filtering unique
	mappedList := make(map[string]string, len(list))

	for _, proxy := range list {
		mappedList[proxy] = proxy
	}

	log.Info("Filtered :count unique proxies", wd.IntParam("count", len(mappedList)))

	checkerModule, err := module.NewModule(target, mappedList, log)

	if err != nil {
		log.Error("Unable to build module - :err", wd.ErrParam(err))
		panic(err)
	}

	results, err := checkerModule.Process()

	if err != nil {
		log.Error("Failed process - :err", wd.ErrParam(err))
		panic(err)
	}

	log.Info("Filtered :count alive proxies", wd.IntParam("count", len(results)))

	for _, p := range results {
		if strings.HasPrefix(p, "http://") {
			p = strings.TrimLeft(p, "http://")
		}
		if strings.HasPrefix(p, "https://") {
			p = strings.TrimLeft(p, "https://")
		}

		_, err := resultFile.WriteString(p + "\n")

		if err != nil {
			log.Error("Unable to write - :err", wd.ErrParam(err))
			panic(err)
		}
	}
	resultFile.Sync()

	done := time.Now().Sub(started)

	log.Info("Done at :time", wd.FloatParam("time", done.Seconds()))
}
