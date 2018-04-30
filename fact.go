package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const MAX_RECURSION_DEPTH int = 49

/*Fact
This is an implementation of the solution for recursively
calculating a factorial presented in the whitepaper
"Communicating Sequnetial Processes" from C.A.R. Hoare 1978.
(https://spinroot.com/courses/summer/Papers/hoare_1978.pdf)

The solution in the paper introduced a concept described as
"iterative arrays" which is an array of processes whereas
processes in the array can communicate with their neighbouring
processes.
For a better explanation (from someone else) check out this
implementation of the same problem solution:
https://github.com/thomas11/csp/blob/master/csp.go#L236*/
func Fact() chan int {

	// We first create an slice with a capacity
	// of the maximum recursion depth.
	// The slice contains elements of type "chan int"
	// This is our implementation of the "array of processes"
	// whereas our processes are simply go channels which
	// allow us to pass values between co/go-routines.
	//
	// A process/gorouting that tries to read from this process/channel
	// is blocked until another process/goroutine writes to it.
	fac := make([]chan int, MAX_RECURSION_DEPTH+1, MAX_RECURSION_DEPTH+1)

	// fac[0] will be our pipe to the user of the application,
	// i.e. the process/channel the user can write to
	// to start the calculation of the factorial
	//
	// Once the factorial has been calculated, the
	// application will write the solution back to this
	// process/channel so that the user can receive/read from it.
	fac[0] = make(chan int)

	// Next we create a process/channel for each place in our
	// currently empty arrasy of processes/channels.
	// Note that we start with index 1 as index 0 is for the user channel
	for idx := 1; idx <= MAX_RECURSION_DEPTH; idx++ {
		fac[idx] = make(chan int)

		// For each process/channel we start a new co/go-routine
		go func(i int) {
			// For each coroutine we enter a loop
			// which will try to read from the previous
			// channel of the array of processes/channels.
			for {

				// We read from the previous channel/process.
				// Note that channel i-1 is not empty again.
				n := <-fac[i-1]

				// For debugging purposes we print out the
				// channel index (i) and the content
				// we read from the previous channel (n)
				fmt.Printf("i is %d n is :: %d\n", i, n)

				if n == 0 || n == 1 {
					// If the previous channel's content
					// was 1, we have reached the inflection
					// point of the recursion i.e. the max
					// depth we need to calculate the factorial.
					//
					// In this case we write a 1 back to the
					// previous channel.
					fac[i-1] <- 1

				} else if n > 1 {
					// If we are not at the inflection point yet,
					// we subtract 1 from the number we took
					// from the previous process/channel and write
					// the new number to the current channel.
					// This will now immedately enable the process
					// at fac[i+1] to read from fac[i].
					// Basically this is where we pass a value down
					// and enter the next deeper level of the recursion.
					//
					// fac[i+1] will read
					// from this channel and the current
					// coroutine will now block here.
					fac[i] <- (n - 1)

					// At some point the i+1 channel
					// will write a value back into our channel
					// (whenever fac[i+1] performs a "fac[i-1] <- ..."),
					// which means we will be unblocked as we
					// can now read from our own channel again.
					//
					// At this point we are walking back up the
					// recursion until we are back at the beginning
					// of the array of processes/channels at process
					// fac[0].
					r := <-fac[i]

					// Along the way up to fac[0] we multiply
					// the number coming from the process prior to ours (n)
					// with the number from our own channel (r) and write
					// the product to the channel prior to ours in the array,
					// which will trigger the next goroutine waiting
					// for that channel.
					fac[i-1] <- (r * n)
				}
			}
		}(idx)
	}

	return fac[0]

}

func main() {

	// We let the user enter a number of which we will
	// calculate the factorial through above recursion or
	// "iterative array".
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Calculate factorial of: ")

	userInRaw, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf(err.Error())
	}

	userIn, err := strconv.Atoi(strings.TrimSpace(userInRaw))
	if err != nil {
		log.Fatalf(err.Error())
	}

	if userIn > MAX_RECURSION_DEPTH {
		log.Fatalf("Sorry, this number is to big. You can choose number up to %d", MAX_RECURSION_DEPTH)
	}

	// By calling Fact, we initialize the co/go-routines
	// userProcess will now be the input channel/process
	// to and from which the user can write and read, respectively,
	// to trigger and communicate with the factorial method.
	userProcess := Fact()

	// We write the user input to the first process/channel
	// in the array of processes in Fact() which will trigger
	// the calculation of the factorial by unblocking the goroutine
	// reading from channel with index 0 i.e. goroutine/channel with index 1.
	//
	// We are now blocked until Fact() write back to this channel/process.
	userProcess <- userIn

	// Once Fact() has finished and written back to the "user" channel/process
	// we can read from it and output the answer
	fmt.Printf("Fin: %d! = %d", userIn, <-userProcess)
}
