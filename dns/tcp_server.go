package dns

// import (
// 	"fmt"
// 	"net"
// )

// type TcpServer struct {
// 	Address  Address
// 	CachSize int
// }

// func MakeTcpServer(addr Address, cashSize int) *TcpServer {
// 	return &TcpServer{addr, cashSize}
// }

// func (s *TcpServer) Run() {
// 	addr := fmt.Sprintf("%s:%d", s.Address.Ip, s.Address.Port)
// 	server, err := net.Listen("tcp", addr)
// 	if err != nil {
// 		fmt.Println("Ошибка при запуске TCP сервера:", err)
// 		return
// 	}
// 	defer server.Close()

// 	fmt.Print("Tcp is running\n")

// 	for {
// 		conn, err := server.Accept()
// 		if err != nil {
// 			fmt.Println("Ошибка при принятии соединения по TCP:", err)
// 			continue
// 		}

// 		go handleTCPConnection(conn)
// 	}
// }

// func handleTCPConnection(conn net.Conn) {
// 	defer conn.Close()

// 	buf := make([]byte, 1024)
// 	_, err := conn.Read(buf)
// 	if err != nil {
// 		fmt.Println("Ошибка при чтении данных:", err)
// 		return
// 	}

// 	response, err := generateDnsResponse(&buf, n)
// 	if err != nil {
// 		fmt.Println("Ошибка при обработке DNS запроса:", err)
// 	}

// 	conn.Write(*response)
// }