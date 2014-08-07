package main

import (
  "net"
  "io"
  "github.com/gorilla/websocket"
  "flag"
  "log"
  "net/http"
)

var addr = flag.String("addr", ":8080", "http service address")
var remote = flag.String("remote", "localhost:80", "target service address")

func serveWs(w http.ResponseWriter, r *http.Request) {
  if r.Method != "GET" {
    http.Error(w, "Method not allowed", 405)
    return
  }
  conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
  if _, ok := err.(websocket.HandshakeError); ok {
    http.Error(w, "Not a websocket handshake", 400)
    return
  } else if err != nil {
    log.Println(err)
    return
  }

  remote, err := net.Dial("tcp", *remote)
  if err != nil {
    log.Println(err)
    return
  }
  defer remote.Close()

  pump(conn, remote)
}

func pump(conn *websocket.Conn, remote net.Conn) {
  for {
    messageType, r, err := conn.NextReader()
    if err != nil {
      log.Println("Getting Reader:", err)
      return
    }
    w, err := conn.NextWriter(messageType)
    if err != nil {
      log.Println("Getting Writer:", err)
      return
    }
    go func() {
      if _, err := io.Copy(remote, r); err != nil {
        log.Println("Outgoing:", err)
        return
      }
    }()
    if _, err := io.Copy(w, remote); err != nil {
      log.Println("Incoming:", err)
      return
    }
    if err := w.Close(); err != nil {
      log.Println("Close:", err)
      return
    }
  }
}

func main() {
  flag.Parse()
  http.HandleFunc("/mqtt", serveWs)
  err := http.ListenAndServe(*addr, nil)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}
