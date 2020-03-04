package fsm

import (
	"fmt"
	"time"

	"../io"
)

const numFloors = 4

type state int

const (
	door_open int = 0
	running   int = 1
	idle      int = 2
)

var orders [numFloors * 3]bool // [. . .   . . .   . . .   . . . ] (3 x 1.etj, 3 x 2.etj ....)
func hasOrder() bool {
	for i := 0; i < len(orders); i++ {
		if orders[i] {
			return true
		}
	}
	return false
}
func removeOrdersInFloor(floor int) { // Remove orders + turn off lamps
	for i := 0; i < 3; i++ { // up, down, cab
		orders[floor*3+i] = false
		io.SetButtonLamp(io.ButtonType(i), floor, false)
	}
}
func orderInFloor(curr_floor int) bool {
	for b := 0; b <= 2; b++ {
		if orders[curr_floor*3+b] {
			removeOrdersInFloor(curr_floor)
			return true

		}
	}
	return false
}
func stopForOrder(curr_floor int, curr_dir io.MotorDirection) bool {
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
func takeAnyOrder(curr_floor int) io.MotorDirection {
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
	fmt.Printf("couldnt find any orders \n")
	return io.MD_Stop
}

func whereToGo(curr_floor int, curr_dir io.MotorDirection) io.MotorDirection {
	if orderInFloor(curr_floor) { // take orders in curr floor if order goes up / cab order
		return io.MD_Stop
	}
	if curr_dir == io.MD_Up {

		// if order in curr floor wants to go down
		for f := curr_floor + 1; f < numFloors; f++ { //DONT take the order, if there are other orders in up dir above you / cab orders above
			if orders[f*3+int(io.BT_HallUp)] || orders[f*3+int(io.BT_Cab)] {
				return io.MD_Up
			}
		}
		//Equivalent logic but for down direction
	} else if curr_dir == io.MD_Down {

		for f := 0; f < curr_floor; f++ {
			if orders[f*3+int(io.BT_HallUp)] || orders[f*3+int(io.BT_Cab)] {
				return io.MD_Down
			}
		}
	}
	return takeAnyOrder(curr_floor)
}

//}

//func OrderInFloor(int floor) bool{  //check if order in floor
//}
/*

func randomOrderDir(curr_floor int) io.MotorDirection {
	for f := 0; f < numFloors; f++ {
		for b := 0; b <= 2; b++ {
			if orders[3*f+b] {
				if f > curr_floor {
					return io.MD_Up
				} else if f < curr_floor {
					return io.MD_Up
				} else {
					return io.MD_Stop
				}
			}
		}
	}
	return io.MD_Stop
}*/

func Fsm(drv_buttons chan io.ButtonEvent, drv_floors chan int) {
	Door_timer := time.NewTimer(120 * time.Second) //init door timer

	//INIT PHASE

	var d io.MotorDirection = io.MD_Up
	curr_dir := io.MD_Up
	io.SetMotorDirection(d)
	curr_floor := <-drv_floors //wait until reaches floor
	d = io.MD_Stop
	io.SetMotorDirection(d)
	var curr_state state
	curr_state = 2 //idle

	for {
		select {
		case <-Door_timer.C: // door is closing
			fmt.Printf("door closing \n")
			io.SetDoorOpenLamp(false)
			if orderInFloor(curr_floor) {
				Door_timer = time.NewTimer(3 * time.Second)
				io.SetDoorOpenLamp(true)
				curr_state = 0 // go to door open state
			} else if hasOrder() {
				curr_state = 1 //running
				d = whereToGo(curr_floor, curr_dir)
				io.SetMotorDirection(d)
				curr_dir = d
			} else {
				curr_state = 2
			} //idle

			//check if order in floor before you leave

			/*if OrderInSameDirection(curr_floor, curr_dir) {
				d = curr_dir
			} else {
				d = randomOrderDir(curr_floor)
			}

			io.SetMotorDirection(d)
			*/

			// se om det finnes orders
			// gå i riktig retning

			/*
			   if (a.Floor>curr_floor) {
			       d=io.MD_Up
			       io.SetMotorDirection(d)
			   } else if (a.Floor<curr_floor) {
			       d=io.MD_Down
			       io.SetMotorDirection(d)
			   }
			*/

		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			io.SetButtonLamp(a.Button, a.Floor, true)
			orders[(a.Floor)*3+int(a.Button)] = true
			fmt.Println(orders)
			/*
				if d == io.MD_Stop { // Idle state
					if a.Floor == curr_floor {
						Door_timer = time.NewTimer(3 * time.Second)
						RemoveOrdersInFloor(a.Floor)
					} else if a.Floor > curr_floor {
						d = io.MD_Up
						io.SetMotorDirection(d)
					} else if a.Floor < curr_floor {
						d = io.MD_Down
						io.SetMotorDirection(d)
					}
				}
			*/

		case a := <-drv_floors:
			curr_floor = a
			if stopForOrder(curr_floor, curr_dir) {
				removeOrdersInFloor(curr_floor)
				d = io.MD_Stop
				io.SetMotorDirection(d)
				curr_state = 0 //door_open
				fmt.Printf("STOPPING \n")
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
			/*
				for i := 0; i < 3; i++ { // i : up, down, cab
					if orders[a*3+i] { // if order in floor
						// TODO, IF ALSO IN RIGHT DIRECTION
						d = io.MD_Stop
						io.SetMotorDirection(d)
						RemoveOrdersInFloor(a)

						//Open door evt, kjøre til neste order
						io.SetDoorOpenLamp(true)
						Door_timer = time.NewTimer(3 * time.Second) //nsek timer
					}
				}
			*/
			/*
			   fmt.Printf("%+v\n", a)
			   if a == numFloors-1 {
			       d = io.MD_Down
			   } else if a == 0 {
			       d = io.MD_Up
			   }
			   io.SetMotorDirection(d)
			*/
		}

		switch curr_state {
		case 0: //door open
			fmt.Printf("door open \n")
			removeOrdersInFloor(curr_floor)

			Door_timer = time.NewTimer(3 * time.Second)
			io.SetDoorOpenLamp(true)
		case 1: //running
			fmt.Printf("running \n")
		case 2: //idle
			fmt.Printf("idle \n")
			//check for new orders!
			//d = whereToGo(curr_dir, curr_floor)
			d = whereToGo(curr_floor, curr_dir)
			io.SetMotorDirection(d)
			if d == io.MD_Stop && orderInFloor(curr_floor) {
				Door_timer = time.NewTimer(3 * time.Second)
				io.SetDoorOpenLamp(true)
				removeOrdersInFloor(curr_floor)

				curr_state = 0 //door open
			} else if d != io.MD_Stop {
				fmt.Printf("Setting direction")
				curr_dir = d
				curr_state = 1 //running
			}

		}

	}
}
