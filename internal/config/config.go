package config

import "time"

const (
	AddressPort = 8080
	ApiTimeout  = time.Second * 30
	DbFileName  = "app.db"
	DbDriver    = "sqlite"
)
