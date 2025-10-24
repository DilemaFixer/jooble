package config

import "os"

func IsDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}
