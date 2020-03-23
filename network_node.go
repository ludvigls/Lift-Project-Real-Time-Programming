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

// CountMsg is a struct sending an alive message with a from id
type CountMsg struct {
	Message string
	ID      int // from
	Iter    int
}

// counter, func for testing network
func counter(countCh chan<- int, startFrom int) {
	count := startFrom
	for {
		count++
		countCh <- count
		time.Sleep(1 * time.Second)
	}
}

// isMaster returns true if the ID is the smallest on the network (in the PeerList)
func isMaster(PeerList []string, ID int) bool {
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

// initializeID initializes the ID to the id to highest_id+1
func initializeID(PeerList []string) int {
	highestID := -1
	peerID := -1
	for i := 0; i < len(PeerList); i++ {
		peerID, _ = strconv.Atoi(PeerList[i])
		if peerID > highestID {
			highestID = peerID
		}
	}
	return highestID + 1
}

// getMostRecentMsg, listens for messages for a while, gets the most recent message
func getMostRecentMsg(peerUpdateCh chan peers.PeerUpdate, PeerList []string) []string {
	//TODO : this function was made to prevent some bug
	//Atm it does nothing, but things still work...
	timeOut := false
	timer := time.NewTimer(200 * time.Millisecond) //emptys the message stack for 100ms
	//fmt.Println("WAITING")
	for !timeOut {
		select {
		case <-timer.C:
			//fmt.Println("TIME OUT!!")
			timeOut = true
		case a := <-peerUpdateCh:
			fmt.Println("THIS FUNCTION HAS A PURPOSE :0 !, UPDATING PEERLIST!!") // TODO CODE IS NEVER HERE!!!
			PeerList = a.Peers
		}
	}
	return PeerList
}

func main() { // `go run network_node.go -id=our_id` -liftPort=15657
	var idStr string
	var count_glob int
	var PeerList []string
	var hasBeenMaster bool
	globState := make(map[int]fsm.State)
	numFloors := 4
	liftPort := "15657"

	flag.StringVar(&idStr, "id", "", "id of this peer")
	flag.StringVar(&liftPort, "port", "", "port to the lift connected")
	flag.Parse()

	io.Init("localhost:"+liftPort, numFloors)

	if idStr == "" {
		idStr = "-1"
	}

	idInt, _ := strconv.Atoi(idStr)

	// We make a channel for receiving updates on the id's of the peers that are alive on network
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool) // We can disable/enable the transmitter after it has been started (This could be used to signal that we are somehow "unavailable".)

	// We make channels for sending and receiving our custom data types
	countTx := make(chan CountMsg)
	countRx := make(chan CountMsg)

	localStateTx := make(chan fsm.State)
	localStateRx := make(chan fsm.State)

	countCh := make(chan int)
	idCh := make(chan int)

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

	if idInt != -1 { //Nodes with IDs are allowed to transmit
		//fmt.Println("Starting transmitting from ID: ", id)
		go peers.Transmitter(15647, idStr, peerTxEnable)
		go bcast.Transmitter(16569, countTx)
		go bcast.Transmitter(16570, localStateTx)

		go fsm.Fsm(drv_buttons, drv_floors, numFloors, fsm_n_order_chan, n_fsm_order_chan, fsm_n_state_chan, idInt)
	}

	go io.Io(drv_buttons, drv_floors)
	go orderDelegator.OrderDelegator(n_od_order_chan, od_n_order_chan, globstate_chan, numFloors)

	//Everyone sends out its count msg
	go func(idCh chan int) {
		CountMsg := CountMsg{"I'm sending the global state of all lifts", idInt, 0}
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

			//TODO - oppdater globstate når flere heiser blir borte

			// update globState

			//send out globState

			if idInt == -1 {
				PeerList = getMostRecentMsg(peerUpdateCh, PeerList)
				idInt = initializeID(PeerList)

				//Initialize transmit features for node
				fmt.Println("Transmitting with ID: ", idInt)
				go peers.Transmitter(15647, strconv.Itoa(idInt), peerTxEnable)
				go bcast.Transmitter(16569, countTx)
				go bcast.Transmitter(16570, localStateTx)

				go fsm.Fsm(drv_buttons, drv_floors, numFloors, fsm_n_order_chan, n_fsm_order_chan, fsm_n_state_chan, idInt)
				idCh <- idInt
			}

			// Should I become master?
			if isMaster(PeerList, idInt) {
				//fmt.Printf("I am primary and count from:  %d \n", count_glob)
				if !hasBeenMaster {
					go counter(countCh, count_glob)
					//go orderDelegator.OrderDelegator(n_od_order_chan, od_n_order_chan, globstate_chan, numFloors)
					hasBeenMaster = true
				}
			}
		case a := <-countRx:
			idPeer := a.ID
			if isMaster(PeerList, idPeer) {
				count_glob = a.Iter // Every nodes backups masters state
			}
		case a := <-localStateRx: // recieved local state from any lift
			if isMaster(PeerList, idInt) {
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
			if isMaster(PeerList, idInt) {
				if a.Id == idInt {
					n_fsm_order_chan <- a
				}
			}
			//else { send ut på nettet }
			//send order to master

			//add case for incoming message from master with new orders, send to fsm
		}
	}
}
