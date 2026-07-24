package config

import "os"

var ChaveSecreta = []byte(os.Getenv("JWT_SECRET"))
