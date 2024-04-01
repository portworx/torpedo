package tests

import (
	"fmt"
	"github.com/portworx/torpedo/pkg/log"
)

const MultiplyNumDuringSummation = "test-dummy-resiliency"

func PerformMultiplication() error {
	log.InfoD("Error Injection begins here .. ")
	mul := 1
	for i := 1; i <= 10; i++ {
		mul *= i
		log.InfoD("Multiplication is [%v]", mul)
	}
	fmt.Println("Multiplication of the first 10 integers:", mul)
	log.InfoD("Error Injection is completed")
	return nil
}
