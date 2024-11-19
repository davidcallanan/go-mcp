package javaserver

import "io"
import "math"
import "time"
import "bufio"
import "github.com/davidcallanan/go-mcp/javaio"
import "github.com/google/uuid"

type Connection struct {
	ctx javaio.ClientContext
	inputStream *bufio.Reader
	outputStream *bufio.Writer
	endStream func()
	eventHandlers EventHandlers
	isClosed bool
}

type EventHandlers struct {
	OnStatusRequestV1 func() StatusResponseV1
	OnStatusRequestV2 func() StatusResponseV2
	OnStatusRequestV3 func() StatusResponseV3
	OnPlayerJoinRequest func(data PlayerJoinRequest) PlayerJoinResponse
	OnPlayerJoin func()
	OnPlayerMove func(data PlayerMove)
}

func NewConnection(stream io.ReadWriter, endStream func(), eventHandlers EventHandlers) *Connection {
	conn := &Connection {
		ctx: javaio.InitialClientContext,
		inputStream: bufio.NewReader(stream),
		outputStream: bufio.NewWriter(stream),
		endStream: endStream,
		eventHandlers: eventHandlers,
		isClosed: false,
	}
	
	go func() {
		conn.receiveLoop()
	}()

	go func() {
		conn.keepAliveLoop()
	}()

	return conn
}

func (conn *Connection) receiveLoop() {
	for !conn.isClosed {
		conn.handleReceive()
	}
}

func (conn *Connection) keepAliveLoop() {
	timer := time.NewTicker(time.Second * 20)
	
	for now := range timer.C {
		if conn.isClosed {
			break
		}
		if conn.ctx.State != javaio.StatePlay {
			continue
		}
	
		conn.send(javaio.KeepAlive {
			Payload: now.Unix(),
		})
	}
}

func (conn *Connection) close() {
	conn.endStream()
	conn.isClosed = true
}

type StatusResponseV1 struct {
	// Color-coding is not supported.
	// Description is treated as plain-text.
	// Section character must not be used, otherwise there will be undefined behaviour.
	PreventResponse bool
	Description string
	MaxPlayers int
	OnlinePlayers int
}

type StatusResponseV2 struct {
	PreventResponse bool
	IsClientSupported bool
	Version string
	Description string
	MaxPlayers int
	OnlinePlayers int
}

type StatusResponseV3 struct {
	PreventResponse bool
	IsClientSupported bool
	Version string
	Description string
	FaviconPng []byte
	MaxPlayers int
	OnlinePlayers int
	PlayerSample []string
}

type PlayerJoinRequest struct {
	ClientsideUsername string
}

type PlayerJoinResponse struct {
	PreventResponse bool
	Uuid uuid.UUID
}

type PlayerMove struct {
	HasPos bool
	HasLook bool
	X float64
	Y float64
	Z float64
	Yaw float32
	Pitch float32
	OnGround bool
}

func (conn *Connection) send(packet interface{}) {
	javaio.EmitClientboundPacketUncompressed(packet, conn.ctx, conn.outputStream)
}

func (conn *Connection) handleReceive() {
	packet, err := javaio.ParseServerboundPacketUncompressed(conn.inputStream, conn.ctx, conn.ctx.State)

	if err != nil {
		switch err.(type) {
		case javaio.UnsupportedPayloadError:
			println("Unsupported payload from client")
			return
		case javaio.MalformedPacketError:
			println("Malformed packet from client.. closing connection")
			conn.close()
			return
		default:
			panic(err)
		}
	}

	switch packet := packet.(type) {
		// Determining Protocol
	case javaio.ProtocolDetermined:
		conn.processProtocolDetermined(packet)

		// Handshaking
	case javaio.Handshake:
		conn.processHandshake(packet)

		// Status
	case javaio.Packet_0051_StatusRequest:
		conn.processStatusRequest(packet)
	case javaio.Packet_0051_Ping:
		conn.processPing(packet)

		// Login
	case javaio.LoginStart:
		conn.processLoginStart(packet)

		// Play
	case javaio.Packet_PlayerPosSb:
		conn.processMovePos(packet)
	case javaio.Packet_PlayerLookSb:
		conn.processMoveLook(packet)
	case javaio.Packet_PlayerPosAndLookSb:
		conn.processMoveAll(packet)

		// Pre-Netty
	case javaio.Packet_002E_StatusRequest:
		conn.processLegacyStatusRequest(packet)

		// Very Pre-Netty
	case javaio.VeryLegacyStatusRequest:
		conn.processVeryLegacyStatusRequest(packet)

		// Default
	default:
		// println("Unrecognized packet type")
	}
}

