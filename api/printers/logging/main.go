package logging

import "log"

func PrintPanic(message string) error {
	log.Panicln(message)
	return nil
}

func PrintFatal(message string) error {
	log.Fatalln(message)
	return nil
}

func PrintInfo(message string) error {
	log.Println(message)
	return nil
}
