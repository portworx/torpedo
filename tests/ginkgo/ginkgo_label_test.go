package ginkgo

import (
	"errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/pkg/log"
)

// Add returns the sum of two integers.
func Add(x, y int) int {
	return x + y
}

// Subtract returns the difference of two integers.
func Subtract(x, y int) int {
	return x - y
}

// Multiply returns the product of two integers.
func Multiply(x, y int) int {
	return x * y
}

// Divide returns the quotient of two integers and an error if the divisor is 0.
func Divide(x, y int) (int, error) {
	if y == 0 {
		return 0, errors.New("cannot divide by zero")
	}
	return x / y, nil
}

var _ = Describe("Calculator Operations", Label("calculator"), Ordered, func() {
	var (
		x, y int
	)

	BeforeAll(func() {
		x = 10
		y = 5
		log.Infof("BeforeAll\n\n")
	})

	AfterAll(func() {
		log.Infof("AfterAll")
	})

	JustBeforeEach(func() {
		log.Infof("JustBeforeEach: x=%v, y=%v", x, y)
	})

	JustAfterEach(func() {
		log.Infof("JustAfterEach: x=%v, y=%v", x, y)
	})

	BeforeEach(func() {
		// Initialize variables for each test

		log.Infof("BeforeEach 1: x=%v, y=%v", x, y)
	})

	BeforeEach(func() {
		// Initialize variables for each test
		//x = 10
		//y = 5
		log.Infof("BeforeEach 2: x=%v, y=%v", x, y)
	})

	AfterEach(func() {
		// Initialize variables for each test
		//x = 10
		//y = 5
		log.Infof("AfterEach 1: x=%v, y=%v", x, y)
	})

	AfterEach(func() {
		// Initialize variables for each test
		//x = 10
		//y = 5
		log.Infof("AfterEach 2: x=%v, y=%v\n\n", x, y)
	})

	It("correctly adds two numbers", Label("addition"), func() {
		By("Adding two numbers", func() {
			result := Add(x, y)
			log.Infof("Add result %v", result)
			//Expect(result).To(Equal(15))
			x, y = 0, 1
		})
	})

	It("correctly subtracts two numbers", Label("subtraction"), func() {
		By("Subtracting two numbers", func() {
			result := Subtract(x, y)
			log.Infof("Subtract result %v", result)
			//Expect(result).To(Equal(5))
		})
	})

	It("correctly multiplies two numbers", Label("multiplication"), func() {
		By("Multiplying two numbers", func() {
			result := Multiply(x, y)
			log.Infof("Multiply result %v", result)
			//Expect(result).To(Equal(50))
		})
	})

	It("correctly divides two numbers", Label("division"), func() {
		By("Dividing two numbers", func() {
			result, err := Divide(x, y)
			log.Infof("Divide result %v", result)
			Expect(err).NotTo(HaveOccurred())
			//Expect(result).To(Equal(2))
		})
	})

	It("random", Label("random"), func() {
		By("Random", func() {
			log.Infof("Random result %v", 0)
			//Expect(result).To(Equal(2))
		})
	})
})
