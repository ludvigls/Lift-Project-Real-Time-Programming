package fsm

import "../io"
import "fmt"

func Fsm(drv_buttons chan io.ButtonEvent){

    //numFloors := 4
//15657
    //io.Init("localhost:15652", numFloors)
    
    //var d elevio.MotorDirection = elevio.MD_Up
    //elevio.SetMotorDirection(d)
    

    
    
    for {
        select {
        
        case a := <- drv_buttons:
            fmt.Printf("%+v\n", a)
            io.SetButtonLamp(a.Button, a.Floor, true)
        /*  
        case a := <- drv_floors:
            fmt.Printf("%+v\n", a)
            if a == numFloors-1 {
                d = elevio.MD_Down
            } else if a == 0 {
                d = elevio.MD_Up
            }
            elevio.SetMotorDirection(d)
            
           
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
