package javaio

import "bytes"
import "bufio"

/**  Serverbound packet emission is not implemented.  **/

// Clientbound

func EmitClientboundPacketUncompressed(packet interface{}, state int, output *bufio.Writer) {
	var packetId int32 = -1
	var packetIdBuf bytes.Buffer
	var dataBuf bytes.Buffer
	packetIdWriter := bufio.NewWriter(&packetIdBuf)
	dataWriter := bufio.NewWriter(&dataBuf)

	switch state {
	case StateHandshaking:
		panic("Packet cannot be emitted in handshaking state")
	case StateStatus:
		switch packet := packet.(type) {
		case *StatusResponse:
			packetId = 0x00
			EmitStatusResponse(*packet, dataWriter)
		case *Pong:
			packetId = 0x01
			EmitPong(*packet, dataWriter)
		default:
			panic("Packet cannot be emitted in status state")
		}
	case StateLogin:
		switch packet := packet.(type) {
		case *LoginSuccess:
			packetId = 0x02
			EmitLoginSuccess(*packet, dataWriter)	
		default:
			panic("Packet cannot be emitted in login state")
		}
	case StatePlay:
		switch packet := packet.(type) {
		case *JoinGame:
			packetId = 0x26
			EmitJoinGame(*packet, dataWriter)	
		default:
			panic("Packet cannot be emitted in play state (likely because not implemented)")
		}
	default:
		panic("State does not match one of non-invalid predefined enum types")
	}
	
	if packetId == -1 {
		panic("Internal package bug: packet id was not set while preparing to emit a packet")
	}

	EmitVarInt(packetId, packetIdWriter)
	dataWriter.Flush()
	packetIdWriter.Flush()
	length := packetIdBuf.Len() + dataBuf.Len()
	lengthInt32 := int32(length)

	if length > int(lengthInt32) {
		panic("Emitted packet data was too large to hold its size in VarInt")
	}

	EmitVarInt(lengthInt32, output)
	output.Write(packetIdBuf.Bytes())
	output.Write(dataBuf.Bytes())
	output.Flush()
}

func EmitClientboundPacketCompressed(packet interface{}, state int, output *bufio.Writer) {
	panic("EmicClientboundPacketCompressed not implemented")
}
