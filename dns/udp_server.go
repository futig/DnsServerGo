package dns

import (
	"fmt"
	"net"
)

type UdpServer struct {
	Address  Address
	CachSize int
}

func MakeUdpServer(addr Address, cashSize int) *UdpServer {
	return &UdpServer{addr, cashSize}
}

func (s *UdpServer) Run() {
	addr := net.UDPAddr{
		Port: int(s.Address.Port),
		IP:   net.ParseIP(s.Address.Ip),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println("Ошибка при запуске UDP сервера:", err)
		return
	}
	defer conn.Close()
	fmt.Print("Udp is running\n")
	for {
		buf := make([]byte, 1024)
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Ошибка при чтении данных:", err)
			continue
		}
		handleUDPConnection(conn, clientAddr, &buf, n)
	}
}

func handleUDPConnection(conn *net.UDPConn, clientAddr *net.UDPAddr, rawRequest *[]byte, n int) {
	header, question, err := parseRequest(rawRequest, n)
	if err != nil {
		fmt.Println("Ошибка при обработке DNS запроса:", err)
	}

	fmt.Println(header)
	fmt.Println(question)

	// _, err = conn.WriteToUDP(*response, clientAddr)
	// if err != nil {
	// 	fmt.Println("Ошибка при отправке данных:", err)
	// }
}
