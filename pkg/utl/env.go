package utl

import "os"

func GetenvDefault(key, defaultVal string) string {
	v, ex := os.LookupEnv(key)
	if !ex {
		return defaultVal
	}
	return v
}
