package nc

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

// RetrieveString dials tcp and returns response in string
func RetrieveString(address, port, query string, params ...string) (string, error) {
	addr := address + ":" + port
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return "", err
	}

	q := fmt.Sprintf(query, iface(params)...)

	tcpconn := conn.(*net.TCPConn)
	buf := new(bytes.Buffer)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		_, err := io.Copy(tcpconn, strings.NewReader(q))
		if err != nil {
			fmt.Println(err)
		}
		tcpconn.CloseWrite()
		wg.Done()
	}()
	go func() {
		_, err := io.Copy(buf, tcpconn)
		if err != nil {
			fmt.Println(err)
		}
		tcpconn.CloseRead()
		wg.Done()
	}()
	wg.Wait()
	return buf.String(), nil
}

func iface(list []string) []interface{} {
	vals := make([]interface{}, len(list))
	for i, v := range list {
		vals[i] = v
	}
	return vals
}
