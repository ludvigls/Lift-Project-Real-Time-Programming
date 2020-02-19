package main
import "./fsm"
import "./io"

func main(){

numFloors:=4
io.Init("localhost:15652", numFloors)


drv_buttons := make(chan io.ButtonEvent)
drv_floors  := make(chan int)
   

go io.Io(drv_buttons,drv_floors)

go fsm.Fsm(drv_buttons)




for{}

}