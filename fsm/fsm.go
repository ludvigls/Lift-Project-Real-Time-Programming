package fsm

import "../io"

import (
    "fmt"
    "time"
)

const numFloors = 4
var orders[numFloors*3] bool // [. . .   . . .   . . .   . . . ] (3 x 1.etj, 3 x 2.etj ....)
var curr_dir io.MotorDirection

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
    var d io.MotorDirection = io.MD_Up
    curr_dir = d
    io.SetMotorDirection(d)
    curr_floor :=<-drv_floors //wait until reaches floor
    d=io.MD_Stop
    io.SetMotorDirection(d)

    for {
        select {
            case <- Door_timer.C : // door is closing
                io.SetDoorOpenLamp(false)
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

                if (d==io.MD_Stop){ // Idle state
                    if (a.Floor==curr_floor) {
                        Door_timer = time.NewTimer(3*time.Second)
                        RemoveOrdersInFloor(curr_floor)
                    } else if (a.Floor>curr_floor) {
                        d=io.MD_Up
                        curr_dir = d
                        io.SetMotorDirection(d)
                    } else if (a.Floor<curr_floor) {
                        d=io.MD_Down
                        curr_dir = d
                        io.SetMotorDirection(d)
                    }
                }
                
            case a := <- drv_floors:
                curr_floor = a

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
                for i:=0; i < 3; i++ { // i : up, down, cab
                    if (orders[a*3 + i]) { // if order in floor 
                        d=io.MD_Stop
                        io.SetMotorDirection(d)
                        RemoveOrdersInFloor(a)
                        
                        io.SetDoorOpenLamp(true)
                        Door_timer = time.NewTimer(3*time.Second)
                    }
                }

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
        
    }    
}
