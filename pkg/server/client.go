package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type Client struct {
	username string
	conn     net.Conn
	group    *Group
}

func NewClient(conn net.Conn) (Client, error) {
	c := Client{
		username: "",
		conn:     conn,
	}

	if err := c.getUsername(); err != nil {
		return Client{}, err
	}

	return c, nil
}

func (c *Client) getUsername() error {
	if err := c.message("Enter your name: ", ""); err != nil {
		return err
	}

	input, err := c.readInput()
	if err != nil {
		return err
	}

	c.username = input

	return nil
}

func (c *Client) readInput() (string, error) {
	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		log.Printf("error while reading from conn: %v", err)
		return "", err
	}
	if errors.Is(err, io.EOF) {
		msg = "/exit"
	}

	return strings.Trim(msg, "\n"), nil
}

func (c *Client) message(msg string, groupname string) error {
	if _, err := c.conn.Write([]byte(fmt.Sprintf("%s> %s\n", groupname, msg))); err != nil {
		return err
	}

	return nil
}

func (c *Client) replay(user Client, gr string) string {
	msg, err := c.readInput()
	if err != nil {
		log.Printf("failed to replay %s: %v", user.username, err)
		return ""
	}
	str := fmt.Sprintf(fmt.Sprintf("%s> %s replayed to you: %s\n", gr, c.username, msg))
	if _, err := user.conn.Write([]byte(str)); err != nil {
		log.Printf("failed to replay %s: %v", user.username, err)
		return ""
	}
	return msg
}

func (c *Client) chatting(user Client, gr string, msg string) string {
	str := fmt.Sprintf(fmt.Sprintf("%s> %s wrote to you: %s\n", gr, c.username, msg))
	if _, err := user.conn.Write([]byte(str)); err != nil {
		log.Printf("failed to replay %s: %v", user.username, err)
		return ""
	}
	return msg
}
