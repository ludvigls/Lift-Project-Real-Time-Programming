package fsm

import "../io"

import (
    "fmt"
    "time"
)

const numFloors = 4
var orders[numFloors*3] bool // [. . .   . . .   . . .   . . . ] (3 x 1.etj, 3 x 2.etj ....)

func RemoveOrdersInFloor(floor int) {
    for i:=0; i < 3; i++ { // up, down, cab
        orders[floor*3 + i] = false
        io.SetButtonLamp(io.ButtonType(i), floor, false)
    }
}
func OrderInSameDirection(curr_floor int) bool{
    if curr_dir == io.MD_Up{
        for i:=curr_floor+1;i<numFloors;i++{
            if(orders[i*3+int(io.BT_HallUp)]||orders[i+int(io.BT_Cab)]){ //OR CAB
                return true
            }
        }
    }
    else if(curr_dir==io.MD_Down){
        for i:=0;i<curr_floor-1;i++{
            if(orders[i*3+int(io.BT_HallDown)]||orders[i+int(io.BT_Cab)){ //OR CAB
                return true
            }
        }
    }
    return false
}
//func OrderInFloor(int floor) bool{  //check if order in floor
//}
func randomOrderDir(curr_floor) io.MotorDirection{
    for f:=0;f<numFloors;f++{
        for b:=0;b<=2;b++{
            if orders[3*f+b]:
                if f>curr_floor{
                    return io.MD_Up
                }
                else if f<curr_floor{
                    return io.MD_up
                }
        }
    }
    return io.MD_Stop
}
func Fsm(drv_buttons chan io.ButtonEvent, drv_floors chan int){
    Door_timer := time.NewTimer(120*time.Second) //init door timer

    //INIT PHASE
    var d io.MotorDirection = io.MD_Up
    io.SetMotorDirection(d)
    curr_floor :=<-drv_floors //wait until reaches floor
    d=io.MD_Stop
    io.SetMotorDirection(d)

    for {
        select {
            case <- Door_timer.C : // door is closing
                io.SetDoorOpenLamp(false)

                //check if order in floor before you leave
                else if (OrderInSameDirection(curr_floor)){
                    d=curr_dir
                }
                else{
                    d=randomOrderDir(curr_floor)
                }
                SetMotorDirection(curr_dir)

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
                        RemoveOrdersInFloor(a.Floor)
                    } else if (a.Floor>curr_floor) {
                        d=io.MD_Up
                        io.SetMotorDirection(d)
                    } else if (a.Floor<curr_floor) {
                        d=io.MD_Down
                        io.SetMotorDirection(d)
                    }
                }
                
            case a := <- drv_floors:
                curr_floor=a
                for i:=0; i < 3; i++ { // i : up, down, cab
                    if (orders[a*3 + i]) { // if order in floor 
                        // TODO, IF ALSO IN RIGHT DIRECTION
                        d=io.MD_Stop
                        io.SetMotorDirection(d)
                        RemoveOrdersInFloor(a)
                        
                        //Open door evt, kjøre til neste order
                        io.SetDoorOpenLamp(true)
                        Door_timer = time.NewTimer(3*time.Second) //nsek timer
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
