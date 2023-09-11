/*
The client package defines a client, which is a collection of torrents.
Each torrent stores the metainfo of the torrent and the peer connection
info.
*/
package client

import (
	"fmt"
	"strings"
)

type Client struct {
	Torrents map[string]torrent
}

func New() Client {
	return Client{Torrents: make(map[string]torrent)}
}

func (c *Client) AddTorrent(input string) error {
	torrent, err := newTorrent(input)
	if err != nil {
		return err
	}
	newName := torrent.metainfo.Info.Name

	// Later, to make this more efficient, do the following check immediately
	// after setting the metainfo
	for name := range c.Torrents {
		if name == newName {
			return fmt.Errorf("torrent %s already exists", name)
		}
	}

	c.Torrents[newName] = *torrent
	c.StartTorrent(newName)
	return nil
}

func (c *Client) RemoveTorrent(prefix string) error {
	for name := range c.Torrents {
		if strings.HasPrefix(name, prefix) {
			fmt.Printf("Removed torrent %s\n", name)
			c.StopTorrent(name)
			delete(c.Torrents, name)
			return nil
		}
	}

	return fmt.Errorf("no torrent matches prefix %s", prefix)
}

func (c *Client) StartTorrent(prefix string) error {
	return nil
}

func (c *Client) StopTorrent(prefix string) error {
	return nil
}

func (c Client) ShowTorrents() {
	for name := range c.Torrents {
		fmt.Printf("%s\n", name)
	}
}
