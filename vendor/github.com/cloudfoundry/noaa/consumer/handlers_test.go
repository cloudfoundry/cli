package consumer_test

import (
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type HttpHandler struct {
	Messages <-chan []byte
}

func NewHttpHandler(m <-chan []byte) *HttpHandler {
	return &HttpHandler{Messages: m}
}

func (h *HttpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	mp := multipart.NewWriter(rw)
	defer mp.Close()

	rw.Header().Set("Content-Type", `multipart/x-protobuf; boundary=`+mp.Boundary())

	for message := range h.Messages {
		partWriter, err := mp.CreatePart(nil)
		if err != nil {
			return
		}

		partWriter.Write(message)
	}
}

type websocketHandler struct {
	messages  <-chan []byte
	keepAlive time.Duration
}

func NewWebsocketHandler(m <-chan []byte, keepAlive time.Duration) *websocketHandler {
	return &websocketHandler{messages: m, keepAlive: keepAlive}
}

func (h *websocketHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(*http.Request) bool { return true },
	}

	ws, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Printf("websocket handler: Not a websocket handshake: %s", err)
		return
	}
	defer ws.Close()

	closeCode, closeMessage := h.runWebsocketUntilClosed(ws)
	ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(closeCode, closeMessage), time.Time{})
}

func (h *websocketHandler) runWebsocketUntilClosed(ws *websocket.Conn) (closeCode int, closeMessage string) {
	keepAliveExpired := make(chan struct{})
	clientWentAway := make(chan struct{})

	// TODO: remove this loop (but keep ws.ReadMessage()) once we retire support in the cli for old style keep alives
	go func() {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				close(clientWentAway)
				return
			}
		}
	}()

	go func() {
		NewKeepAlive(ws, h.keepAlive).Run()
		close(keepAliveExpired)
	}()

	closeCode = websocket.CloseNormalClosure
	closeMessage = ""
	for {
		select {
		case <-clientWentAway:
			return
		case <-keepAliveExpired:
			closeCode = websocket.ClosePolicyViolation
			closeMessage = "Client did not respond to ping before keep-alive timeout expired."
			return
		case message, ok := <-h.messages:
			if !ok {
				return
			}
			err := ws.WriteMessage(websocket.BinaryMessage, message)
			if err != nil {
				return
			}
		}
	}
}

type KeepAlive struct {
	conn              *websocket.Conn
	pongChan          chan struct{}
	keepAliveInterval time.Duration
}

func NewKeepAlive(conn *websocket.Conn, keepAliveInterval time.Duration) *KeepAlive {
	return &KeepAlive{
		conn:              conn,
		pongChan:          make(chan struct{}, 1),
		keepAliveInterval: keepAliveInterval,
	}
}

func (k *KeepAlive) Run() {
	k.conn.SetPongHandler(k.pongHandler)
	defer k.conn.SetPongHandler(nil)

	timeout := time.NewTimer(k.keepAliveInterval)
	for {
		err := k.conn.WriteControl(websocket.PingMessage, nil, time.Time{})
		if err != nil {
			return
		}
		timeout.Reset(k.keepAliveInterval)
		select {
		case <-k.pongChan:
			time.Sleep(k.keepAliveInterval / 2)
			continue
		case <-timeout.C:
			return
		}
	}
}

func (k *KeepAlive) pongHandler(string) error {
	select {
	case k.pongChan <- struct{}{}:
	default:
	}
	return nil
}
