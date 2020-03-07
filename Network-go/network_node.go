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
type AliveMsg struct {
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

func isMaster(PeersList []string, myID int) bool {
	if myID == -1 {
		return false
	}

	for i := 0; i < len(PeersList); i++ {
		peerID, _ := strconv.Atoi(PeersList[i])
		if peerID < myID {
			return false
		}
	}
	return true
}

func main() { //  `go run network_node.go -id=our_id`
	var id string
	var count_glob int
	var PeerList []string
	var hasBeenMaster bool

	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	id_int, _ := strconv.Atoi(id)

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	// if id == "" { //Useless for now
	// 	localIP, err := localip.LocalIP()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		localIP = "DISCONNECTED"
	// 	}
	// 	id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())

	// }

	if id == "" {
		id = "-1"
		id_int, _ = strconv.Atoi(id)
	}

	// We make a channel for receiving updates on the id's of the peers that are alive on network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)

	//go peers.Transmitter(15647, id, peerTxEnable)

	// We make channels for sending and receiving our custom data types
	aliveTx := make(chan AliveMsg)
	aliveRx := make(chan AliveMsg)

	countCh := make(chan int)
	idCh := make(chan string)

	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.

	// A transmitter and receiver transmitting and recieving to the same port
	go peers.Receiver(15647, peerUpdateCh)
	go bcast.Receiver(16569, aliveRx)
	if id != "-1" {
		fmt.Println("I HAS ID!!")
		go peers.Transmitter(15647, id, peerTxEnable)
		go bcast.Transmitter(16569, aliveTx)
	}

	//Everyone sends I'm alive functionality every sec
	go func(idCh chan string) {
		AliveMsg := AliveMsg{"I'm Alive", id, 0}
		for {
			select {
			case a := <-idCh:
				AliveMsg.ID = a
			default:
				AliveMsg.Iter = count_glob
				aliveTx <- AliveMsg
				time.Sleep(100 * time.Millisecond)
			}
		}
	}(idCh)

	fmt.Println("Initialized with id:", id)
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			PeerList = p.Peers

			// Initialize ID

			// if id == "-1" {
			// 	if len(p.Peers) == 1 {
			// 		id = "1"
			// 		id_int, _ = strconv.Atoi(id)
			// 		idCh <- id // ok cuz local message
			// 		peerTxEnable <- false
			// 		go peers.Transmitter(15647, id, peerTxEnable)
			// 	}
			// }

			if id == "-1" {
				time_out := false
				timer := time.NewTimer(10 * time.Millisecond) //uses Xms to get latest message
				for !time_out {
					select {
					case <-timer.C: // door is closing
						time_out = true
					case a := <-peerUpdateCh:
						PeerList = a.Peers
						//fmt.Println("PEER LIST!!!", PeerList[1])
					}
				}

				//Find highest ID in peers list
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

				fmt.Println("Initialized with id %s", id)
				go peers.Transmitter(15647, id, peerTxEnable)
				go bcast.Transmitter(16569, aliveTx)
			}

			// if id == "10000" { // not yet assigned an id
			// 	fmt.Printf("Initializing my id... \n \n")
			// 	highest_id := -1
			// 	p_id := -1
			// 	for i := 0; i < len(p.Peers); i++ { // check whos on the network
			// 		p_id, _ = strconv.Atoi(p.Peers[i])
			// 		if p_id > highest_id {
			// 			highest_id = p_id
			// 		}
			// 	}
			// 	id_int = highest_id + 1
			// 	id = strconv.Itoa(id_int)
			// 	fmt.Println(id_int)
			// }

			//fmt.Println(PeerList)
			if isMaster(PeerList, id_int) {
				fmt.Printf("I am primary and count from:  %d \n", count_glob)
				if !hasBeenMaster {
					go counter(countCh, count_glob)
					hasBeenMaster = true
				}
			}
		case a := <-aliveRx:
			//fmt.Println("alive message from: ", a.ID)
			id_i, _ := strconv.Atoi(a.ID)
			if isMaster(PeerList, id_i) { // Every node stores from primary
				count_glob = a.Iter // backup the count
			}
		case a := <-countCh: // LOCAL message only heard on local computer
			count_glob = a
			fmt.Printf("Primary counting: %d \n", count_glob)
		}
	}
}
