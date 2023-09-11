package client

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"io"
	"reflect"

	"github.com/marksamman/bencode"
)

// Some of the following fields don't make sense, but the goal with these
// structs is to store data directly from the decoded torrent file stream.
// Thus, this represents the structure of that stream.
type file struct {
	Length int64    `json:"length"`
	Path   []string `json:"path"`
}

type info struct {
	Files            []file `json:"files"`
	hasMultipleFiles bool
	Length           int64  `json:"length"`
	Name             string `json:"name"`
	PieceLength      int64  `json:"piece length"`
	Pieces           string `json:"pieces"`
}

type metainfo struct {
	Announce     string     `json:"announce"`
	AnnounceList [][]string `json:"announce-list"`
	Comment      string     `json:"comment"`
	CreatedBy    string     `json:"created by"`
	CreationDate int64      `json:"creation date"`
	Info         info       `json:"info"`
	infoHash     [20]byte
}

func (m metainfo) checkFieldsPostUnmarshal() error {
	if m.Announce == "" && m.AnnounceList == nil {
		return errors.New("no url in metainfo to announce to")
	} else if reflect.DeepEqual(m.Info, info{}) {
		return errors.New("no info in metainfo")
	} else if m.Info.Name == "" {
		return errors.New("no name field in metainfo info")
	} else if m.Info.PieceLength == 0 || m.Info.Pieces == "" {
		return errors.New("no piece info in metainfo info")
	} else if len(m.Info.Files) == 0 && m.Info.Length == 0 {
		return errors.New("neither files nor length exists in metainfo info")
	}

	return nil
}

func (m *metainfo) setRemainingFields(d map[string]interface{}) {
	m.infoHash = sha1.Sum(bencode.Encode(
		d["info"].(map[string]interface{}),
	))
	m.Info.hasMultipleFiles = len(m.Info.Files) != 0
}

func newMetainfo(r io.Reader) (*metainfo, error) {
	decodedStream, err := bencode.Decode(r)
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(decodedStream)
	if err != nil {
		return nil, err
	}

	m := metainfo{}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}

	err = m.checkFieldsPostUnmarshal()
	if err != nil {
		return nil, err
	}
	m.setRemainingFields(decodedStream)

	return &m, nil
}
