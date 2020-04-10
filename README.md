# Lift project in Go - Real-time Programming (TTK4145)
Project in the course Real-time Programming (https://www.ntnu.edu/studies/courses/TTK4145#tab=omEmnet)

# Project description
The purpose of this project was creating software for controlling n elevators working in parallel across m floors. The software was run both on a physical model lift and in a lift simulator. We chose to program in Go as we saw it fit for the task and wanted to become familiar with a new programming language. This is our first project with network communication and multithreading. 

Course github page : https://github.com/TTK4145/Project
Project description : https://github.com/TTK4145/Project
UDP go Nework driver : https://github.com/TTK4145/Network-go
Lift simulator: https://github.com/TTK4145/Simulator-v2

# Implementation
We will utilize a master slave architecture. The slaves will send their incoming tasks to master, who will distribute them to the most fit elevator. The system will be fault tolerant. For example, each node needs to operate independently in cases where they drop out of the network or when an elevator experiences power loss. The system will also handle the master dropping out by having a backup slave lift taking over the master role when necessary. The end result will be a scalable and robust system, making sensible decisions for an underdetermined amount of elevators.

As seen in our communication overview (CommunicationOverview.png) our system consists of several modules. As this is slightly outdated and the master/slave logic and Communication has been merged into the one network module (network_node.go). The IO module communicates with the physical lift or the lift simulator. The FSM (Finite state machine) controls an individual lift. The Order delegator is only active for the master lift, it takes in orders from the other lifts and delegates them according to a cost function. The network module is responsible for sending information to other lifts over UDP communication. 

# How to build and run

**Running one lift simulator** : `./SimElevatorServer --port 15657`
This repository contains the SimElevatorServer bin file built for Linux x86-64. The simulator can be cloned and built for several systems from this repository https://github.com/TTK4145/Simulator-v2. the `15657` port is an arbitrary choice and other ports needs to be used for running several lift simulators at once. 

**Running one lift controller** : `go run network_node.go -id=1 -port=15657`
This process starts up the network module which initializes the IO, FSM and orderDelegator. The port should match the port of the lift simulator or the physical lift. All lifts must have unique ids. The master lift will be the lift with the lowest id. 
