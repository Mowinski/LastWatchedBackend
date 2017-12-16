package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/naoina/toml"
)

type config struct {
	LogFileName string
	Address     string
	Port        int
}

var logger *log.Logger

func main() {
	configFile := getConfigFileName()
	cfg, err := prepareConfig(configFile)
	if err != nil {
		log.Fatal("Can not read config file, check `", configFile, "` or set LASTWATCHEDMOVIE_CONFIG environment, error: ", err)
	}

	setLogger(cfg.LogFileName)

	addr := cfg.Address + ":" + strconv.Itoa(cfg.Port)
	logger.Print("Server start on: ", addr)
	router := newRouter()

	logger.Fatal(http.ListenAndServe(addr, router))
}

func prepareConfig(filename string) (cfg config, err error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return cfg, err
	}

	err = toml.NewDecoder(file).Decode(&cfg)

	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func getConfigFileName() string {
	const envName = "LASTWATCHEDMOVIE_CONFIG"
	configFile, hasEnv := os.LookupEnv(envName)

	if !hasEnv {
		configFile = "config.toml"
	}

	return configFile
}

func setLogger(logFileName string) {
	outputFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Can not open log file '", logFileName, "', error: ", err)
	}
	logger = log.New(outputFile, "", log.Lshortfile)
	logger.SetOutput(io.MultiWriter(os.Stdout, outputFile))
}
