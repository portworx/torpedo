package task

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"time"
)

// ErrTimedOut is returned when an operation times out
var ErrTimedOut = errors.New("timed out performing task")

// DoRetryWithTimeout performs given task with given timeout and timeBeforeRetry
func DoRetryWithTimeout(t func() error, timeout, timeBeforeRetry time.Duration) error {
	done := make(chan bool, 1)
	quit := make(chan bool, 1)

	go func() {
		for {
			select {
			case q := <-quit:
				if q {
					logrus.Infof("Timed out, quitting")
					return
				}

			default:
				err := t()
				if err == nil {
					logrus.Infof("Task done: %v", t)
					done <- true
					return
				}
				logrus.Infof("Will retry task")
				time.Sleep(timeBeforeRetry)
			}
		}
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		quit <- true
		return ErrTimedOut
	}
}
