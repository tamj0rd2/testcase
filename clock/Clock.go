package clock

import (
	"github.com/adamluzsi/testcase/clock/internal"
	"time"
)

func TimeNow() time.Time {
	return internal.GetTime()
}

func Sleep(d time.Duration) {
	<-After(d)
}

func After(d time.Duration) <-chan time.Time {
	startedAt := internal.GetTime()
	ch := make(chan time.Time)
	go func() {
	wait:
		for {
			select {
			case <-internal.Listen(): // FIXME: flaky behaviour with time travelling
				continue wait
			case <-time.After(internal.RemainingDuration(startedAt, d)):
				break wait
			}
		}
		ch <- TimeNow()
		close(ch)
	}()
	return ch
}