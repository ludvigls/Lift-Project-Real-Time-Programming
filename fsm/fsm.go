package fsm

import (
	"fmt"
	"time"

	"../io"
)

type Order struct {
	Location io.ButtonEvent
	Id       int
}

type State struct {
	Exe_orders []bool
	Floor      int
	Dir        int
	Id         int
}

//const numFloors = 4

type state int

const (
	door_open int = 0
	running   int = 1
	idle      int = 2
)

func sendState(localstate_chan chan State, floor int, dir int, orders []bool, id int) {
	state := State{orders, floor, dir, id}
	//fmt.Println("ALMOST SENT STATE")
	localstate_chan <- state
	//fmt.Println("SENT STATE")
	return
}
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
func orderInFloor(curr_floor int, orders []bool) bool {
	for b := 0; b <= 2; b++ {
		if orders[curr_floor*3+b] {
			//removeOrdersInFloor(curr_floor, orders)
			return true

		}
	}
	return false
}
func stopForOrder(curr_floor int, curr_dir io.MotorDirection, numFloors int, orders []bool) bool {
	if curr_dir == io.MD_Up {
		if orders[curr_floor*3+int(io.BT_HallUp)] || orders[curr_floor*3+int(io.BT_Cab)] { // take orders in curr floor if order goes up / cab order
			return true
		}
		if orders[curr_floor*3+int(io.BT_HallDown)] { // if order in curr floor wants to go down
			for f := curr_floor + 1; f < numFloors; f++ { //DONT take the order, if there are other orders in up dir above you / cab orders above
				if orders[f*3+int(io.BT_HallUp)] || orders[f*3+int(io.BT_Cab)] || orders[f*3+int(io.BT_HallDown)] {
					return false
				}
			}
			return true
		}
		//Equivalent logic but for down direction
	} else if curr_dir == io.MD_Down {
		if orders[curr_floor*3+int(io.BT_HallDown)] || orders[curr_floor*3+int(io.BT_Cab)] {
			return true
		}
		if orders[curr_floor*3+int(io.BT_HallUp)] {
			for f := 0; f < curr_floor; f++ {
				fmt.Printf("floor= %d \n", f)
				if orders[f*3+int(io.BT_HallDown)] || orders[f*3+int(io.BT_Cab)] || orders[f*3+int(io.BT_HallUp)] { //OR CAB

					return false
				}
			}

			return true
		}
	}
	//removeOrdersInFloor(curr_floor)
	return false //maybe wrong
}
func takeAnyOrder(curr_floor int, numFloors int, orders []bool) io.MotorDirection {
	for f := 0; f < numFloors; f++ {
		for b := 0; b <= 2; b++ {
			if orders[f*3+b] {
				if f > curr_floor {
					return io.MD_Up
				} else if f < curr_floor {
					return io.MD_Down
				} else {
					return io.MD_Stop
				}
			}
		}
	}
	return io.MD_Stop
}

func whereToGo(curr_floor int, curr_dir io.MotorDirection, numFloors int, orders []bool) io.MotorDirection {
	if orderInFloor(curr_floor, orders) { // take orders in curr floor if order goes up / cab order
		return io.MD_Stop
	}
	if curr_dir == io.MD_Up {

		// if order in curr floor wants to go down
		for f := curr_floor + 1; f < numFloors; f++ { //DONT take the order, if there are other orders in up dir above you / cab orders above
			if orders[f*3+int(io.BT_HallUp)] || orders[f*3+int(io.BT_Cab)] || orders[f*3+int(io.BT_HallDown)] {
				return io.MD_Up
			}
		}
		//Equivalent logic but for down direction
	} else if curr_dir == io.MD_Down {

		for f := 0; f < curr_floor; f++ {
			if orders[f*3+int(io.BT_HallUp)] || orders[f*3+int(io.BT_Cab)] || orders[f*3+int(io.BT_HallDown)] {
				return io.MD_Down
			}
		}
	}
	return takeAnyOrder(curr_floor, numFloors, orders)
}

