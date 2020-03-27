package fsm

import (
	"time"

	"../io"
)

// Order struct containing info on one order. So what type of button, where it was pressed and from which lift it is from.
type Order struct {
	Location io.ButtonEvent //TODO, change name to button
	ID       int
}

// State struct containing the state of an individual lift. What direction it has, ID, Floor and its orders
type State struct {
	ExeOrders []bool // [up down cab (1)   up down cab (2) ...]
	Floor     int
	Dir       int //
	ID        int
}

// Locally sends the state
func sendState(localstateCh chan State, floor int, dir int, orders []bool, id int) {
	state := State{orders, floor, dir, id}
	localstateCh <- state
	return
}

//Check if the lift has any orders
func hasOrder(orders []bool) bool {
	for i := 0; i < len(orders); i++ {
		if orders[i] {
			return true
		}
	}
	return false
}

func removeOrdersInFloor(floor int, orders []bool) { // Remove orders + turn off lamps
	for i := 0; i < 3; i++ { // up, down, cab
		orders[floor*3+i] = false
		io.SetButtonLamp(io.ButtonType(i), floor, false)
	}
}

func isOrderInFloor(currFloor int, orders []bool) bool {
	for b := 0; b <= 2; b++ {
		if orders[currFloor*3+b] {
			return true
		}
	}
	return false
}

func shouldStopForOrder(currFloor int, currDir io.MotorDirection, numFloors int, orders []bool) bool {
	if currDir == io.MD_Up {
		if orders[currFloor*3+int(io.BT_HallUp)] || orders[currFloor*3+int(io.BT_Cab)] { // take orders in curr floor if order goes up / cab order
			return true
		}
		if orders[currFloor*3+int(io.BT_HallDown)] { // if order in curr floor wants to go down
			for f := currFloor + 1; f < numFloors; f++ { //DONT take the order, if there are other orders in up dir above you / cab orders above
				if orders[f*3+int(io.BT_HallUp)] || orders[f*3+int(io.BT_Cab)] || orders[f*3+int(io.BT_HallDown)] {
					return false
				}
			}
			return true
		}
		//Equivalent logic but for down direction
	} else if currDir == io.MD_Down {
		if orders[currFloor*3+int(io.BT_HallDown)] || orders[currFloor*3+int(io.BT_Cab)] {
			return true
		}
		if orders[currFloor*3+int(io.BT_HallUp)] {
			for f := 0; f < currFloor; f++ {
				if orders[f*3+int(io.BT_HallDown)] || orders[f*3+int(io.BT_Cab)] || orders[f*3+int(io.BT_HallUp)] { //OR CAB

					return false
				}
			}
			return true
		}
	}
	return false
}

func selectArbitraryOrder(currFloor int, numFloors int, orders []bool) io.MotorDirection { //TODO, is this func needed??
	for f := 0; f < numFloors; f++ {
		for b := 0; b <= 2; b++ {
			if orders[f*3+b] {
				if f > currFloor {
					return io.MD_Up
				} else if f < currFloor {
					return io.MD_Down
				} else {
					return io.MD_Stop
				}
			}
		}
	}
	return io.MD_Stop
}

func whereToGo(currFloor int, currDir io.MotorDirection, numFloors int, orders []bool) io.MotorDirection {
	// Take orders in curr floor
	if isOrderInFloor(currFloor, orders) {
		return io.MD_Stop
	}
	// if lift is going up and there are orders going up.
	if currDir == io.MD_Up {
		for f := currFloor + 1; f < numFloors; f++ { //DONT take the order, if there are other orders in up dir above you / cab orders above
			if orders[f*3+int(io.BT_HallUp)] || orders[f*3+int(io.BT_Cab)] || orders[f*3+int(io.BT_HallDown)] {
				return io.MD_Up
			}
		}

		//if lift is going down, and there are orders going down.
	} else if currDir == io.MD_Down {
		for f := 0; f < currFloor; f++ {
			if orders[f*3+int(io.BT_HallUp)] || orders[f*3+int(io.BT_Cab)] || orders[f*3+int(io.BT_HallDown)] {
				return io.MD_Down
			}
		}
	}
	return selectArbitraryOrder(currFloor, numFloors, orders)
}

