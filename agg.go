package agg

func Aggregate(f func() error) func() error {
	cCalls := make(chan chan error)

	var waitCall func()
	waitCall = func() {
		cErrOut := make(chan error)
		cNumCalls := make(chan int)

		// wait for the first call
		cCalls <- cErrOut
		go func() {
			err := f()
			// request the number of queued callers and broadcast err to the queue
			for n := <-cNumCalls; n > 0; n-- {
				cErrOut <- err
			}
		}()

		// queue additional callers until a request for the number of callers is received
		for n := 1; n > 0; {
			select {
			case cCalls <- cErrOut:
				n++
			case cNumCalls <- n:
				// start listening for new callers and break loop
				go waitCall()
				n = 0
			}
		}
	}
	go waitCall()

	return func() error {
		errc := <-cCalls
		return <-errc
	}
}
