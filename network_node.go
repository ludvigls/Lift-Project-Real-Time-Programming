package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"./fsm"
	"./io"
	"./orderDelegator"

	"./network/bcast"
	"./network/peers"
)

type CountMsg struct {
	Message string
	ID      string
	Iter    int
}

func counter(countCh chan<- int, startFrom int) {
	count := startFrom
	for {
		count++
		countCh <- count
		time.Sleep(1 * time.Second)
	}
}

func isMaster(PeerList []string, ID int) bool {
	// Returns the true if the ID is the smallest in the PeerList
	if ID == -1 {
		return false // Unitialized node cannot be master
	}
	for i := 0; i < len(PeerList); i++ {
		peerID, _ := strconv.Atoi(PeerList[i])
		if peerID < ID {
			return false
		}
	}
	return true
}

func initializeID(PeerList []string) int {
	//Initializing the id to highest_id+1
	highest_id := -1
	p_id := -1
	for i := 0; i < len(PeerList); i++ { //find the highest id
		p_id, _ = strconv.Atoi(PeerList[i])
		if p_id > highest_id {
			highest_id = p_id
		}
	}
	return highest_id + 1
}

// func getMostRecentMsg(peerUpdateCh chan peers.PeerUpdate) []string {
// 	time_out := false
// 	var PeerList []string

// 	timer := time.NewTimer(2000 * time.Millisecond) //emptys the message stack for 100ms
// 	fmt.Println("WAITING")
// 	for !time_out {
// 		select {
// 		case <-timer.C:
// 			fmt.Println("TIME OUT!!")
// 			time_out = true
// 		case a := <-peerUpdateCh:
// 			fmt.Println("UPDATING PEERLIST!!") // TODO CODE IS NEVER HERE!!!
// 			PeerList = a.Peers
// 		}
// 	}
// 	return PeerList
// }

