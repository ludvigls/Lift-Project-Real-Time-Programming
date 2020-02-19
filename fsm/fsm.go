package fsm

import "../io"
import "fmt"

func Fsm(drv_buttons chan io.ButtonEvent, drv_floors chan int){
    numFloors := 4
    //var orders[numFloors*3] bool// [opp, X, inni, opp, ned, inni...., X, ned, inni]

    var d io.MotorDirection = io.MD_Up
    
    for {
        select {
            case a := <- drv_buttons:
                fmt.Printf("%+v\n", a)
                io.SetButtonLamp(a.Button, a.Floor, true)
                //legg til knappetrykk i order list

                //if (standing still):
                    // take the first order in the list
              
            case a := <- drv_floors:
                // if (order in floor):
                    // stop and remove order
                fmt.Printf("%+v\n", a)
                if a == numFloors-1 {
                    d = io.MD_Down
                } else if a == 0 {
                    d = io.MD_Up
                }
                io.SetMotorDirection(d)
        /* 
           
        case a := <- drv_obstr:
            fmt.Printf("%+v\n", a)
            if a {
                elevio.SetMotorDirection(elevio.MD_Stop)
            } else {
                elevio.SetMotorDirection(d)
            }
            
        case a := <- drv_stop:
            fmt.Printf("%+v\n", a)
            for f := 0; f < numFloors; f++ {
                for b := elevio.ButtonType(0); b < 3; b++ {
                    elevio.SetButtonLamp(b, f, false)
                }
            }
            */
        }
        
    }    
}
