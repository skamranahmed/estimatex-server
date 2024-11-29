## 👨‍💻 EstimateX Server (`estimatex-server`)

`estimatex-server` is the backend component for [`estimatex`](https://github.com/skamranahmed/estimatex), enabling real-time story point estimation through WebSocket communication. This server handles room management, user connections, and message broadcasting to facilitate collaborative estimation sessions.

### ✨ Features
- Real-time WebSocket communication
- Room-based collaboration with admin controls
- Support for multiple concurrent estimation sessions
- Automatic room cleanup on admin disconnect
- Configurable room capacity
- Structured event system for client-server communication

### ❓ How It Works
1. Clients connect to the server using WebSocket.
2. The server facilitates communication by:
   - Handling room creation and joining
   - Receiving client events (e.g., room actions, votes), validating them, and executing the appropriate logic
   - Broadcasting events, votes and results in real-time
   - Managing session state (e.g., participants, room data)
3. Outputs final session results upon completion

### 🙌 Getting Started

#### Prerequisites
- Go 1.x or higher
- Make (optional, for using Makefile commands)

#### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/skamranahmed/estimatex-server.git
   cd estimatex-server
   ```

2. Install dependencies:
   ```bash
   make dep
   # or
   go mod tidy && go mod download
   ```

#### Running the Server
```bash
make run
# or
go run main.go
```
The server will start on port `8080`

### 🚀 API Reference

#### WebSocket Endpoint
- URL Path: `/ws`
- Protocol: `ws://` or `wss://`

#### Query Parameters
- `action`: Either `CREATE_ROOM` or `JOIN_ROOM`. It is a required parameter.
- `name`: Client's display name. It is a required parameter.
- `max_room_capacity`: Maximum number of participants. It is a required parameter when `action` is `CREATE_ROOM`. 
- `room_id`: ID of the room to join. It is a required parameter when `action` is `JOIN_ROOM`.

#### Events
The server implements a bidirectional event system:

##### Incoming Events
- `JOIN_ROOM`: When a client joins a room
- `BEGIN_VOTING`: Admin initiates voting
- `MEMBER_VOTED`: Member submits their vote
- `REVEAL_VOTES`: Admin reveals all votes

##### Outgoing Events
- `ROOM_JOIN_UPDATES`: Room membership updates
- `ROOM_CAPACITY_REACHED`: Room is full
- `BEGIN_VOTING_PROMPT`: Prompt for admin to start voting
- `ASK_FOR_VOTE`: Request for members to vote
- `VOTING_COMPLETED`: All votes received
- `REVEAL_VOTES_PROMPT`: Prompt for admin to reveal votes
- `VOTES_REVEALED`: Final vote results
- `AWAITING_ADMIN_VOTE_START`: Waiting for admin to start next vote

##### Incoming + Outgoing Events
- `CREATE_ROOM`: Room creation event

### 🧠 Project Structure
```
.
├── cmd/
│   └── app.go          # Server setup and configuration
├── internal/
│   ├── api/            # API response handling
│   ├── controller/     # WebSocket connection management
│   ├── entity/         # Domain models
│   ├── event/          # Event definitions
│   └── session/        # Session management
├── main.go             # Application entry point
├── Makefile            # Build and run commands
└── README.md           # Documentation
```

#### Available Make Commands
- `make dep`: Install dependencies
- `make run`: Start the server

### 📝 License
This project is licensed under the [MIT License](https://choosealicense.com/licenses/mit/)