package mclogin

import (
	"log"
	"net"
	"website/webserver"

	"github.com/krasseGOrganisation/krasserGoMinecraftServer/protocol/netparse"
	serverboundHandshake "github.com/krasseGOrganisation/krasserGoMinecraftServer/protocol/nettypes/packets/handshake/serverbound"
	serverboundLogin "github.com/krasseGOrganisation/krasserGoMinecraftServer/protocol/nettypes/packets/login/serverbound"
	"github.com/krasseGOrganisation/krasserGoMinecraftServer/protocol/nettypes/protocolids"
)

type MCLoginClient struct {
	net.TCPConn

	protocolVersion    int32
	communicationState protocolids.CommunicationState
}

func NewMCLoginClient(conn *net.TCPConn) (client *MCLoginClient) {
	return &MCLoginClient{
		TCPConn:            *conn,
		protocolVersion:    -1,
		communicationState: protocolids.STATE_HANDSHAKE,
	}
}

func (client MCLoginClient) State() protocolids.CommunicationState {
	return client.communicationState
}

func (client *MCLoginClient) SetState(state protocolids.CommunicationState) {
	client.communicationState = state
}

func (client MCLoginClient) ProtocolVersion() int32 {
	return client.protocolVersion
}

func (client *MCLoginClient) SetProtocolVersion(protocolId int32) {
	client.protocolVersion = protocolId
}

func (client *MCLoginClient) acceptPackets() {
readLoop:
	for {
		pid, data := netparse.ReadPacket(client, true)
		if pid == (protocolids.PacketIdentifier{}) {
			break
		}
		packet, ok := pid.Packet()
		if !ok {
			log.Printf("Received unknown packet")
		}

		err := packet.FromBytes(client.ProtocolVersion(), data)
		if err != nil {
			log.Printf("Error parsing packet %s: %v", pid, err)
			break
		}

		if packet.GetIdentifier().State == protocolids.STATE_HANDSHAKE {
			// only allow a hand
			p, ok := packet.(*serverboundHandshake.Intention)
			if !ok || p.GetNextCommunicationState() != protocolids.STATE_LOGIN {
				break
			}
			client.SetProtocolVersion(p.ProtocolVersion)
			client.SetState(protocolids.STATE_LOGIN)
			continue
		} else if packet.GetIdentifier().State != protocolids.STATE_LOGIN {
			break
		}
		switch packet := packet.(type) {
		case *serverboundLogin.Hello:
			ch, ok := webserver.MCLoginHasActiveLogin(packet.UUID)
			if !ok {
				break readLoop
			}
			//TODO: set encription
			// netwrite.Write(client, &clientbound.Disconnect{
			// 	Reason: "Success",
			// })
			close(ch)
			break readLoop
		default:
			// just drop connection for other packets
			break readLoop
		}
	}
	client.TCPConn.Close()
}
