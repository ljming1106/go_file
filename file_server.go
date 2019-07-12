package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
)

const Http = "http"
const Tcp = "tcp"
const Udp = "udp"

var waitGroup = sync.WaitGroup{}

func main() {
	/*三种访问方式：
	 1、http
	  1）创建goroutine，监听端口
	  2）当有连接进来时，新建goroutine处理内容

	 2、tcp
 	  1）创建goroutine，监听端口
	  2）当有连接进来时，处理内容

	 3、udp
	tcp和udp的实现
	https://www.jianshu.com/p/dec62eff73ba
	*/
	accessWay := [...]string{Http, Tcp, Udp}
	for _, Type := range accessWay {
		switch Type {
		case Http:
			go httpAccess()
		case Tcp:
			go tcpAccess()
		case Udp:
			go udpAccess()
		}
	}
	//设置信号量，等待返回goroutine返回才结束
	waitGroup.Add(2)
	waitGroup.Wait()
	fmt.Println("所有groutine已经退出")
}

func httpAccess() {
	for {
		httpHandler := func(w http.ResponseWriter, req *http.Request) {
			res := dealReadData()
			io.WriteString(w, res)
		}
		http.HandleFunc("/", httpHandler)
		log.Println(http.ListenAndServe(":8000", nil).Error())
	}
}

func tcpAccess() {
	listener, err := net.Listen("tcp", "localhost:8001")
	defer deferTcpDeal(listener)
	if err != nil {
		log.Println(err.Error())
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf(err.Error())
			continue
		}

		go handleTcpConn(conn)
	}
}

func udpAccess() {
	defer waitGroup.Done()
	for {
		conn, err := net.ListenPacket("udp", "localhost:8002")
		//go handleUdpConn(conn)
		handleUdpConn(conn)
		if err != nil {
			log.Printf(err.Error())
		}

	}
}

func deferTcpDeal(c net.Listener) {
	waitGroup.Done()
	c.Close()
}

func deferUdpDeal(c net.PacketConn) {
	c.Close()
}

func handleTcpConn(c net.Conn) {
	defer c.Close()
	res := dealReadData()
	io.WriteString(c, res)
}

func handleUdpConn(conn net.PacketConn) {
	buffer := make([]byte, 1024)
	_, remoteAddr, err := conn.ReadFrom(buffer)
	if err != nil {
		log.Panicln("ReadFrom err", err.Error())
	}
	go handleUdpConn1(conn, remoteAddr)
}

func handleUdpConn1(conn net.PacketConn,remoteAddr net.Addr) {
	defer deferUdpDeal(conn)
	res := dealReadData()
	conn.WriteTo([]byte(res), remoteAddr)
}

func dealReadData() string {
	data, err := ioutil.ReadFile("example.txt")
	var res string
	switch err {
	case nil:
		res = string(data)
	default:
		res = err.Error()
	}
	return res
}
