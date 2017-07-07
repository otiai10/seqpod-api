package models

import "os"

var dbname string

func init() {
	if os.Getenv("GO_ENV") == "production" {
		dbname = "fastpot_production"
	} else {
		dbname = "fastpot_dev"
	}
}
