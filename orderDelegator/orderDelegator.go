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

/*func cost(order fsm.Order, state fsm.State) {
	return order.Location
}
*/
func OrderDelegator(order_chan chan fsm.Order, state_chan chan fsm.State, numFloors int, numElev int) {
	//go testOrder(order_chan)
	//go testState(state_chan)
	/*states := make([]fsm.State, numElev)
	for i := 0; i < numElev; i++ {
		orders := make([]bool, numFloors*3) //inits as false :D
		var state fsm.State
		state.Exe_orders = orders
		state.Floor = 0
		state.Dir = 0

	}
	*/

	var states map[int]fsm.State

	for {
		select {
		case a := <-state_chan:
			fmt.Printf("\nIn floor %d\n", a.Floor)
			states[a.Id] = a

		case a := <-order_chan:
			fmt.Printf("Order in floor %d", a.Location.Floor)
		}
	}
}
