package main

import (
	"embed"
	"net"
	"time"
)

//go:embed hc_alert.html
var TemplateEmbedded embed.FS

var validHCTypes = map[string]bool{"tcp": true, "cloud": true}

type Target struct {
	IP      net.IP
	Host    net.IP
	URL     string
	Port    int
	Packets int
}

type Result struct {
	Success bool
	Message string
}

type userAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TCPHCType struct {
	typ string // tcp / cloud
}

type TCPOptions struct {
	hcType  TCPHCType
	ip      net.IP
	port    int
	packets int
	URL     string
	isAuth  bool
}

type TCPChecker struct {
	Target
	TCPOptions
	Timeout time.Duration
}
