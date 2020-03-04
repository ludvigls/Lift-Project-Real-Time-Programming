package orderDelegator

import (
	"fmt"

	"../fsm"
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
			fmt.Printf("\nWE GOT STATEEEEEEEEEEEEEEEEEEEEEEEEEE  on CHAN\n")
		}
	}
}

func OrderDelegator(order_chan chan int, state_chan chan fsm.State, numFloors int) {
	go testOrder(order_chan)
	go testState(state_chan)
	orders := make([]bool, numFloors*3) //inits as false :D

	state := fsm.State{fsm.exe_orders: orders, fsm.floor: 0, fsm.dir: 0}
	for {
		select {
		case a := <-state_chan:

		}
	}
}
