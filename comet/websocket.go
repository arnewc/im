package main

import (
	"net/http"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"time"
	"encoding/json"
	"im/libs/proto"
	"im/libs/define"
)




func InitWebsocket(bind string) (err error) {

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(DefaultServer, w, r)
	})


	err = http.ListenAndServe(bind, nil)
	return err

}

// serveWs handles websocket requests from the peer.
func serveWs(server *Server, w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  DefaultServer.Options.ReadBufferSize,
		WriteBufferSize: DefaultServer.Options.WriteBufferSize,
	}

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error(err)
		return
	}


	go server.writePump(conn)
	go server.readPump(conn)
}



func (s *Server) readPump(conn *websocket.Conn) {
	defer func() {
		conn.Close()
	}()

	conn.SetReadLimit(s.Options.MaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(s.Options.PongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(s.Options.PongWait));
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway,websocket.CloseAbnormalClosure) {
				log.Errorf("readPump ReadMessage err:%v", err)
			}
		}
		log.Infof("message :%v", message)

		var p = &proto.Proto{Ver: 0, Operation: define.OP_SEND, Body: message}
		log.Debugf("message: %s", message)

		s.Buckets[0].chs["0"].signal <- p

	}
}

func (s *Server) writePump(conn *websocket.Conn) {
	ticker := time.NewTicker(s.Options.PingPeriod)
	log.Printf("ticker :%v", ticker)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()
	for {
		select {
		case message, ok := s.Buckets[0].chs["0"].signal:
			conn.SetWriteDeadline(time.Now().Add(s.Options.WriteWait))
			if !ok {
				// The hub closed the channel.
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			log.Printf("TextMessage :%v", websocket.TextMessage)
			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			log.Printf("message :%v", message)
			// w.Write(message)




			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(s.Options.WriteWait))
			log.Printf("websocket.PingMessage :%v", websocket.PingMessage)
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}


// func (server *Server) run() {
// 	for{
// 		select {
// 		// case server.Buckets
// 		}
// 	}
// }


