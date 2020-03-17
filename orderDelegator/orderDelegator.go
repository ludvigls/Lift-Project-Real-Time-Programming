package orderDelegator

import (
	"fmt"
	"math"

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

func cost(order fsm.Order, state fsm.State, numFloors int) int {
	num_orders := 0
	for i := 0; i < numFloors*3; i++ {
		if state.Exe_orders[i] {
			num_orders += 1
		}
	}
	dir := 0
	dist_cost := int(math.Abs(float64(order.Location.Floor - state.Floor)))
	if int(order.Location.Button) == 0 {
		dir = 1
	} else if int(order.Location.Button) == 1 {
		dir = -1
	}
	dir_cost := 0
	if dir == state.Dir {
		dir_cost = 1
	}

	return num_orders + dist_cost + dir_cost
}

func OrderDelegator(order_chan chan fsm.Order, states_chan chan map[int]fsm.State, numFloors int) {
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

	states := make(map[int]fsm.State)

	//orders := make([]bool, numFloors*3) //inits as false :D
	//orders[4] = true
	//orders[5] = true
	/*
		var state fsm.State
		state.Exe_orders = orders //only for testing
		state.Floor = 0
		state.Dir = 0
		state.Id = 2
		states[state.Id] = state
	*/
	for {
		select {
		case a := <-states_chan:
			//fmt.Printf("\nIn floor %d\n", a.Floor)
			states = a
			fmt.Println("Order del got states")
			fmt.Println(states)

			/*case a := <-order_chan:
			//fmt.Printf("Order in floor %d", a.Location.Floor) /
			if a.Location.Button == io.BT_Cab { //cab orders should always be taken at the
				fmt.Printf("\nGive order to id: %d \n", a.Id)
			} else {
				costs := make(map[int]int)
				for k, v := range states {
					costs[k] = cost(a, v, numFloors)
				}
				min_id := -1
				min_cost := 1000
				for id, cost := range costs {
					if cost < min_cost {
						min_id = id
						min_cost = cost
					}
				}
				//send order to correct elev
				fmt.Printf("\nGive order to id: %d \n", min_id)
			}
			*/
		}
	}
}
