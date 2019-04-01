package valuable

import (
	// writing less code is more (readable)!
	// writing logrus everytime we want to just write log statements is not helpful either!
	log "github.com/sirupsen/logrus"
)

// Operation is example valuable operation that logs valuable information
func Operation() {
	// if you notice here we didn't inject a special logger instance
	// less code is more (readable)!
	log.Infoln("this is a valuable message from a valuable operation")

}
