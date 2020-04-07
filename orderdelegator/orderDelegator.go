package orderdelegator

import (
	"fmt"
	"math"
	"strconv"

	"../fsm"
	"../io"
)

func cost(order fsm.Order, state fsm.State, numFloors int) int {

	if state.ExeOrders[order.Location.Floor*3+int(io.BT_Cab)] || state.ExeOrders[order.Location.Floor*3+int(io.BT_HallUp)] || state.ExeOrders[order.Location.Floor*3+int(io.BT_HallDown)] {
		fmt.Println("Already have an order on the floor ")
		return 0
	}
	numOrders := 0
	for i := 0; i < numFloors*3; i++ {
		if state.ExeOrders[i] {
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
	dirCost := 1
	if dir == state.Dir {
		dirCost = 0
	}

	return numOrders + distCost + dirCost
}

//OrderDelegator is the 'main' function of the orderDelegator module
func OrderDelegator(n_od_orderCh chan fsm.Order, od_n_orderCh chan fsm.Order, n_od_globstateCh chan map[string]fsm.State, numFloors int) {
	states := make(map[string]fsm.State)

	for {
		select {
		case a := <-n_od_globstateCh:
			states = a

		case a := <-n_od_orderCh: // Only master recieve things from here
			//fmt.Printf("Order in floor %d", a.Location.Floor) /
			if a.Location.Button == io.BT_Cab { //cab orders should always be taken at the
				fmt.Println("CAB ORDER TAKEN BY MYSELF:", a.ID)
				od_n_orderCh <- a
			} else {
				costs := make(map[string]int)
				for k, v := range states {
					intK, _ := strconv.Atoi(k)
					if intK > 0 {
						costs[k] = cost(a, v, numFloors)
					}
				}
				minID := -1
				minCost := 1000
				for id, cost := range costs {
					if cost < minCost {
						minID, _ = strconv.Atoi(id)
						minCost = cost
					}
				}

				if minID == -1 {
					fmt.Println("No network connection, will give order to myself")
				} else {
					a.ID = minID //send order to elev with smallest cost
				}

				fmt.Println("GAVE ORDER TO ID:", a.ID)
				od_n_orderCh <- a
				//fmt.Printf("\nGive order to id: %d \n", minID)
			}
		}
	}
}
