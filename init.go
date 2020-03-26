// package main

// import (
// 	"./fsm"
// 	"./io"
// 	"./orderDelegator"
// )

// func main() {

// 	numFloors := 4
// 	numElev := 2
// 	io.Init("localhost:15657", numFloors)

// 	drv_buttons := make(chan io.ButtonEvent)
// 	drv_floors := make(chan int)
// 	orderCh := make(chan fsm.Order)
// 	globstateCh := make(chan map[int]fsm.State)
// 	localstateCh := make(chan fsm.State)
// 	go io.Io(drv_buttons, drv_floors)

// 	go fsm.Fsm(drv_buttons, drv_floors, numFloors, orderCh, localstateCh, 1)
// 	go orderDelegator.OrderDelegator(orderCh, globstateCh, numFloors, numElev)
// 	for {
// 	}

// }
