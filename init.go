package main

import (
	"./fsm"
	"./io"
	"./orderDelegator"

  "./timer"
)

func main() {

	numFloors := 4
	numElev := 2
	io.Init("localhost:15657", numFloors)

	drv_buttons := make(chan io.ButtonEvent)
	drv_floors := make(chan int)
	order_chan := make(chan fsm.Order)
	state_chan := make(chan fsm.State)

  timer_chan := make(chan int)

	go io.Io(drv_buttons, drv_floors)

	go fsm.Fsm(drv_buttons, drv_floors, numFloors, order_chan, state_chan, 1)
	go orderDelegator.OrderDelegator(order_chan, state_chan, numFloors, numElev, timer_chan)

  go timer.Timer_organizer(timer_chan)
  go timer.Timer()
	for {
	}

}
