package gosocketio

import (
	"net/url"
	"strconv"

	"github.com/zhenwenk/golang-socketio/protocol"
	"github.com/zhenwenk/golang-socketio/transport"
)

const (
	webSocketProtocol       = "ws://"
	webSocketSecureProtocol = "wss://"
	socketioUrl             = "/socket.io/?EIO=3&transport=websocket"
)

/**
Socket.io client representation
*/
type Client struct {
	methods
	Channel
}

/**
Get ws/wss url by host and port
*/
func GetUrl(host string, port int, secure bool) string {
	var prefix string
	if secure {
		prefix = webSocketSecureProtocol
	} else {
		prefix = webSocketProtocol
	}
	return prefix + host + ":" + strconv.Itoa(port) + socketioUrl
}

/**
Get ws/wss url by host and port
*/
func GetSocketUrl(urlString string) (string, string, error) {
	urlInfo, err := url.Parse(urlString)
	if err != nil {
		return "", "", err
	}

	var port int = 80
	var secure bool = false
	if urlInfo.Scheme == "https" {
		port = 443
		secure = true
	}

	socketUrl := GetUrl(urlInfo.Host, port, secure)
	if urlInfo.RawQuery != "" {
		socketUrl += "&" + urlInfo.RawQuery
	}
	var namesapce string = ""
	if urlInfo.Path != "" {
		namesapce = urlInfo.Path
	}

	return socketUrl, namesapce, nil
}

/**
connect to host and initialise socket.io protocol

The correct ws protocol url example:
ws://myserver.com/socket.io/?EIO=3&transport=websocket

You can use GetUrlByHost for generating correct url
*/
func Dial(originUrl string, tr transport.Transport) (*Client, error) {

	socketUrl, namesapce, err := GetSocketUrl(originUrl)
	if err != nil {
		return nil, err
	}

	c := &Client{}
	c.initChannel(namesapce)
	c.initMethods()

	c.conn, err = tr.Connect(socketUrl)
	if err != nil {
		return nil, err
	}

	go inLoop(&c.Channel, &c.methods)
	go outLoop(&c.Channel, &c.methods)
	go pinger(&c.Channel)

	return c, nil
}

/**
Close client connection
*/
func (c *Client) Close() {
	closeChannel(&c.Channel, &c.methods)
}

/**
Send Empty Message for client connection
*/
func (c *Client) SendOpenSequence() {
	c.out <- protocol.MustEncode(&protocol.Message{Type: protocol.MessageTypeEmpty}, c.Channel.namespace)
}
