package javasock

import "io"
import "bufio"
import "github.com/davidcallanan/gomcp/javaio"

type client struct {
	state int
	input *bufio.Reader
	output *bufio.Writer
	closeCallback func()
	isClosed bool
}

func (client *client) close() {
	client.closeCallback()
	client.isClosed = true
}

type Server struct {
}

func NewServer() Server {
	return Server {
	}
}

func (server *Server) AddConnection(input io.Reader, output io.Writer, closeCallback func()) {
	client := &client {
		state: javaio.StateHandshaking,
		input: bufio.NewReader(input),
		output: bufio.NewWriter(output),
		closeCallback: closeCallback,
	}

	go func() {
		for !client.isClosed {
			server.handleReceive(client)
		}
	}()
}

func (server *Server) handleReceive(client *client) {
	packet, err := javaio.ParseServerboundPacketUncompressed(client.input, client.state)

	if err != nil {
		switch err.(type) {
		case *javaio.UnsupportedPayloadError:
			return
		case *javaio.MalformedPacketError:
			client.close()
			return
		default:
			panic(err)
		}
	}

	switch p := packet.(type) {
	case javaio.Handshake:
		server.ProcessHandshake(client, p)
	case javaio.StatusRequest:
		println("Ignored status request")
	case javaio.Ping:
		server.ProcessPing(client, p)
	default:
		panic("Unrecognized packet type")
	}
}

func (server *Server) ProcessHandshake(client *client, handshake javaio.Handshake) {
	client.state = handshake.NextState
}

func (server *Server) ProcessPing(client *client, ping javaio.Ping) {
	javaio.EmitPong(javaio.Pong {
		Payload: ping.Payload,
	}, client.output)
}
