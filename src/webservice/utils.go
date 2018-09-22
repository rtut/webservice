package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/lib/pq"
)

const (
	configPath   = "./config/conf.json"
	dbType       = "postgres"
	dbConnStr    = "host=%s port=%s user=%s password=%s dbname=%s"
	fileLocalLog = "/tmp/local.log"
	sizeLocalLog = 104857600
)

var (
	LocalLog  *log.Logger
	RemoteLog *log.Logger
)

func init() {
	var err error
	InitLocalLogger()
	LocalLog.Println("start app")
	readyConnStr := fmt.Sprintf(
		dbConnStr, ServiceConfig.DbHost, ServiceConfig.DbPort,
		ServiceConfig.DbUser, ServiceConfig.DbPass, ServiceConfig.DbName)
	DB, err = sql.Open(dbType, readyConnStr)
	if err != nil {
		LocalLog.Println(err)
	}
}

var DB *sql.DB

type serviceConfig struct {
	DbHost      string `json:"db_host"`
	DbUser      string `json:"db_user"`
	DbPass      string `json:"db_pass"`
	DbPort      string `json:"db_port"`
	DbName      string `json:"db_name"`
	ServicePort int    `json:"service_port"`
}

var ServiceConfig serviceConfig

func (sc *serviceConfig) LoadConfig() {
	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Println("Error load config json file")
		os.Exit(1)
	}
	err = json.Unmarshal(file, &sc)
	if err != nil {
		fmt.Println("Error convert json to struct")
	}
}

func InitLogger(fileLog string, size int64, isFill bool) *log.Logger {
	// set location of log file

	fileInfo, err := os.Stat(fileLog)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Failed to open log file", fileLog, ":", err)
		}
	} else {
		if fileInfo.Size() >= size {
			os.Remove(fileLog)
			os.OpenFile(fileLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		}
	}
	flag.Parse()

	file, err := os.OpenFile(fileLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file ", fileLog, " : ", err)
		os.Exit(1)
	}
	multi := io.MultiWriter(file, os.Stdout)
	if isFill {
		return log.New(multi, "", log.LstdFlags|log.Llongfile)
	} else {
		return log.New(multi, "", 0)
	}
}

func InitLocalLogger() {
	LocalLog = InitLogger(fileLocalLog, sizeLocalLog, true)
}

func CheckErr(msg string, err error) {
	if err != nil {
		LocalLog.Println(msg, err)
	}
}
