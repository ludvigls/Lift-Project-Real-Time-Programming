package io


func Io(drv_buttons chan<- ButtonEvent,drv_floors chan<- int){
	go PollButtons(drv_buttons)
	go PollFloorSensor(drv_floors)
}
