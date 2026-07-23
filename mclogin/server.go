package mclogin

import (
	"log"
	"net"
)

func StartMCLoginServer(port int) {
	server, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
	if err != nil {
		log.Fatalf("could start MC login server (:%d): %+v", port, err)
	}

	go acceptConnections(server)
	log.Printf("Started MC login server on :%d", port)
}

func acceptConnections(server *net.TCPListener) {
	for {
		conn, err := server.AcceptTCP()
		if err != nil {
			log.Printf("error accepting TCP Client")
			continue
		}
		conn.SetKeepAlive(true)

		go NewMCLoginClient(conn).acceptPackets()
	}
}
