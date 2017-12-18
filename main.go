package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"github.com/Mowinski/LastWatchedBackend/database"
	"github.com/Mowinski/LastWatchedBackend/logger"
	"github.com/naoina/toml"
)

type databaseCfg struct {
	Host     string
	Port     int
	User     string
	DBName   string
	Password string
}

type config struct {
	LogFileName string
	Address     string
	Port        int
	Database    databaseCfg
}

func main() {
	configFile := getConfigFileName()
	cfg, err := prepareConfig(configFile)
	if err != nil {
		log.Fatal("Can not read config file, check `", configFile, "` or set LASTWATCHEDMOVIE_CONFIG environment, error: ", err)
	}

	err = logger.SetLogger(cfg.LogFileName)
	if err != nil {
		log.Fatal("Can not open log file '", cfg.LogFileName, "', error: ", err)
	}

	dns := getDNS(cfg.Database)
	err = database.ConnectWithDatabase(dns)
	if err != nil {
		logger.Logger.Fatal("Can not connect to database, error:", err)
	}

	addr := cfg.Address + ":" + strconv.Itoa(cfg.Port)
	logger.Logger.Print("Server start on: ", addr)
	router := newRouter()

	logger.Logger.Fatal(http.ListenAndServe(addr, router))
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

func getDNS(databaseCfg databaseCfg) string {
	return databaseCfg.User + ":" + databaseCfg.Password + "@tcp(" +
		databaseCfg.Host + ":" + strconv.Itoa(databaseCfg.Port) + ")/" + databaseCfg.DBName + "?parseTime=true"
}
