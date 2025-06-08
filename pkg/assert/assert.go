package assert

import "log"

func runAssert(msg string) {
	log.Fatal(msg)
}

func Assert(truth bool, msg string) {
	if !truth {
		runAssert(msg)
	}
}

func NoError(err error, msg string) {
	if err != nil {
		runAssert(msg)
	}
}
