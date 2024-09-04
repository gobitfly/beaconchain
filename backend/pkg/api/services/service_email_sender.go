package services

import (
	"sync"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	"github.com/gobitfly/beaconchain/pkg/commons/utils"
)

// structs
type QueuedEmail struct { // or MessageQueue, if single sender service
	Email   EMail
	Timeout uint
}
type EMail struct {
	Recipient   string
	Subject     string
	Message     string
	Attachments []byte
}

// data
var Queue []QueuedEmail

// map by recipient, last send times, etc.
// mutex

// TODO use single combined sender service? (email, webhooks, ...)
// collects & queues mails, sends in batches regularly (possibly aggregating multiple messasages to the same user to avoid spam?)
// TODO ratelimiting
// TODO send via SMTP/mailgun/others?
func (s *Services) startEmailSenderService(wg *sync.WaitGroup) {
	o := sync.Once{}
	for {
		startTime := time.Now()
		// lock mutex
		for _, item := range Queue {
			if item.Timeout > 0 {
				// drop item
				continue
			}
			/*err := s.SendEmail(item.Email)
			if err == nil {
				// drop item
			}*/
		}
		log.Infof("=== message sending done in %s", time.Since(startTime))
		o.Do(func() {
			wg.Done()
		})
		utils.ConstantTimeDelay(startTime, 30*time.Second)
	}
}

// collect, try to send regularly, drop after timeout
func (s *Services) QueueEmail(message EMail, timeout uint) {
	// queue handling, eventually
	// lock mutex, add to queue
}

// send, no queueing (fire-and-forget)
func (s *Services) SendEmail(message EMail) error {
	// try sending via mailgun, SMTP, ...
	return nil
}
