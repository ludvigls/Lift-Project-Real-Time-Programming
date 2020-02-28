package main

import (
	"./network/bcast"
	"./network/localip"
	"./network/peers"
	"flag"
	"fmt"
	"os"
	"time"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.
type AliveMsg struct {
	Message string 
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

func main() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	
	var id string
	var count_glob int

	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	
	//Useless for now
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}


	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
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
		go counter(countCh, 0)
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
					time.Sleep(500 * time.Millisecond)
				}
		}
	}(idCh)
	


	fmt.Println("Started with id: ", id)
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			if (len(p.Lost) > 0) { //someone lost
			    if p.Lost[0] == "1" && id == "2" { // Secondary lost heartbeat from primary
			    	fmt.Printf("I SHOULD BECOME PRIMARY and count from %d \n", count_glob)
			    	id = "1"

			    	// TODO, only sent once, not correct way (works with no package loss)
			    	idCh <- id

			    	// Update about role change, 2 becomes 1. Aka 2 lost, 1 new.
			    	//p.New = "1"
			    	//p.Lost[0] = "2"
			    	//p.Peers[0] = "1"
	
			    	//TODO, This is shit and should not be like this
			    	go counter(countCh, count_glob)

			    	//peerUpdateCh <- peers.PeerUpdate{ p.Peers, p.New, p.Lost } // TODO, does not work with several lifts

			    } else if p.Lost[0] == "2" && id == "1" { // Primary lost heartbeat from secondary
			    	fmt.Printf("I SHOULD CREATE NEW BACKUP \n")
			    } else {
			    	fmt.Printf("We lost someone that isnt backup, should be handled... \n")
			    }
			}
		case a := <-aliveRx: //msg on channel aliveRx
			fmt.Printf("Recieving from: %s \n", a.Id)
			if id == "2" && a.Id == "1" { // If I'm secondary and msg from primary
				fmt.Printf("Received from primary: %#v\n", a)
				count_glob = a.Iter // backup the count
			}
		
		case a := <- countCh: // a LOCAL message only heard on local computer
			count_glob = a
			fmt.Printf("Primary counting... %d \n", count_glob)
		}
	}
}
