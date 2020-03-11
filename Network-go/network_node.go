package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"./network/bcast"
	"./network/peers"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.
type GlobStateMsg struct {
	Message string
	ID      string // TO ID
	Iter    int
}

func counter(countCh chan<- int, startFrom int) {
	// Will be replaced with a spam orders --> glob order converter
	count := startFrom
	for {
		count++
		countCh <- count
		time.Sleep(1 * time.Second)
	}
}

func isMaster(PeersList []string, ID int) bool {
	// Returns the true if the ID is the smallest in the PeerList
	if ID == -1 {
		return false // Unitialized node cannot be master
	}
	for i := 0; i < len(PeersList); i++ {
		peerID, _ := strconv.Atoi(PeersList[i])
		if peerID < ID {
			return false 
		}
	}
	return true
}

func main() { // `go run network_node.go -id=our_id`
	var id string
	var count_glob int
	var PeerList []string
	var hasBeenMaster bool

	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	id_int, _ := strconv.Atoi(id)

	if id == "" {
		id = "-1"
		id_int, _ = strconv.Atoi(id)
	}

	// We make a channel for receiving updates on the id's of the peers that are alive on network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)

	// We make channels for sending and receiving our custom data types
	globStateTx := make(chan GlobStateMsg)
	globStateRx := make(chan GlobStateMsg)

	countCh := make(chan int)
	idCh := make(chan string)

	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	// A transmitter and receiver transmitting and recieving to the same port

	//Every node initialized as pure recievers
	go peers.Receiver(15647, peerUpdateCh)
	go bcast.Receiver(16569, globStateRx)
	if id != "-1" { //Nodes with IDs are allowed to transmit
		fmt.Println("Starting transmitting from ID: ", id)
		go peers.Transmitter(15647, id, peerTxEnable)
		go bcast.Transmitter(16569, globStateTx)
	}

	//Everyone sends out its global state (alive message)
	go func(idCh chan string) {
		GlobStateMsg := GlobStateMsg{"I'm sending the global state of all lifts", id, 0}
		for {
			select {
				case a := <-idCh: //Needed when node is initialized without id
					GlobStateMsg.ID = a
				default:
					GlobStateMsg.Iter = count_glob //Everyone sends the global state in its alive message
					globStateTx <- GlobStateMsg
					time.Sleep(100 * time.Millisecond)
			}
		}
	}(idCh)

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

				//Initialize to highest_ID in peers list + 1
				highest_id := -1
				p_id := -1
				for i := 0; i < len(PeerList); i++ { //find the highest id
					p_id, _ = strconv.Atoi(PeerList[i])
					if p_id > highest_id {
						highest_id = p_id
					}
				}
				id_int = highest_id + 1
				id = strconv.Itoa(id_int)

				//Initialize transmit features for node
				fmt.Println("Transmitting with ID: ", id_int)
				go peers.Transmitter(15647, id, peerTxEnable)
				go bcast.Transmitter(16569, globStateTx)
				idCh <- id
			}

			// Should I become master?
			if isMaster(PeerList, id_int) {
				//fmt.Printf("I am primary and count from:  %d \n", count_glob)
				if !hasBeenMaster {
					go counter(countCh, count_glob)
					//Order deligator
					//take in local msg --> one global msg
					hasBeenMaster = true
				}
			}
		case a := <-globStateRx:
			id_i, _ := strconv.Atoi(a.ID)
			if isMaster(PeerList, id_i) {
				count_glob = a.Iter // Every nodes backups masters state
			}
		case a := <-countCh: // LOCAL message only heard on local computer
			count_glob = a
			fmt.Printf("Primary counting: %d \n", count_glob) // Counting only happening from master
		}
	}
}
