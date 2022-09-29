package main

import (
	"log"
	"time"
)

var (
	numberOfFreeWRSeats = 3 // total number of seats in the waiting room

	// based on the task description i was asked to limiting the use of Shared
	// Memory as much as possible, so i decided to use Go channels for managing
	// inter-process communication between Barber() and Customer() process.
	//
	// using channels eliminates the need for shared pointer referencing between
	// processes.
	custReadyChannel = make(chan struct{}) // customer ready communication channel
	accessWRSeatsChannel = make(chan struct{}) // waiting room seats communication channel
	barberReadyChannel = make(chan struct{}) // barber ready communication channel
)

func wait(ch chan struct{}) {
	for range ch {
		return
	}
}

func signal(ch chan struct{}) {
	ch <- struct{}{}
}

func Barber() {
	for { // Run in an infinite loop.
		// checking if Customer is in waiting room, if no Customer
		// is in waiting room we sleep until a Customer is available.
		wait(custReadyChannel)
		// waiting for accessWRSeatsChannel before accessing or
		// modifying numberOfFreeWRSeats varaible, not doing this
		// will result in a race condition where Barber and Customer
		// is trying to access the numberOfFreeWRSeats variable at the
		// same time.
		wait(accessWRSeatsChannel)
		time.Sleep(1 * time.Second)
		log.Printf("cutting Customer's hair \n")
		numberOfFreeWRSeats++ // increasing the number of waiting room seats to accomodate new customers
		signal(barberReadyChannel) // signal the Customer that Barber done with Customer
		signal(accessWRSeatsChannel) // free lock on numberOfFreeWRSeats variable
	}
}

func Customer() {
	for { // Run in an infinite loop to simulate multiple customers
		wait(accessWRSeatsChannel) // Try to get access to the waiting room seats.
		if numberOfFreeWRSeats > 0 { // If there are any free seats:
			numberOfFreeWRSeats -= 1	//  	sit down in a chair
			signal(custReadyChannel) //     notify the barber, who's waiting until there is a customer
			signal(accessWRSeatsChannel) // don't need to lock the chairs anymore
			wait(barberReadyChannel) //     wait until the barber is ready
			log.Println("Customer hair cut completed")
			log.Printf("-------------------------------------------------------------------------- \n\n")
		} else { // otherwise, there are no free seats
			signal(accessWRSeatsChannel) // but don't forget to release the lock on the seats!
		}
	}
}

func main() {
	go Barber()
	go Customer()

	time.Sleep(10 * time.Millisecond)
	// this sends a signal to customer that numberOfFreeWRSeats varaible can
	// be accessed without doing this the Customer will keep on waiting forever,
	// same with the Barber since the Barber depends on the Customer to start running.
	signal(accessWRSeatsChannel)

	for {} // running an infinite loop so we can see the logs in std output
}