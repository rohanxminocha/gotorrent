package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/marksamman/bencode"
)

type tracker struct {
	interval int64
	url      string
}

type safePeers struct {
	mu    sync.Mutex
	peers map[peer]bool
}

func tryToAnnounce(url string, t *torrent) ([]peer, error) {
	if strings.HasPrefix(url, "udp") || strings.HasPrefix(url, "dht") {
		return nil, errors.New("no support for UDP/DHT")
	}

	client := http.Client{
		// Arbitrary but existent timeout duration
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decodedResp, err := bencode.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	failureReason, requestFailed := decodedResp["failure reason"].(string)
	if requestFailed {
		return nil, fmt.Errorf("request failed with reason: %s", failureReason)
	}

	tracker := tracker{
		interval: decodedResp["interval"].(int64),
		url:      url,
	}
	t.trackers = append(t.trackers, tracker)

	peers := make([]peer, 0)
	for _, p := range decodedResp["peers"].([]interface{}) {
		bytes, err := json.Marshal(p)
		if err != nil {
			return nil, err
		}

		peerConnection := peerConnection{}
		err = json.Unmarshal(bytes, &peerConnection)
		if err != nil {
			return nil, err
		}

		// Instantiate peer that is, by default, choked and not interested
		peer := peer{
			choked:     true,
			available:  false,
			connection: peerConnection,
			interested: false,
		}
		peers = append(peers, peer)
	}

	return peers, nil
}

func (t *torrent) requestPeers() error {
	announceUrls := make([]string, 0)
	announceUrls = append(announceUrls, t.metainfo.Announce)
	for _, url := range t.metainfo.AnnounceList {
		announceUrls = append(announceUrls, url...)
	}

	safePeers := safePeers{peers: make(map[peer]bool)}
	var wg sync.WaitGroup
	for _, u := range announceUrls {
		wg.Add(1)
		requestUrl := fmt.Sprintf(
			"%s?info_hash=%s&peer_id=%s&uploaded=%d&downloaded=%d&left=%d",
			u,
			url.QueryEscape(string(t.metainfo.infoHash[:])),
			url.QueryEscape(string(t.id)),
			0,
			0,
			t.metainfo.Info.Length,
		)

		// Asynchronously GET peers from all announce URLs
		go func() {
			defer wg.Done()
			peers, err := tryToAnnounce(requestUrl, t)
			if err != nil {
				log.Println(err)
				return
			}

			// Wrap peer writing with mutex lock/unlock so that only one
			// goroutine can access it at a time
			// This also filters out duplicate peers
			safePeers.mu.Lock()
			for _, p := range peers {
				if _, peerExists := safePeers.peers[p]; !peerExists {
					safePeers.peers[p] = true
					t.peers = append(t.peers, p)
				}
			}
			safePeers.mu.Unlock()
		}()
	}

	wg.Wait()
	if len(safePeers.peers) == 0 {
		return errors.New("no peers found after announcing to all urls")
	}

	return nil
}