func (conn *Connection) processProtocolDetermined(data javaio.ProtocolDetermined) {
	conn.ctx.State = data.NextState
}

func (conn *Connection) processHandshake(handshake javaio.Handshake) {
	conn.ctx.Protocol = javaio.DecodePostNettyVersion(handshake.Protocol)
	conn.ctx.State = handshake.NextState
}

func (conn *Connection) processStatusRequest(_ javaio.Packet_0051_StatusRequest) {
	if conn.eventHandlers.OnStatusRequestV3 == nil {
		return
	}

	res := conn.eventHandlers.OnStatusRequestV3()

	if (res.PreventResponse) {
		return
	}

	protocol := int32(0)
	if (res.IsClientSupported) {
		protocol = javaio.EncodePostNettyVersion(conn.ctx.Protocol)
	}

	playerSample := make([]javaio.Packet_0051_StatusResponse_Player, len(res.PlayerSample), len(res.PlayerSample))

	for i, text := range res.PlayerSample {
		playerSample[i] = javaio.Packet_0051_StatusResponse_Player {
			Name: text,
			Uuid: "65bd239f-89f2-4cc7-ae8b-bb625525904e",
		}
	}

	conn.send(javaio.Packet_0051_StatusResponse {
		Protocol: protocol,
		Version: res.Version,
		Description: res.Description,
		MaxPlayers: res.MaxPlayers,
		OnlinePlayers: res.OnlinePlayers,
		PlayerSample: playerSample,
	})
}

func (conn *Connection) processLegacyStatusRequest(_ javaio.Packet_002E_StatusRequest) {
	if conn.eventHandlers.OnStatusRequestV2 == nil {
		return
	}

	res := conn.eventHandlers.OnStatusRequestV2()

	if (res.PreventResponse) {
		return
	}

	protocol := 0
	if (res.IsClientSupported) {
		protocol = int(conn.ctx.Protocol)
	}

	conn.send(javaio.Packet_002E_StatusResponse {
		Protocol: protocol,
		Version: res.Version,
		Description: res.Description,
		MaxPlayers: res.MaxPlayers,
		OnlinePlayers: res.OnlinePlayers,
	})
}

func (conn *Connection) processVeryLegacyStatusRequest(_ javaio.VeryLegacyStatusRequest) {
	if conn.eventHandlers.OnStatusRequestV1 == nil {
		return
	}
	
	res := conn.eventHandlers.OnStatusRequestV1()

	if (res.PreventResponse) {
		return
	}

	conn.send(javaio.VeryLegacyStatusResponse {
		Description: res.Description,
		MaxPlayers: res.MaxPlayers,
		OnlinePlayers: res.OnlinePlayers,
	})
}

func (conn *Connection) processPing(ping javaio.Packet_0051_Ping) {
	conn.send(javaio.Packet_0051_Pong {
		Payload: ping.Payload,
	})
}

func (conn *Connection) processLoginStart(data javaio.LoginStart) {
	if conn.eventHandlers.OnPlayerJoinRequest == nil {
		return
	}

	res := conn.eventHandlers.OnPlayerJoinRequest(PlayerJoinRequest {
		ClientsideUsername: data.ClientsideUsername,
	})

	if res.PreventResponse {
		return
	}
	
	playerUuid := res.Uuid

	conn.send(javaio.LoginSuccess {
		Uuid: playerUuid,
		Username: data.ClientsideUsername,
	})

	conn.ctx.State = javaio.StatePlay

	conn.send(javaio.JoinGame {
		EntityId: 0,
		Gamemode: javaio.GamemodeCreative,
		Hardcore: false,
		Dimension: javaio.DimensionOverworld,
		ViewDistance: 1,
		ReducedDebugInfo: false,
		EnableRespawnScreen: false,
	})

	conn.send(javaio.CompassPosition {
		Location: javaio.BlockPosition { X: 0, Y: 64, Z: 0 },
	})

	conn.send(javaio.PlayerPositionAndLook {
		X: 0, Y: 64, Z: 0, Yaw: 0, Pitch: 0,
	})

	var blocksA [4096]uint32
	var blocksB [4096]uint32
	var blocksC [4096]uint32

	for i := range blocksA {
		if i < 256 {
			blocksA[i] = 33
		} else {
			blocksA[i] = 1
		}
	}

	for i := range blocksB {
		blocksB[i] = 1
	}

	for i := range blocksC {
		if i > 4095 - 256 {
			blocksC[i] = 9
		} else {
			blocksC[i] = 10
		}
	}

	for x := -3; x <= 3; x++ {
		for z := -3; z <= 3; z++ {
			conn.send(javaio.ChunkData {
				X: int32(x), Z: int32(z), IsNew: true,
				Sections: [][]uint32 { nil, blocksA[:], blocksB[:], blocksC[:] },
			})
		}
	}

	if conn.eventHandlers.OnPlayerJoin != nil {
		conn.eventHandlers.OnPlayerJoin()
	}
}

