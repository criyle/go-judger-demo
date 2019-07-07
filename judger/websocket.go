package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	envJudgerToken = "JUDGER_TOKEN"
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 50 * time.Second
)

type job struct {
	ID   string `json:"id"`
	Lang string `json:"language"`
	Code string `json:"code"`
}

// Model is the database model as well as transfer model
type Model struct {
	ID      string `json:"id"`
	*Update `json:"update"`
}

// Update is the judger updates
type Update struct {
	Status string `json:"status"`
	Time   uint64 `json:"time,omitempty"`
	Memory uint64 `json:"memory,omitempty"`
	Date   uint64 `json:"date,omitempty"`
	Stdin  string `json:"stdin,omitempty"`
	Stdout string `json:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty"`
	Log    string `json:"log,omitempty"`
}

type judger struct {
	conn      *websocket.Conn // connection
	submit    chan job        // job submitted by web
	update    chan Model      // update web model
	disconnet chan *struct{}  // disconneted
}

func (j *judger) readLoop() {
	defer func() {
		j.disconnet <- nil
		j.conn.Close()
		close(j.update)
	}()

	j.conn.SetReadDeadline(time.Now().Add(pongWait))
	j.conn.SetPongHandler(func(string) error {
		j.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	var v job
	for {
		err := j.conn.ReadJSON(&v)
		if err != nil {
			log.Println("wsr: ", err)
			break
		}
		j.submit <- v
	}
}

func (j *judger) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		j.conn.Close()
	}()
	for {
		select {
		case m, ok := <-j.update:
			j.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				j.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			err := j.conn.WriteJSON(m)
			if err != nil {
				log.Println("wsw: ", err)
				return
			}

		case <-ticker.C:
			j.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := j.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func dialWS(url string) (*judger, error) {
	header := make(http.Header)
	token := os.Getenv(envJudgerToken)
	header["Authorization"] = []string{"Token", token}

	d := websocket.Dialer{}
	conn, resp, err := d.Dial(url, header)
	if err != nil {
		log.Println("dialWs: ", resp)
		return nil, err
	}

	j := &judger{
		conn:      conn,
		submit:    make(chan job),
		update:    make(chan Model),
		disconnet: make(chan *struct{}),
	}
	go j.readLoop()
	go j.writeLoop()
	return j, nil
}
