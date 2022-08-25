package server

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Server struct {
	groups  map[string]*Group
	clients *[]Client
}

func New() Server {
	return Server{
		groups:  make(map[string]*Group),
		clients: &[]Client{},
	}
}

func (s *Server) Run(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Println("Waiting for connections")
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("error while accepting conn: %v\n", err)
			continue
		}

		log.Println("Accepted new connection")

		go s.handleConnection(conn)
	}
}



func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	c, err := NewClient(conn)
	if err != nil {
		log.Printf("failed to initialize new client: %v\n", err)
		return
	}
	log.Printf("%s joined our server!\n", c.username)
	*s.clients = append(*s.clients, c)

	if err := c.message(fmt.Sprintf("Welcome %s!", c.username), ""); err != nil {
		log.Printf("error while messaging client: %v, %v", err, c)
	}

	if err := c.message("To see list of commands type /help", ""); err != nil {
		log.Printf("error while messaging client: %v, %v", err, c)
	}

	for {
		in, err := c.readInput()
		if err != nil {
			log.Printf("error while reading client input: %v\n", err)
			continue
		}

		ins := strings.Fields(in)
		if len(ins) < 1 {
			log.Printf("empty input")
			continue
		}
		switch ins[0] {
		case "/groups":
			s.listGroups(c)
		case "/create":
			if len(ins) != 2 {
				c.message("ERR: /create command requires one argument - group's name", "")
				continue
			}
			s.createGroup(&c, ins[1])
		case "/join":
			if len(ins) != 2 {
				c.message("ERR: /join command requires one argument - group's name", "")
				continue
			}
			s.joinGroup(&c, ins[1])
		case "/members":
			if len(ins) != 2 {
				c.message("ERR: /members command requires one argument - group's name", "")
				continue
			}
			s.listGroupMembers(c, ins[1])
		case "/chat":
			if len(ins) != 2 {
				c.message("ERR: /chat command requires one argument - group's name", "")
				continue
			}
			s.groupChat(c, ins[1])
		case "/pm":
			if len(ins) != 2 {
				c.message("ERR: /pm command requires one argument - user's name", "")
				continue
			}
			s.personalChat(c, ins[1])
		case "/help":
			s.listCommands(c)
		case "/exit":
			log.Printf("%s left server\n", c.username)
			return
		default:
			c.message("ERR: unrecognized command!", "")
		}
	}
}

func (s *Server) listCommands(c Client) {
	msg := strings.Builder{}
	msg.WriteString("Availabe commands:\n")
	for k, v := range commands {
		msg.WriteString(fmt.Sprintf("%s - %s\n\n", k, v))
	}

	if err := c.message(msg.String(), ""); err != nil {
		log.Printf("error while messaging client: %v - %v", err, c)
	}
}

func (s *Server) listGroups(c Client) {
	msg := strings.Builder{}
	msg.WriteString("Availabe groups:\n")

	for k := range s.groups {
		msg.WriteString(fmt.Sprintf("%s\n", k))
	}

	if err := c.message(msg.String(), ""); err != nil {
		log.Printf("error while messaging client: %v - %v", err, c)
	}
}

func (s *Server) createGroup(c *Client, groupName string) {
	if _, ok := s.groups[groupName]; ok {
		c.message("ERR: this group already exists!", "")
		return
	}

	s.groups[groupName] = NewGroup(c)
	c.group = s.groups[groupName]
	c.message(fmt.Sprintf("Successfully created and joined group - %s", groupName), "")
}

func (s *Server) joinGroup(c *Client, groupName string) {
	group, ok := s.groups[groupName]
	if !ok {
		c.message("ERR: this group does not exist, if you want, you can create one with /create command!", "")
		return
	}

	c.group = group
	group.clients = append(group.clients, c)
	c.message(fmt.Sprintf("Successfully joined group - %s", groupName), "")
}

func (s *Server) listGroupMembers(c Client, groupname string) {
	if _, ok := s.groups[groupname]; !ok {
		c.message("ERR: this group does not exist!", "")
		return
	}
	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("%s has following members: ", groupname))

	for _, v := range s.groups[groupname].clients {
		msg.WriteString(fmt.Sprintf("%s\n", v.username))
	}

	if err := c.message(msg.String(), ""); err != nil {
		log.Printf("error while listing group members: %v - %v", err, c)
	}

}

func (s *Server) groupChat(c Client, groupname string) {
	if _, ok := s.groups[groupname]; !ok {
		c.message("ERR: this group does not exist!", "")
		return
	}

	isUserExist := false

	for _, v := range s.groups[groupname].clients {
		if v.username == c.username {
			isUserExist = true
		}
	}

	if !isUserExist {
		c.message("ERR: you are not member of this group. Please join first to chat", "")
		return
	}

	if len(s.groups[groupname].clients) < 2 {
		c.message("ERR: this group does not have enough members to chat. Please wait for others to join", "")
		return
	}

	c.message("Succesfully started chatting!", groupname)
	log.Printf("%s has started chatting in %s!", c.username, groupname)
	for {
		msg, err := c.readInput()
		if msg == "/exit" {
			c.message("group chat ended", groupname)
			fmt.Printf("%s left the group chat\n", c.username)
			str := fmt.Sprintf("%s left the group chat\n", c.username)

			for _, v := range s.groups[groupname].clients{
				if *v != c {
					v.message(str, groupname)
				}
			}
			return
		}

		msgs := strings.Fields(msg)
		if len(msgs) < 1 {
			log.Printf("empty input")
			continue
		}
		if msgs[0] == "/replay" {
			if len(msgs) != 2 {
				c.message("ERR: /reply command requires one argument - user's name", "")
				continue
			} else {
				isUserExist := false
				user := Client{}
				for _, v := range s.groups[groupname].clients {
					if msgs[1] == v.username {
						isUserExist = true
						user = *v
					}
				}
				if isUserExist {
					str := c.replay(user, groupname)
					fmt.Printf("FROM %s TO %s> %s\n", c.username, user.username, str)

					toUsers := fmt.Sprintf("FROM %s TO %s> %s\n", c.username, user.username, str)

					for _, v := range s.groups[groupname].clients{
						if *v != user {
							v.message(toUsers, groupname)
						}
					}

				} else {
					c.message("ERR: user with this name does not exist", groupname)
				}
			}
		} else {

			if err != nil {
				log.Println("ERR: error while reading input")
				c.message("ERR: error while reading your input", groupname)
				continue
			}

			fmt.Printf("%s> ", c.username)
			fmt.Println(msg)
			
			str := fmt.Sprintf("%s> %s", c.username, msg)

			for _, v := range s.groups[groupname].clients{
				if *v != c {
					v.message(str, groupname)
				}
			}
		}

	}
}

func (s *Server) personalChat(c Client, username string) {
	isUserExist := false
	user := Client{}
	for _, v := range *s.clients {
		if username == v.username {
			isUserExist = true
			user = v
		}
	}
	if isUserExist {
		user.message(fmt.Sprintf("%s wants to chat with you.", c.username), "PM")
	}else {
		c.message("ERR: user with this name does not exist", "server")
		return
	}

	for {
		msg, err := c.readInput()
		if err != nil {
			log.Println("ERR: error while reading input")
			c.message("ERR: error while reading your input", "PM")
			continue
		}
		if msg == "/exit" {
			user.message(fmt.Sprintf("%s left the personal chat", c.username), "PM")
			return
		}
		
		c.chatting(user, c.username, msg)
		
	}

}