func (conn *Connection) processMovePos(data javaio.Packet_PlayerPosSb) {
	if conn.eventHandlers.OnPlayerMove == nil {
		return
	}

	conn.eventHandlers.OnPlayerMove(PlayerMove {
		HasPos: true,
		HasLook: false,
		X: data.X,
		Y: data.Y,
		Z: data.Z,
	})
}

func (conn *Connection) processMoveLook(data javaio.Packet_PlayerLookSb) {
	if conn.eventHandlers.OnPlayerMove == nil {
		return
	}

	conn.eventHandlers.OnPlayerMove(PlayerMove {
		HasPos: false,
		HasLook: true,
		Yaw: data.Yaw,
		Pitch: data.Pitch,
	})
}

func (conn *Connection) processMoveAll(data javaio.Packet_PlayerPosAndLookSb) {
	if conn.eventHandlers.OnPlayerMove == nil {
		return
	}

	conn.eventHandlers.OnPlayerMove(PlayerMove {
		HasPos: true,
		HasLook: true,
		X: data.X,
		Y: data.Y,
		Z: data.Z,
		Yaw: data.Yaw,
		Pitch: data.Pitch,
	})
}

type PlayerToSpawn struct {
	EntityId int32
	Uuid uuid.UUID
	X float64
	Y float64
	Z float64
	Yaw float32
	Pitch float32
}

func (conn *Connection) SpawnPlayer(player PlayerToSpawn) {
	conn.send(javaio.Packet_SpawnPlayer {
		EntityId: player.EntityId,
		Uuid: player.Uuid,
		X: player.X,
		Y: player.Y,
		Z: player.Z,
		Yaw: uint8(math.Round(float64(player.Yaw) / 360 * 255)),
		Pitch: uint8(math.Round(float64(player.Pitch) / 360 * 255)),
	})
}

type PlayerInfoToAdd struct {
	Uuid uuid.UUID
	Username string
	Ping int32
}

func (conn *Connection) AddPlayerInfo(players []PlayerInfoToAdd) {
	packet := javaio.PlayerInfoAdd {
		Players: make([]javaio.PlayerInfo, len(players)),
	}

	for i, player := range players {
		packet.Players[i] = javaio.PlayerInfo {
			Uuid: player.Uuid,
			Username: player.Username,
			Ping: player.Ping,
		}
	}

	conn.send(packet)
}

type EntityTranslation struct {
	EntityId int32
	DeltaX float64
	DeltaY float64
	DeltaZ float64
	Yaw float32
	Pitch float32
	OnGround bool
}

func (conn *Connection) TranslateEntity(data EntityTranslation) {
	// TODO: 8 block bound check
	conn.send(javaio.Packet_EntityTranslate {
		EntityId: data.EntityId,
		DeltaX: int16(math.Round(data.DeltaX * 4096)),
		DeltaY: int16(math.Round(data.DeltaY * 4096)),
		DeltaZ: int16(math.Round(data.DeltaZ * 4096)),
		Yaw: uint8(math.Round(float64(data.Yaw) / 360 * 255)),
		Pitch: uint8(math.Round(float64(data.Pitch) / 360 * 255)),
		OnGround: data.OnGround,
	})
}

type EntityVelocity struct {
	EntityId int32
	X float64
	Y float64
	Z float64
}

func (conn *Connection) SetEntityVelocity(data EntityVelocity) {
	// TODO: bound check
	conn.send(javaio.Packet_EntityVelocity {
		EntityId: data.EntityId,
		X: int16(math.Round(data.X * 400)),
		Y: int16(math.Round(data.Y * 400)),
		Z: int16(math.Round(data.Z * 400)),
	})
}
