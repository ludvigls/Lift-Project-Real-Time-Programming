package main

import (
	"fmt"

	"./fsm"
	"./io"
	"./orderDelegator"
)

func testOrder(order_chan chan int) {
	for {
		select {
		case a := <-order_chan:
			fmt.Printf("\nWE GOT ORDER %d on CHAN\n", a)
		}
	}
}
func testState(state_chan chan fsm.State) {
	for {
		select {
		case <-state_chan:
			fmt.Printf("\nWE GOT STATEEEEEEEEEEEEEEEEEEEEEEEEEE %d on CHAN\n")
		}
	}
}
func main() {

	numFloors := 4
	io.Init("localhost:15657", numFloors)

	drv_buttons := make(chan io.ButtonEvent)
	drv_floors := make(chan int)
	order_chan := make(chan int)
	state_chan := make(chan fsm.State)

	go io.Io(drv_buttons, drv_floors)

	go fsm.Fsm(drv_buttons, drv_floors, numFloors, order_chan, state_chan)
	go orderDelegator.OrderDelegator(order_chan, state_chan, numFloors)
	for {
	}

}
