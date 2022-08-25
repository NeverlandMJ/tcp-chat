package server

var commands = map[string]string{
	"/help":    "shows all possible commands and thier discriptions",
	"/groups":  "lists all available groups",
	"/create":  "creates groups with the following name.\n This command accepts one argument, which is groupname",
	"/join":    "join takes groupname as an argument and adds you into that group",
	"/members": "this command lists group members. \nYou should spicify the groupname",
	"/chat":    "starts groupchat inside the given groupname. \n to use this command you have to be member of this group",
	"/replay":  "after starting the chat you can use this command to replay spicific member",
	"/exit":    "exits chat or program",
	"/pm": "starts personal messaging chat with the given username",
}