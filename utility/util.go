package utility

import (
	"fmt"
	"log"
)

func HandleError(err error, msg string) {
	if err != nil {
		log.Fatal(fmt.Sprintf("%s: %s", msg, err.Error()))
	}
}
