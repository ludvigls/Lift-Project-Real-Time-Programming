package fsm

import (
	"fmt"
	"time"

	"../io"
)

// Order struct containing info on one order. Type of button, where it was pressed and from which lift it is from.
type Order struct {
	Location io.ButtonEvent
	ID       int
}

// State struct containing the state of an individual lift. What direction it has, ID, Floor and its orders
type State struct {
	ExeOrders []bool // [up down cab (1)   up down cab (2) ... up down cab (numFloors)]
	Floor     int
	Dir       int
	ID        int
}

// Sends the state on local go channel
func sendState(localstateCh chan State, floor int, dir int, orders []bool, id int) {
	state := State{orders, floor, dir, id}
	localstateCh <- state
	return
}

// Check if the lift has any orders
func hasOrder(orders []bool) bool {
	for i := 0; i < len(orders); i++ {
		if orders[i] {
			return true
		}
	}
	return false
}

// Remove orders and turn off lamps
func removeOrdersInFloor(floor int, orders []bool) {
	for i := 0; i < 3; i++ { // up, down, cab
		orders[floor*3+i] = false
		io.SetButtonLamp(io.ButtonType(i), floor, false)
	}
}

// Add orders and turn on lamps
func addOrder(floor int, buttonType io.ButtonType, orders []bool) {
	orders[floor*3+int(buttonType)] = true
	io.SetButtonLamp(buttonType, floor, true)
}

// Checks whether there is an order in the floor or not
func isOrderInFloor(currFloor int, orders []bool) bool {
	for b := 0; b <= 2; b++ {
		if orders[currFloor*3+b] {
			return true
		}
	}
	return false
}

// Checks whether the lift should stop for an order or not
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
	} else if currDir == io.MD_Down { //Equivalent logic but for down direction
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

//Select the first order it finds
func selectArbitraryOrder(currFloor int, numFloors int, orders []bool) io.MotorDirection {
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

// Outputs a favorable motordirection given its state
func whereToGo(currFloor int, currDir io.MotorDirection, numFloors int, orders []bool) io.MotorDirection {
	// Take orders in curr floor
	if isOrderInFloor(currFloor, orders) {
		return io.MD_Stop
	}
	// Take orders if lift is going up and there are orders going up
	if currDir == io.MD_Up {
		for f := currFloor + 1; f < numFloors; f++ { //DONT take the order, if there are other orders in up dir above you / cab orders above
			if orders[f*3+int(io.BT_HallUp)] || orders[f*3+int(io.BT_Cab)] || orders[f*3+int(io.BT_HallDown)] {
				return io.MD_Up
			}
		}

		//Take orders if lift is going down, and there are orders going down
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
	doorTimer := time.NewTimer(1200 * time.Second) //init door timer to a large time
	orders := make([]bool, numFloors*3)            // [. . .   . . .   . . .   . . . ] (3 x 1.floor, 3 x 2.floor ...)
	// [up down cab (1)   up down cab (2) ... up down cab(numFloors)]

	// Turn off all button lights
	for f := 0; f < numFloors; f++ {
		io.SetButtonLamp(io.BT_HallUp, f, false)
		io.SetButtonLamp(io.BT_HallDown, f, false)
		io.SetButtonLamp(io.BT_Cab, f, false)
	}

	//Ascends to the floor above
	var d io.MotorDirection = io.MD_Up
	currDir := io.MD_Up
	io.SetMotorDirection(d)
	currFloor := <-drvFloors //wait until lift reaches floor
	io.SetFloorIndicator(currFloor)
	d = io.MD_Stop
	io.SetMotorDirection(d)

	//Go to idle state
	currState := 2 //idle
	sendState(localstateCh, currFloor, int(currDir), orders, id)

	for {
		select {
		case <-doorTimer.C: // door is closing
			io.SetDoorOpenLamp(false)
			if isOrderInFloor(currFloor, orders) {
				removeOrdersInFloor(currFloor, orders)
				doorTimer = time.NewTimer(3 * time.Second)
				io.SetDoorOpenLamp(true)
				currState = 0 //door open
			} else if hasOrder(orders) {
				currState = 1 //running
				d = whereToGo(currFloor, currDir, numFloors, orders)
				currDir = d
				io.SetMotorDirection(d)
			} else {
				currState = 2 //idle
			}

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
			} else if a == 0 || a == numFloors-1 { // shouldnt stop for order OR in top/bot floor
				currState = 2 //idle
			}
			if a == 0 || a == numFloors-1 { //change dir if lift is at top / bottom floor
				if currDir == io.MD_Up {
					currDir = io.MD_Down
				} else {
					currDir = io.MD_Up
				}
			}
		case a := <-n_fsm_orderCh:
			addOrder(a.Location.Floor, a.Location.Button, orders)
		}

		switch currState {
		case 0: //door open
			fmt.Printf("State: door open \n")
			if isOrderInFloor(currFloor, orders) {
				removeOrdersInFloor(currFloor, orders)
				doorTimer = time.NewTimer(3 * time.Second)
			}
			io.SetDoorOpenLamp(true)
		case 1: //running
			fmt.Printf("State: running \n")
			io.SetDoorOpenLamp(false)
		case 2: //idle
			fmt.Printf("State: idle \n")
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
