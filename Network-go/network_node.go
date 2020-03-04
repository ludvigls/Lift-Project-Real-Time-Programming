package main

import (
	"./network/bcast"
	"./network/peers"
	"flag"
	"fmt"
	"time"
	"strconv"
	// Needed for old 'if id == "" { '
	//"os"
	//"./network/localip"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.
type AliveMsg struct {
	Message string // Orders - id
	Id      string
	Iter    int // Will change to state for actual project
}

func counter(countCh chan<- int, start_from int) { // Will be fsm for actual project
	count := start_from
	for {
		count++
		countCh <- count
		time.Sleep(1* time.Second)
	}
}

func main() {	//  `go run network_node.go -id=our_id`
	var id string
	var count_glob int

	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	id_int, _ := strconv.Atoi(id)
	prev_role_id := -1

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)

	if id == "" {
		id_int = 0
		id = "0"
	}

	/*
	if id == "" { //Useless for now
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}
	*/

	// We make a channel for receiving updates on the id's of the peers that are alive on network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	aliveTx := make(chan AliveMsg)
	aliveRx := make(chan AliveMsg)

	countCh := make(chan int)
	idCh := make(chan string)

	if id == "1" {
		go counter(countCh, 0) // Fsm only run from master
	}
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	
	// A transmitter and receiver transmitting and recieving to the same port
	go bcast.Transmitter(16569, aliveTx)
	go bcast.Receiver(16569, aliveRx)

	//Everyone sends I'm alive functionality every sec
	go func(idCh chan string) {
		AliveMsg := AliveMsg{"I'm Alive", id, 0}
		for {
			select { 
				case a := <- idCh:
					AliveMsg.Id = a
				default:
					AliveMsg.Iter = count_glob
					aliveTx <- AliveMsg
					time.Sleep(100 * time.Millisecond)
				}
			}
		}(idCh)

	fmt.Println("Initialized with id:", id)
	for {
		//fmt.Printf("Ey \n", id)
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			if (id == "0") { // not yet assigned an id
				fmt.Printf("Initializing my id... \n \n")
				highest_id := -1
				p_id := -1
				for i:=0; i < len(p.Peers); i++ { // check whos on the network
					p_id, _ = strconv.Atoi(p.Peers[i])
					if (p_id > highest_id) {
						highest_id = p_id
					}
				}
				prev_role_id = id_int
				id_int = highest_id+1
				id = strconv.Itoa(id_int)

	    		if (id_int == 1) {
			    	fmt.Printf("Will become primary and count from:  %d \n", count_glob)
			    	go counter(countCh, count_glob) //TODO, This is MIGHT be shit (maybe a prev counter is running??)
	    		}

				idCh <- id // ok cuz local message
		    	go peers.Transmitter(15647, id, peerTxEnable) // create new
		    	peerTxEnable <- false // remove myself


			}

			if (len(p.Lost) > 0) { //someone lost
			    for i:=0; i < len(p.Lost); i++ {
			    	lost_id, _ := strconv.Atoi(p.Lost[i])

			    	if (lost_id < id_int) { // Role change
			    		if (lost_id == prev_role_id) { // I didnt lose anyone.
			    			break
			    		}

			    		fmt.Println("Lost someone smaller that wasnt me, will decrement this id: ", id)
			    		prev_role_id = id_int 
			    		id_int -= 1
			    		id = strconv.Itoa(id_int)

			    		if (id_int == 1) {
					    	fmt.Printf("Will become primary and count from:  %d \n", count_glob)
					    	go counter(countCh, count_glob) //TODO, This is MIGHT be shit (maybe a prev counter is running??)
			    		}

				    	idCh <- id // ok cuz local message
				    	peerTxEnable <- false
				    	go peers.Transmitter(15647, id, peerTxEnable)


			    	}// else {
			    	//	fmt.Println("I lost someone with larger id (or prev myself), so wont change id")
			    	//}
			    }
			}
		case a := <-aliveRx:
			//fmt.Printf("Recieving from id: %s, state:%d \n", a.Id, a.Iter)
			if id == "2" && a.Id == "1" { // Store counter if i'm secondary
				count_glob = a.Iter // backup the count
			}
		case a := <- countCh: // LOCAL message only heard on local computer
			count_glob = a
			fmt.Printf("Primary counting: %d \n", count_glob)
		}
	}
}
