package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBhost     string
	DBname     string
	DBport     string
	DBuser     string
	DBpassword string
}

func LoadConfigurations() (Config, error) {

	if err := godotenv.Load(".env"); err != nil {
		return Config{}, err
	}

	var conf Config

	conf.DBhost = os.Getenv("dbhost")
	conf.DBport = os.Getenv("dbport")
	conf.DBname = os.Getenv("dbname")
	conf.DBpassword = os.Getenv("dbpassword")
	conf.DBuser = os.Getenv("dbuser")

	return conf, nil
}
