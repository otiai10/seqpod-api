package models

import "os"

var dbname string

func init() {
	if os.Getenv("GO_ENV") == "production" {
		dbname = "seqpod_production"
	} else {
		dbname = "seqpod_dev"
	}
}
