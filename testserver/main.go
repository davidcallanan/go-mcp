package main

import "fmt"
import "net"
import "github.com/davidcallanan/gomcp/javaserver"
import "github.com/google/uuid"

type Player struct {
	conn *javaserver.Connection
	uuid uuid.UUID
	username string
	playerEids map[uuid.UUID]int32
	theNextEid int32
	x float64
	y float64
	z float64
	yaw float32
	pitch float32
}

func main() {
	const maxPlayers = 20
	const version = "1.14-1.15"
	players := make([]*Player, 0, maxPlayers)

	listener, err := net.Listen("tcp4", "localhost:25565")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Test server is now listening...")

	for {
		connection, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		fmt.Println("Accepted a connection!")
		player := &Player{}

		player.conn = javaserver.NewConnection(connection, func() {
			connection.Close()
		}, javaserver.EventHandlers {
			OnStatusRequestV1: func() javaserver.StatusResponseV1 {
				return javaserver.StatusResponseV1 {
					Description: "Hello, World!",
					MaxPlayers: maxPlayers,
					OnlinePlayers: len(players),
				}
			},
		
			OnStatusRequestV2: func() javaserver.StatusResponseV2 {
				return javaserver.StatusResponseV2 {
					IsClientSupported: false,
					Version: version,
					Description: "§e§lHello, World!",
					MaxPlayers: maxPlayers,
					OnlinePlayers: len(players),
				}
			},
		
			OnStatusRequestV3: func() javaserver.StatusResponseV3 {
				return javaserver.StatusResponseV3 {
					IsClientSupported: true,
					Version: version,
					Description: "§e§lHello, World!\n§r§aWelcome to this amazing server",
					MaxPlayers: maxPlayers,
					OnlinePlayers: len(players),
					PlayerSample: []string {
						"§aThis is",
						"§cthe most",
						"§8amazing thing",
						"§9§lever!!!",
					},
				}
			},
		
			OnPlayerJoinRequest: func(data javaserver.PlayerJoinRequest) javaserver.PlayerJoinResponse {
				fmt.Printf("Player %s has requested to join the game.\n", data.ClientsideUsername)
				
				if len(players) >= maxPlayers {
					fmt.Println("Player has been silently denied to join due to player limit.")
					return javaserver.PlayerJoinResponse {
						PreventResponse: true,
					}
				}
				
				players = append(players, player)
				player.uuid = uuid.New()
				player.username = data.ClientsideUsername
				return javaserver.PlayerJoinResponse {
					Uuid: player.uuid,
				}
			},
		
			OnPlayerJoin: func() {
				fmt.Println("Player of whom I forget their username has joined the game.")

				player.x = 0
				player.y = 64
				player.z = 0
				player.theNextEid = 0
				player.playerEids = make(map[uuid.UUID]int32)

				player.conn.AddPlayerInfo([]javaserver.PlayerInfoToAdd {
					{ Uuid: uuid.New(), Username: "JohnDoe", Ping: 0 },
					{ Uuid: uuid.New(), Username: "CatsEyebrows", Ping: 5 },
					{ Uuid: uuid.New(), Username: "ElepantNostrel23", Ping: 500 },
				})

				for _, p := range players {
					// Create self eid for other players
					p.playerEids[player.uuid] = p.nextEid()

					// Add self to tab list for other players
					p.conn.AddPlayerInfo([]javaserver.PlayerInfoToAdd {
						{ Uuid: player.uuid, Username: player.username, Ping: 0 },
					})
					
					// Add other players to self tab list
					player.conn.AddPlayerInfo([]javaserver.PlayerInfoToAdd {
						{ Uuid: p.uuid, Username: player.username, Ping: 0 },
					})

					if p.uuid != player.uuid {
						// Create other player eids for self
						player.playerEids[p.uuid] = player.nextEid()

						// Spawn self for already connected players
						p.conn.SpawnPlayer(javaserver.PlayerToSpawn {
							EntityId: p.playerEids[player.uuid],
							Uuid: player.uuid,
							X: player.x,
							Y: player.y,
							Z: player.z,
							Yaw: 0,
							Pitch: 0,
						})

						// Spawn already connected players for self
						player.conn.SpawnPlayer(javaserver.PlayerToSpawn {
							EntityId: player.playerEids[p.uuid],
							Uuid: p.uuid,
							X: p.x,
							Y: p.y,
							Z: p.z,
							Yaw: 0,
							Pitch: 0,
						})
					}
				}
			},
			OnPlayerMove: func(data javaserver.PlayerMove) {
				prevX := player.x
				prevY := player.y
				prevZ := player.z
				
				if data.HasPos {
					player.x = data.X
					player.y = data.Y
					player.z = data.Z
				}

				if data.HasLook {
					player.yaw = data.Yaw
					player.pitch = data.Pitch
				}

				for _, p := range players {
					if p.uuid == player.uuid {
						return
					}
					
					p.conn.TranslateEntity(javaserver.EntityTranslation {
						EntityId: p.playerEids[player.uuid],
						DeltaX: player.x - prevX,
						DeltaY: player.y - prevY,
						DeltaZ: player.z - prevZ,
						Yaw: player.yaw,
						Pitch: player.pitch,
					})
				}
			},
		})
	}
}

func (player *Player) nextEid() int32 {
	player.theNextEid++
	return player.theNextEid
}