//Fsm is the 'main' function for the fsm module
func Fsm(drvButtons chan io.ButtonEvent, drvFloors chan int, numFloors int, fsm_n_orderCh chan Order, n_fsm_orderCh chan Order, localstateCh chan State, id int) {
	doorTimer := time.NewTimer(1200 * time.Second) //init door timer (TODO, the length of this timer is kinda jalla)
	//var orders [numFloors * 3]bool                 // [. . .   . . .   . . .   . . . ] (3 x 1.etj, 3 x 2.etj ....)
	orders := make([]bool, numFloors*3)

	//INIT PHASE
	var d io.MotorDirection = io.MD_Up
	currDir := io.MD_Up
	io.SetMotorDirection(d)
	currFloor := <-drvFloors //wait until reaches floor
	io.SetFloorIndicator(currFloor)
	d = io.MD_Stop
	io.SetMotorDirection(d)
	currState := 2 //idle
	sendState(localstateCh, currFloor, int(currDir), orders, id)

	for {
		//fmt.Println("Current state", currState)
		select {
		case <-doorTimer.C: // door is closing
			io.SetDoorOpenLamp(false)
			if isOrderInFloor(currFloor, orders) {
				removeOrdersInFloor(currFloor, orders)
				doorTimer = time.NewTimer(3 * time.Second)
				io.SetDoorOpenLamp(true)
				currState = 0 // go to door open state
			} else if hasOrder(orders) {
				currState = 1 //running
				d = whereToGo(currFloor, currDir, numFloors, orders)
				currDir = d
				io.SetMotorDirection(d)
			} else {
				currState = 2
			} //idle

		case a := <-drvButtons:
			fsm_n_orderCh <- Order{a, id}

		case a := <-drvFloors:
			currFloor = a
			io.SetFloorIndicator(currFloor)

			if shouldStopForOrder(currFloor, currDir, numFloors, orders) {
				removeOrdersInFloor(currFloor, orders)
				d = io.MD_Stop
				io.SetMotorDirection(d)
				doorTimer = time.NewTimer(3 * time.Second)
				currState = 0 //door_open
			} else if a == 0 || a == numFloors-1 { // dont stop for order AND in top/bot floor
				currState = 2 //idle
			}
			if a == 0 || a == numFloors-1 { //change dir if you're at top / bottom floor
				if currDir == io.MD_Up {
					currDir = io.MD_Down
				} else {
					currDir = io.MD_Up
				}
			}
		case a := <-n_fsm_orderCh:
			orders[a.Location.Floor*3+int(a.Location.Button)] = true
			io.SetButtonLamp(a.Location.Button, a.Location.Floor, true)
		}

		switch currState {
		case 0: //door open
			if isOrderInFloor(currFloor, orders) {
				removeOrdersInFloor(currFloor, orders)
				doorTimer = time.NewTimer(3 * time.Second)
			}
			io.SetDoorOpenLamp(true)
		case 1: //running
			//fmt.Printf("running \n")
		case 2: //idle
			//fmt.Printf("idle \n")
			d = whereToGo(currFloor, currDir, numFloors, orders)
			io.SetMotorDirection(d)

			if d == io.MD_Stop && isOrderInFloor(currFloor, orders) {
				doorTimer = time.NewTimer(3 * time.Second)
				io.SetDoorOpenLamp(true)
				removeOrdersInFloor(currFloor, orders)
				currState = 0 //door open
			} else if d != io.MD_Stop {
				currDir = d
				currState = 1 //running
			}

		}
		sendState(localstateCh, currFloor, int(currDir), orders, id)
	}
}
