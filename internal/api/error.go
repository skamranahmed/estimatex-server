package api

import (
	"github.com/gorilla/websocket"
)

// SendErrorResponse: sends an error message via websocket and then closes the websocket connection
func SendErrorResponse(wsConnection *websocket.Conn, errorDescription string) {
	wsConnection.WriteMessage(websocket.TextMessage, []byte(errorDescription))
	wsConnection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Server closing connection"))
	wsConnection.Close()
}
