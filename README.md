## EstimateX Server (`estimatex-server`)

`estimatex-server` is the backend component for [`estimatex`](https://github.com/skamranahmed/estimatex), enabling real-time story point estimation through WebSocket communication. This server handles room management, user connections, and message broadcasting to facilitate collaborative estimation sessions.

### How It Works
1. Clients connect to the server using WebSocket.
2. The server facilitates communication by:
   - Handling room creation and joining.
   - Receiving client events (e.g. - room actions, votes), validating them, and executing the appropriate logic.
   - Broadcasting events, votes and results in real-time.
   - Managing session state (e.g. - participants, room data).
3. Outputs final session results upon completion.

### License
This project is licensed under the [MIT License](https://choosealicense.com/licenses/mit/)