func main() { // `go run network_node.go -id=our_id`
	var id string
	var count_glob int
	var PeerList []string
	var hasBeenMaster bool
	globState := make(map[int]fsm.State) //maybe remove idk
	numFloors := 4
	lift_port := "15657"

	flag.StringVar(&id, "id", "", "id of this peer")
	flag.StringVar(&lift_port, "lift_port", "", "lift port of my lift")
	flag.Parse()

	io.Init("localhost:"+lift_port, numFloors)
	id_int, _ := strconv.Atoi(id)

	if id == "" {
		id = "-1"
		id_int, _ = strconv.Atoi(id)
	}

	// We make a channel for receiving updates on the id's of the peers that are alive on network
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool) // We can disable/enable the transmitter after it has been started (This could be used to signal that we are somehow "unavailable".)

	// We make channels for sending and receiving our custom data types
	countTx := make(chan CountMsg)
	countRx := make(chan CountMsg)

	localStateTx := make(chan fsm.State)
	localStateRx := make(chan fsm.State)

	countCh := make(chan int)
	idCh := make(chan string)

	// GO ROUTINES EVERYONE WILL RUN v
	drv_buttons := make(chan io.ButtonEvent)
	drv_floors := make(chan int)
	fsm_n_order_chan := make(chan fsm.Order)
	n_od_order_chan := make(chan fsm.Order)
	od_n_order_chan := make(chan fsm.Order)
	n_fsm_order_chan := make(chan fsm.Order)
	globstate_chan := make(chan map[int]fsm.State)
	globstate_chanRXTX := make(chan map[int]fsm.State) //make this udp
	fsm_n_state_chan := make(chan fsm.State)

	//Every node initialized as pure recievers
	go peers.Receiver(15647, peerUpdateCh)
	go bcast.Receiver(16569, countRx)
	go bcast.Receiver(16570, localStateRx)

	if id != "-1" { //Nodes with IDs are allowed to transmit
		//fmt.Println("Starting transmitting from ID: ", id)
		go peers.Transmitter(15647, id, peerTxEnable)
		go bcast.Transmitter(16569, countTx)
		go bcast.Transmitter(16570, localStateTx)

		go fsm.Fsm(drv_buttons, drv_floors, numFloors, fsm_n_order_chan, n_fsm_order_chan, fsm_n_state_chan, id_int)

	}

	go io.Io(drv_buttons, drv_floors)
	go orderDelegator.OrderDelegator(n_od_order_chan, od_n_order_chan, globstate_chan, numFloors)

	//Everyone sends out its count msg
	go func(idCh chan string) {
		CountMsg := CountMsg{"I'm sending the global state of all lifts", id, 0}
		for {
			select {
			case a := <-idCh: //Needed when node is initialized without id
				CountMsg.ID = a
			default:
				CountMsg.Iter = count_glob //Everyone sends the global state in its alive message
				countTx <- CountMsg
				time.Sleep(100 * time.Millisecond)
			}
		}
	}(idCh)

	// TODO : dont be annonomous?? add comments "remove" placeholder functionality
	go func() {
		for {
			select {
			case a := <-globstate_chanRXTX:
				//NB, master now sends out glob state to port and saves same glob state from port
				globState = a
			case a := <-fsm_n_state_chan:
				//fmt.Println("Sending my state now")
				// globState[a.Id] = a
				// globstate_chan <- globState
				localStateTx <- a
			case a := <-fsm_n_order_chan:
				n_od_order_chan <- a //send order to master

			}
		}
	}()

	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			PeerList = p.Peers

			// Initialize ID
			if id == "-1" {
				//Empty recieve queue (get last msg)
				time_out := false
				timer := time.NewTimer(100 * time.Millisecond) //uses Xms to get latest message
				for !time_out {
					select {
					case <-timer.C:
						time_out = true
					case a := <-peerUpdateCh:
						PeerList = a.Peers
					}
				}

				//PeerList = getMostRecentMsg(peerUpdateCh)
				//print(PeerList)
				id_int = initializeID(PeerList)
				id = strconv.Itoa(id_int)

				//Initialize transmit features for node
				fmt.Println("Transmitting with ID: ", id_int)
				go peers.Transmitter(15647, id, peerTxEnable)
				go bcast.Transmitter(16569, countTx)
				go bcast.Transmitter(16570, localStateTx)

				go fsm.Fsm(drv_buttons, drv_floors, numFloors, fsm_n_order_chan, n_fsm_order_chan, fsm_n_state_chan, id_int)
				idCh <- id
			}

			// Should I become master?
			if isMaster(PeerList, id_int) {
				//fmt.Printf("I am primary and count from:  %d \n", count_glob)
				if !hasBeenMaster {
					go counter(countCh, count_glob)
					//go orderDelegator.OrderDelegator(n_od_order_chan, od_n_order_chan, globstate_chan, numFloors)
					hasBeenMaster = true
				}
			}
		case a := <-countRx:
			id_i, _ := strconv.Atoi(a.ID)
			if isMaster(PeerList, id_i) {
				count_glob = a.Iter // Every nodes backups masters state
			}
		case a := <-localStateRx: // recieved local state from any lift
			if isMaster(PeerList, id_int) {
				globState[a.Id] = a             // update global state
				globstate_chan <- globState     // send out global state on network
				globstate_chanRXTX <- globState //??
			}

		case a := <-countCh: // LOCAL message only heard on local computer
			count_glob = a
			//fmt.Printf("Primary counting: %d \n", count_glob) // Counting only happening from master

//		case a := <-fsm_n_order_chan:
//			n_od_order_chan <- a //send order to master

		case a := <-od_n_order_chan:
			if isMaster(PeerList, id_int) {
				if a.Id == id_int {
					n_fsm_order_chan <- a
				}
			}
			//else { send ut på nettet }
			//send order to master

			//add case for incoming message from master with new orders, send to fsm
		}
	}
}
