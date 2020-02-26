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
type HelloMsg struct {
	Message string
	Id      string
	Iter    int
}


func counter(countCh chan int, start_from int) {
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
	helloTx := make(chan HelloMsg)
	helloRx := make(chan HelloMsg)
	countCh := make(chan int)

	if id == "1" {
		go counter(countCh, 0)
	}

	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	
	// A transmitter and receiver transmitting and recieving to the same port
	go bcast.Transmitter(16569, helloTx)
	go bcast.Receiver(16569, helloRx)


	//Everyone sends I'm alive functionality every sec
	go func() {
		helloMsg := HelloMsg{"I'm Alive", id, 0}
		for {
			helloMsg.Iter = count_glob
			helloTx <- helloMsg
			time.Sleep(500 * time.Millisecond)
		}
	}()


	fmt.Println("Started with id: ", id)
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			if (len(p.Lost) > 0) { //someone lost
			    if p.Lost[0] == "1" { // Primary lost
			    	//Become primary
			    	fmt.Printf("I SHOULD BECOME PRIMARY and count from %d \n", count_glob)
			    	id = "1"

			    	//TODO, This is shit and should not be like this
			    	go counter(countCh, count_glob)
			    	// AT THIS POINT THIS NODE STOPS TRANSMITTING ITS IM ALIVE MESSAGES


			    } else if p.Lost[0] == "2" { // Lost backup
			    	//Create a new backup
			    	fmt.Printf("I SHOULD CREATE NEW BACKUP \n")
			    }
			}
		case a := <-helloRx: //msg on channel helloRx
			fmt.Printf("Recieving.... %s \n", a.Id)
			if id == "2" && a.Id == "1" { // If I'm secondary and msg from primary
				fmt.Printf("Received from primary: %#v\n", a)
				count_glob = a.Iter // backup the count
			}
		
		case a := <- countCh: // a LOCAL message only heard on local computer
			count_glob = a
			fmt.Printf("primary counting... %d \n", count_glob)
		}
	}
}
