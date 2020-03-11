package timer

import(
  "fmt"
  "time"
)

var timer_task_timeout_array_lenght int = 20 *12
var timer_task_timeout_array[20 * 12] int //guessing that we wont have more than 20 elevators
var init_flag int = 1



func Timer_organizer(timer_chan chan int){
  fmt.Println("BIG FUCKING WORDS")
  if init_flag == 1 {
    init_flag = 0
    for c := 0; c < timer_task_timeout_array_lenght; c++ {
      timer_task_timeout_array[c] = -1;
    }
  }
  new_order_index := <- timer_chan
  timer_task_timeout_array[new_order_index] = 16
  fmt.Println("Belzebob")
}

func Timer(){
  decrement_timer := time.NewTimer(1 * time.Second)
  <-decrement_timer.C
  for c := 0; c < timer_task_timeout_array_lenght; c++ {
    if timer_task_timeout_array[c] > 0 {
      timer_task_timeout_array[c]--
      fmt.Println("index:  \t", c)
      fmt.Println("timer =  \t", timer_task_timeout_array[c])
    }
    if (timer_task_timeout_array[c] == 0){
      fmt.Println("TASK TIMEOUT ALARM \t task:  \t", c)
    }
  }
}
