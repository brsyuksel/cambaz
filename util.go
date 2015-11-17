package cambaz

import (
	"log"
)

func printlog(verbose bool, v ...interface{}) {
	if !verbose {
		return
	}
	log.Println(v...)
}
