package main

import (
	"fmt"

	. "github.com/portworx/torpedo/tests/backup/longevity/pxbackupevents"
)

func main() {
	OneSuccessOneFail()
	fmt.Printf("\n\n\n---------------------\n\n\n")
	OneSuccessTwoFail()
}
