package orderDelegator

import (
	"fmt"
	"math"
	"strconv"

	"../fsm"
	"../io"
)

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
	//TODO, 0 and 1 should be swapped
	dir_cost := 1
	if dir == state.Dir {
		dir_cost = 0
	}

	return num_orders + dist_cost + dir_cost
}

func OrderDelegator(n_od_order_chan chan fsm.Order, od_n_order_chan chan fsm.Order, states_chan chan map[string]fsm.State, numFloors int) {

	states := make(map[string]fsm.State) // glob state including all lifts

	for {
		select {
		case a := <-states_chan:
			states = a

		case a := <-n_od_order_chan:
			if a.Location.Button == io.BT_Cab { //cab orders should always be taken by yourself
				fmt.Println("GAVE ORDER TO ID:", a.Id)
				od_n_order_chan <- a
			} else {
				costs := make(map[string]int)
				for k, v := range states {
					costs[k] = cost(a, v, numFloors)
				}
				min_id := -1
				min_cost := 1000
				for id, cost := range costs {
					if cost < min_cost {
						min_id, _ = strconv.Atoi(id)
						min_cost = cost
					}
				}
				//send order to correct elev
				a.Id = min_id
				fmt.Println("GAVE ORDER TO ID:", a.Id)
				od_n_order_chan <- a
			}

		}
	}
}
