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

type netAccess interface {
	access()
	handleConn(interface{})
	deferDeal(interface{})
}

type httpWay string
type tcpWay string
type udpWay string

//三种类型实现netAccess的所有方法
/*HTTP*/
func (httpDo httpWay) access() {
	for {
		httpHandler := func(w http.ResponseWriter, req *http.Request) {
			res := dealReadData()
			io.WriteString(w, res)
		}
		http.HandleFunc("/", httpHandler)
		log.Println(http.ListenAndServe(":8000", nil).Error())
	}
}
func (httpDo httpWay) handleConn(interface{}) {
}
func (httpDo httpWay) deferDeal(interface{}) {
}

/*TCP*/
func (tcpDo tcpWay) access() {
	listener, err := net.Listen("tcp", "localhost:8001")
	defer tcpDo.deferDeal(listener)
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

		go tcpDo.handleConn(conn)
	}
}
func (tcpDo tcpWay) handleConn(conn interface{}) {
	var c net.Conn = conn.(net.Conn)
	defer c.Close()
	res := dealReadData()
	io.WriteString(c, res)
}
func (tcpDo tcpWay) deferDeal(conn interface{}) {
	var c net.Listener = conn.(net.Listener)
	waitGroup.Done()
	c.Close()
}

/*UDP*/
func (udpDo udpWay) access() {
	defer waitGroup.Done()
	for {
		conn, err := net.ListenPacket("udp", "localhost:8002")
		//go handleUdpConn(conn)
		udpDo.handleConn(conn)
		if err != nil {
			log.Printf(err.Error())
		}

	}
}
func (udpDo udpWay) handleConn(conn interface{}) {
	var c net.PacketConn = conn.(net.PacketConn)
	buffer := make([]byte, 1024)
	_, remoteAddr, err := c.ReadFrom(buffer)
	if err != nil {
		log.Panicln("ReadFrom err", err.Error())
	}
	go handleUdpConn1(udpDo, c, remoteAddr)
}
func (udpDo udpWay) deferDeal(conn interface{}) {
	var c net.PacketConn = conn.(net.PacketConn)
	c.Close()
}

func handleUdpConn1(udpDo udpWay, conn net.PacketConn, remoteAddr net.Addr) {
	defer udpDo.deferDeal(conn)
	res := dealReadData()
	conn.WriteTo([]byte(res), remoteAddr)
}

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
func main() {
	var httpDo httpWay = Http
	var tcpDo tcpWay = Tcp
	var udpDo udpWay = Udp

	accessWay := [...]netAccess{httpDo, tcpDo, udpDo}
	for _, Type := range accessWay {
		go Type.access()
	}
	//设置信号量，等待返回goroutine返回才结束
	waitGroup.Add(2)
	waitGroup.Wait()
	fmt.Println("所有groutine已经退出")
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