func Fsm(drv_buttons chan io.ButtonEvent, drv_floors chan int, numFloors int, fsm_n_order_chan chan Order, n_fsm_order_chan chan Order, localstate_chan chan State, id int) {
	Door_timer := time.NewTimer(120 * time.Second) //init door timer
	//var orders [numFloors * 3]bool                 // [. . .   . . .   . . .   . . . ] (3 x 1.etj, 3 x 2.etj ....)
	orders := make([]bool, numFloors*3)
	//INIT PHASE
	fmt.Println("STARTED FSM")

	var d io.MotorDirection = io.MD_Up
	curr_dir := io.MD_Up
	io.SetMotorDirection(d)
	curr_floor := <-drv_floors //wait until reaches floor
	io.SetFloorIndicator(curr_floor)
	d = io.MD_Stop
	io.SetMotorDirection(d)
	var curr_state state
	curr_state = 2 //idle
	sendState(localstate_chan, curr_floor, int(curr_dir), orders, id)

	for {
		//fmt.Println("New tier while loop")
		select {
		case <-Door_timer.C: // door is closing
			io.SetDoorOpenLamp(false)
			if orderInFloor(curr_floor, orders) {
				removeOrdersInFloor(curr_floor, orders)
				Door_timer = time.NewTimer(3 * time.Second)
				io.SetDoorOpenLamp(true)
				curr_state = 0 // go to door open state
			} else if hasOrder(orders) {
				curr_state = 1 //running
				fmt.Println("My direction was")
				fmt.Println(curr_dir)
				d = whereToGo(curr_floor, curr_dir, numFloors, orders)
				fmt.Println("I want to go")
				fmt.Println(d)
				curr_dir = d
				io.SetMotorDirection(d)
			} else {
				curr_state = 2
			} //idle

		case a := <-drv_buttons:
			fsm_n_order_chan <- Order{a, id}
			//io.SetButtonLamp(a.Button, a.Floor, true)
			//orders[(a.Floor)*3+int(a.Button)] = true
			//fmt.Println(orders)

		case a := <-drv_floors:
			curr_floor = a
			io.SetFloorIndicator(curr_floor)

			if stopForOrder(curr_floor, curr_dir, numFloors, orders) {
				removeOrdersInFloor(curr_floor, orders)
				d = io.MD_Stop
				io.SetMotorDirection(d)
				curr_state = 0 //door_open
			} else if a == 0 || a == numFloors-1 { // dont stop for order AND in top/bot floor
				curr_state = 2 //idle
			}
			if a == 0 || a == numFloors-1 { //change dir if you're at top / bottom floor
				// curr_state = 2 //idle
				if curr_dir == io.MD_Up {
					curr_dir = io.MD_Down
				} else {
					curr_dir = io.MD_Up
				}
			}
		case a := <-n_fsm_order_chan:
			//fmt.Println("GOT AN ASSIGNED ORDER")
			orders[a.Location.Floor*3+int(a.Location.Button)] = true
			io.SetButtonLamp(a.Location.Button, a.Location.Floor, true)
		}

		switch curr_state {
		case 0: //door open
			removeOrdersInFloor(curr_floor, orders)

			Door_timer = time.NewTimer(3 * time.Second)
			io.SetDoorOpenLamp(true)
		case 1: //running
			//fmt.Printf("running \n")
		case 2: //idle
			//fmt.Printf("idle \n")
			//check for new orders!
			//d = whereToGo(curr_dir, curr_floor)
			d = whereToGo(curr_floor, curr_dir, numFloors, orders)
			io.SetMotorDirection(d)

			if d == io.MD_Stop && orderInFloor(curr_floor, orders) {
				Door_timer = time.NewTimer(3 * time.Second)
				io.SetDoorOpenLamp(true)
				removeOrdersInFloor(curr_floor, orders)
				curr_state = 0 //door open
			} else if d != io.MD_Stop {
				curr_dir = d
				curr_state = 1 //running
			}

		}

		sendState(localstate_chan, curr_floor, int(curr_dir), orders, id)

	}
}
