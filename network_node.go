package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"./fsm"
	"./io"
	"./orderdelegator"

	"./network/bcast"
	"./network/peers"
)

// Copies map
func copyMap(mapOriginal map[string]fsm.State) map[string]fsm.State {
	mapCopy := make(map[string]fsm.State)
	for k, v := range mapOriginal {
		mapCopy[k] = v
	}
	return mapCopy
}

// isMaster returns true if the ID is the smallest on the network (in the PeerList)
func isMaster(PeerList []string, ID int) bool {
	for i := 0; i < len(PeerList); i++ {
		peerID, _ := strconv.Atoi(PeerList[i])
		if peerID < ID {
			return false
		}
	}
	return true
}

// getMostRecentMsg, listens for messages for a while, gets the most recent message, only happens in initialization
func getMostRecentMsg(peerUpdateCh chan peers.PeerUpdate, PeerList []string) []string {
	timeOut := false
	timer := time.NewTimer(200 * time.Millisecond) //emptys the message stack for 200ms
	for !timeOut {
		select {
		case <-timer.C:
			timeOut = true
		case a := <-peerUpdateCh:
			PeerList = a.Peers
		}
	}
	return PeerList
}

func main() { // go run network_node.go -id=1 -liftPort=15657
	var idStr string
	var liftPort string
	var PeerList []string
	initialized := false
	globState := make(map[string]fsm.State)
	numFloors := 4

	//Get terminal parameters
	flag.StringVar(&idStr, "id", "", "id of this peer")
	flag.StringVar(&liftPort, "port", "", "port to the lift connected")
	flag.Parse()
	idInt, _ := strconv.Atoi(idStr)

	// Channel for receiving updates on the id's of the peers that are alive on network
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool) // We can disable/enable the transmitter after it has been started (This could be used to signal that we are somehow "unavailable".)

	// Channels for sending and receiving our custom data types over UDP
	localStateTx := make(chan fsm.State)
	localStateRx := make(chan fsm.State)
	globStateTx := make(chan map[string]fsm.State) //String because network module couldnt handle int
	globStateRx := make(chan map[string]fsm.State)
	unassignedOrderTx := make(chan fsm.Order)
	unassignedOrderRx := make(chan fsm.Order)
	assignedOrderTx := make(chan fsm.Order)
	assignedOrderRx := make(chan fsm.Order)

	//UDP Recievers
	go peers.Receiver(15647, peerUpdateCh)
	go bcast.Receiver(16570, localStateRx)
	go bcast.Receiver(16571, globStateRx)
	go bcast.Receiver(16572, unassignedOrderRx)
	go bcast.Receiver(16573, assignedOrderRx)

	//UDP Transmitters
	go peers.Transmitter(15647, idStr, peerTxEnable)
	go bcast.Transmitter(16570, localStateTx)
	go bcast.Transmitter(16571, globStateTx)
	go bcast.Transmitter(16572, unassignedOrderTx)
	go bcast.Transmitter(16573, assignedOrderTx)

	// Regular Go channels
	// Channels between Network and IO modules
	drv_buttons := make(chan io.ButtonEvent)
	drv_floors := make(chan int)

	//Channels between network and orderDelegator modules
	fsm_n_orderCh := make(chan fsm.Order, 1000)
	fsm_n_stateCh := make(chan fsm.State, 1000)
	n_fsm_orderCh := make(chan fsm.Order, 1000)

	//Channels between network and orderDelegator modules
	od_n_orderCh := make(chan fsm.Order, 1000)
	n_od_orderCh := make(chan fsm.Order, 1000)
	n_od_globstateCh := make(chan map[string]fsm.State, 1000)

	// Running the modules : fsm, orderDel and IO
	io.Init("localhost:"+liftPort, numFloors)
	go fsm.Fsm(drv_buttons, drv_floors, numFloors, fsm_n_orderCh, n_fsm_orderCh, fsm_n_stateCh, idInt)
	go orderdelegator.OrderDelegator(n_od_orderCh, od_n_orderCh, n_od_globstateCh, numFloors)
	go io.Io(drv_buttons, drv_floors)

	for {
		select {
		case p := <-peerUpdateCh:
			PeerList = p.Peers
			fmt.Printf("Peer update:\n")
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			if !initialized {
				PeerList = getMostRecentMsg(peerUpdateCh, PeerList) // tømme postkassa peerUpdateCh
				initialized = true
			}
			fmt.Printf("  Peers:    %q\n", PeerList)

			//Network is lost, all work as individual lifts
			if len(PeerList) == 0 {
				fmt.Println("Network connection lost, removed global state")
				globStateCopy := make(map[string]fsm.State)
				n_od_globstateCh <- globStateCopy
			}

			if isMaster(PeerList, idInt) || len(p.New) > 0 { //There is a new node on network
				newInt, _ := strconv.Atoi(p.New)
				globStateCopy := copyMap(globState)
				for potentialGhost, _ := range globStateCopy {
					potentialGhostInt, _ := strconv.Atoi(potentialGhost)
					if potentialGhostInt == -newInt {

						fmt.Println("Delegating caborders to recovered lift")
						for f := 0; f < numFloors; f++ {
							if globStateCopy[potentialGhost].ExeOrders[f*3+int(io.BT_Cab)] {
								assignedOrderTx <- fsm.Order{io.ButtonEvent{f, io.BT_Cab}, newInt}
							}
						}
						delete(globStateCopy, potentialGhost)
						globStateTx <- globStateCopy
					}
				}
				//Inform the new node about the global state
				fmt.Println("Informing a new node about the global state")
				globStateTx <- globStateCopy
			}

			// Ensures that no orders are lost
			if isMaster(PeerList, idInt) && len(p.Lost) > 0 && len(PeerList) > 0 { // Network is up, but someone is lost
				globStateCopy := copyMap(globState)
				for i := 0; i < len(p.Lost); i++ {
					fmt.Println("Lost a lift from network, will redelegate its non cab orders")
					for f := 0; f < numFloors; f++ {
						if globStateCopy[p.Lost[i]].ExeOrders[f*3+int(io.BT_HallUp)] {
							orderID, _ := strconv.Atoi(PeerList[0])
							n_fsm_orderCh <- fsm.Order{io.ButtonEvent{f, io.BT_HallUp}, orderID}
							globStateCopy[p.Lost[i]].ExeOrders[f*3+int(io.BT_HallUp)] = false // remove up orders
						}
						if globStateCopy[p.Lost[i]].ExeOrders[f*3+int(io.BT_HallDown)] {
							orderID, _ := strconv.Atoi(PeerList[0])
							n_fsm_orderCh <- fsm.Order{io.ButtonEvent{f, io.BT_HallDown}, orderID}
							globStateCopy[p.Lost[i]].ExeOrders[f*3+int(io.BT_HallDown)] = false // remove down order
						}
					}
					//Create backup state
					ghostID := "-" + p.Lost[i]
					globStateCopy[ghostID] = globStateCopy[p.Lost[i]]
					delete(globStateCopy, p.Lost[i]) // delete regular state
				}
				n_od_globstateCh <- globStateCopy
				globStateTx <- globStateCopy
			}
		case a := <-globStateRx:
			globState = a
			//fmt.Println(globState)
			if !isMaster(PeerList, idInt) {
				globState = a
				n_od_globstateCh <- copyMap(globState)
			}
		case a := <-unassignedOrderRx:
			n_od_orderCh <- a
		case a := <-fsm_n_stateCh:
			localStateTx <- a
		case a := <-fsm_n_orderCh:
			n_od_orderCh <- a //send order to master
			unassignedOrderTx <- a
		case a := <-localStateRx: // recieved local state from any lift
			globStateCopy := copyMap(globState)
			if isMaster(PeerList, idInt) {
				globStateCopy[strconv.Itoa(a.ID)] = a // update global state
				n_od_globstateCh <- globStateCopy     // send out global state on network
				globStateTx <- globStateCopy
			}
		case a := <-assignedOrderRx:
			if a.ID == idInt {
				n_fsm_orderCh <- a
			}
		case a := <-od_n_orderCh:
			if isMaster(PeerList, idInt) {
				assignedOrderTx <- a
				if a.ID == idInt {
					n_fsm_orderCh <- a
				}
			}
		}
	}
}
