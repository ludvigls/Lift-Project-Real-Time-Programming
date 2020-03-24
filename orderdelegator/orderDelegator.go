package orderdelegator

import (
	"fmt"
	"math"

	"../fsm"
	"../io"
)

func cost(order fsm.Order, state fsm.State, numFloors int) int {
	numOrders := 0
	for i := 0; i < numFloors*3; i++ {
		if state.Exe_orders[i] {
			numOrders++
		}
	}
	dir := 0
	distCost := int(math.Abs(float64(order.Location.Floor - state.Floor)))
	if int(order.Location.Button) == 0 {
		dir = 1
	} else if int(order.Location.Button) == 1 {
		dir = -1
	}
	dirCost := 0
	if dir == state.Dir {
		dirCost = 1
	}

	return numOrders + distCost + dirCost
}

//OrderDelegator is the 'main' function of the orderDelegator module
func OrderDelegator(n_od_orderCh chan fsm.Order, od_n_orderCh chan fsm.Order, statesCh chan map[int]fsm.State, numFloors int) {
	states := make(map[int]fsm.State)

	for {
		select {
		case a := <-statesCh:
			states = a
			fmt.Println(states)

		case a := <-n_od_orderCh:
			//fmt.Printf("Order in floor %d", a.Location.Floor) /
			if a.Location.Button == io.BT_Cab { //cab orders should always be taken at the
				fmt.Println("CAB ORDER TAKEN BY MYSELF:", a.Id)
				od_n_orderCh <- a
			} else {
				costs := make(map[int]int)
				for k, v := range states {
					costs[k] = cost(a, v, numFloors)
				}
				minID := -1
				minCost := 1000
				for id, cost := range costs {
					if cost < minCost {
						minID = id
						minCost = cost
					}
				}
				//send order to correct elev
				a.Id = minID
				fmt.Println("GAVE ORDER TO ID:", a.Id)
				od_n_orderCh <- a
				//fmt.Printf("\nGive order to id: %d \n", minID)
			}
		}
	}
}
