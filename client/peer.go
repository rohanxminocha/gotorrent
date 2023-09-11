package client

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type peerConnection struct {
	Id   string `json:"peer id"`
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

type peer struct {
	choked     bool
	available  bool
	connection peerConnection
	interested bool
}

func (t *torrent) handshake() error {
	clientHandshake := fmt.Sprintf(
		"%s%s%s",
		"\023BitTorrent protocol00000000",
		string(t.metainfo.infoHash[:]),
		t.id,
	)

	var wg sync.WaitGroup
	for i := 0; i < len(t.peers); i++ {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			ip := t.peers[idx].connection.Ip
			port := t.peers[idx].connection.Port
			dialer := net.Dialer{
				Timeout: 5 * time.Second,
			}
			conn, err := dialer.Dial(
				"tcp",
				fmt.Sprintf("%s:%d", ip, port),
			)
			if err != nil {
				return
			}
			defer conn.Close()

			// Send the handshake message to the conn and wait on and read the
			// response. Wait time is arbitrary, haven't read any literature
			// on an optimal time yet.
			conn.SetDeadline(time.Now().Add(5 * time.Second))
			conn.Write([]byte(clientHandshake))

			// Can't assume that the handshake starts with
			// "19Bittorrent protocol". Read the length first and then read
			// the rest based on that length.
			pstrlenBytes := make([]byte, 1)
			_, err = conn.Read(pstrlenBytes)
			if err != nil {
				return
			}
			pstrlen := int(pstrlenBytes[0])
			if pstrlen == 0 {
				return
			}

			peerHandshake := make([]byte, 48+pstrlen)
			n, err := conn.Read(peerHandshake)
			if err != nil || n == 0 {
				return
			}

			// Return if the info hash coming in isn't the same as the one
			// we're requesting
			l := len(peerHandshake)
			if string(t.metainfo.infoHash[:]) != string(peerHandshake[l-40:l-20]) {
				return
			}

			// Since we could complete the handshake, set the peer ID and
			// connected fields
			t.peers[idx].connection.Id = string(peerHandshake[l-20 : l-1])
			t.peers[idx].available = true
		}(i)
	}

	wg.Wait()
	return nil
}
