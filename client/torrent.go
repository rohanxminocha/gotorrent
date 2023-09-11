package client

import (
	"fmt"
	"io"
	"math/rand"
	"os"
)

type inputType int

const (
	// Just add support for file paths for now
	path inputType = iota
	// url
	// info hash
	// magnet link
	invalid
)

type torrent struct {
	// torrents are considered clients here, so this is the peer id
	id       []byte
	metainfo metainfo
	trackers []tracker
	peers    []peer
}

func interpretInput(input string) inputType {
	_, err := os.Open(input)
	if err == nil {
		return path
	}

	return invalid
}

func createTorrentFromFileContents(path string) (*torrent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r := io.Reader(f)
	m, err := newMetainfo(r)
	if err != nil {
		return nil, err
	}

	return &torrent{metainfo: *m}, err
}

func createId() ([]byte, error) {
	// Azureus style with arbitrary client id and version number
	base := []byte("-GG0001-")
	randSuffix := make([]byte, 12)
	_, err := rand.Read(randSuffix)
	if err != nil {
		return nil, err
	}

	id := append(base, randSuffix...)
	return id, nil
}

func newTorrent(input string) (*torrent, error) {
	var t *torrent
	var err error

	inputType := interpretInput(input)
	switch inputType {
	case path:
		t, err = createTorrentFromFileContents(input)
		if err != nil {
			return nil, err
		}
	case invalid:
		return nil, fmt.Errorf("input %s is invalid", input)
	}

	t.id, err = createId()
	if err != nil {
		return nil, err
	}

	err = t.requestPeers()
	if err != nil {
		return nil, err
	}

	err = t.handshake()
	if err != nil {
		return nil, err
	}

	return t, nil
}
