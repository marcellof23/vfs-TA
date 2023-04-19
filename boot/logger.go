package boot

import (
	"log"
	"os"
)

func InitLogger() *log.Logger {
	errLogFile, err := os.OpenFile("error-log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer errLogFile.Close()

	logger := log.New(errLogFile, "error: ", 0)
	return logger
}
