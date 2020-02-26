package fsm

import "../io"

import (
    "fmt"
    "time"
)

const numFloors = 4
var orders[numFloors*3] bool // [. . .   . . .   . . .   . . . ] (3 x 1.etj, 3 x 2.etj ....)
var curr_dir io.MotorDirection

const (
        IDLE          int = 0
        running       int = 1
        serving_floor int = 2
)
var curr_floor int =1



func RemoveOrdersInFloor(floor int) {
    for i:=0; i < 3; i++ { // up, down, cab
        orders[floor*3 + i] = false
        io.SetButtonLamp(io.ButtonType(i), floor, false)
    }
}

func Fsm(drv_buttons chan io.ButtonEvent, drv_floors chan int){
    Door_timer := time.NewTimer(120*time.Second) //init door timer
    //var dir_i io.ButtonType

    //INIT PHASE
    var dir io.MotorDirection = io.MD_Up
    curr_dir = dir
    io.SetMotorDirection(dir)
    curr_floor :=<-drv_floors //wait until reaches floor
    dir=io.MD_Stop
    io.SetMotorDirection(dir)
    state := IDLE
	fmt.Println("Leviathan")
	var stopping  int = 0

    for {
        select {
            case <- Door_timer.C : // door is closing
                io.SetDoorOpenLamp(false)
                state=IDLE
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
            
            case a := <- drv_buttons:
                fmt.Printf("%+v\n", a)
                io.SetButtonLamp(a.Button, a.Floor, true)
                
                orders[(a.Floor)*3 + int(a.Button)] = true
                fmt.Println(orders)
                fmt.Println("Beelsebob")
               /* if (dir==io.MD_Stop){ // Idle state
                    if (a.Floor==curr_floor) {
                        Door_timer = time.NewTimer(3*time.Second)
                        RemoveOrdersInFloor(curr_floor)
                    } else if (a.Floor>curr_floor) {
                        dir=io.MD_Up
                        curr_dir = dir
                        io.SetMotorDirection(dir)
                    } else if (a.Floor<curr_floor) {
                        dir=io.MD_Down
                        curr_dir = dir
                        io.SetMotorDirection(dir)
                    }
                }*/
               
            case a := <- drv_floors:
                curr_floor = a
                stopping = fsm_stop_in_floor(curr_floor, int( curr_dir ))
                if (stopping == 1) {
                	Door_timer = time.NewTimer(3*time.Second)
                	state = serving_floor
                }
                fmt.Println("Lucifer")	

                /*
                //Take only orders in same direction osv
                //Remap
                if (curr_dir == io.MD_Up) {
                    dir_i = io.BT_HallUp
                } else {
                    dir_i = io.BT_HallDown
                }

                if (orders[curr_floor*3 + int(dir_i)] || orders[curr_floor*3 + int(io.BT_Cab)]) { // if (order in same dir OR cab order)
                    d=io.MD_Stop
                    io.SetMotorDirection(d)
                    RemoveOrdersInFloor(curr_floor)
                    
                    io.SetDoorOpenLamp(true)
                    Door_timer = time.NewTimer(3*time.Second)
                }
                */

                // Take all orders
                /*for i:=0; i < 3; i++ { // i : up, down, cab
                    if (orders[a*3 + i]) { // if order in floor SimElevatorServer

                        dir=io.MD_Stop
                        io.SetMotorDirection(dir)
                        RemoveOrdersInFloor(a)
                        
                        io.SetDoorOpenLamp(true)
                        Door_timer = time.NewTimer(3*time.Second)
                    }
                }*/

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
        if ( state == IDLE ) {
        	io.SetMotorDirection(io.MD_Stop)
            for c := 0; c < numFloors; c++ {
                for x:=0; x < 3; x++{
                    if orders[c*3 + x] == true {
                    	fmt.Println("go to state RUNNING")
                        state = running
                    }
                }
            }
        }
        if ( state == running ) {
        	fmt.Println("RUNNING")
        	fsm_set_dir()
            io.SetMotorDirection(curr_dir)
            
            
        }
        if ( state == serving_floor ) {
        	RemoveOrdersInFloor(curr_floor)
        	dir = io.MD_Stop
        	io.SetMotorDirection(dir)
        	//fsm_serve_floor()
        }
        
    }    
}

func fsm_set_dir() {
	fmt.Println("dir before")
	fmt.Println("%+v\n", curr_dir)
    continue_in_dir := 0
    if (curr_dir == io.MD_Up) {
        for c:= curr_floor+1; c < numFloors; c++ {
            for x:= 0; x<3; x++{
            	/*fmt.Println("c")
    			fmt.Println("%+v\n", c)
    			fmt.Println("x")
    			fmt.Println("%+v\n", x)
    			fmt.Println("%+v\n", orders[c*3 + x])*/
                if orders[c*3 + x] == true {
                    continue_in_dir = 1
                    fmt.Println("continue")
                }
            }
            
        }
    }else {
        for c:=0 ; c < curr_floor; c++{
            for x:= 0; x<3; x++{
                if orders[c*3 +x] == true {
                    continue_in_dir = 1
                    fmt.Println("continue")
                }
            }
        }
    }
    if continue_in_dir == 0 { curr_dir = (-1) * curr_dir 
    fmt.Println("change dir")}
    fmt.Println("dir after")
    fmt.Println("%+v\n", curr_dir)
}

func fsm_serve_floor() {
    for c:=0; c<3; c++ {
        orders[curr_floor*3 + c] = false
    }
}

func fsm_stop_in_floor(curr_floor int, curr_dir int) int {
	ans := 0
	if (orders[curr_floor * 3 + 2] == true) {ans = 1}
	if (curr_dir == 1) {
		if (orders[curr_floor * 3] == true) {ans = 1}
	}else if (curr_dir == -1) {
		if (orders[curr_floor * 3 + 1] == true) {ans = 1}
	}
    if (ans == 1){fmt.Println("STOPP")}
	return ans
}
	
