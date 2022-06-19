package utils

import (
	"net"
	"strconv"
	"strings"
)

func GetProt() int64 {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer l.Close()
	list := strings.Split(l.Addr().String(), ":")
	i, err := strconv.ParseInt(list[1], 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}
