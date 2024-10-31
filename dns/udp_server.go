package dns

import (
	"fmt"
	"net"
	ut "dnsServer/utils"
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
		handleUDPConnection(conn, clientAddr, buf[:n])
	}
}

func handleUDPConnection(conn *net.UDPConn, clientAddr *net.UDPAddr, rawRequest []byte) {
	request, err := parseRequest(rawRequest)
	if err != nil {
		fmt.Println("Ошибка при обработке DNS запроса:", err)
		return
	}

	stackIPs := make(ut.Stack, 0)
	stackIPs  = stackIPs.PushRange(rootServersIPs)
	curIp := ""
	var answerResponse []byte
	for !stackIPs.IsEmpty() {
		stackIPs, curIp, _ = stackIPs.Pop()
		addr := fmt.Sprintf("%v:53", curIp)
		fmt.Print(addr)
		rawResponse, err := AskServerUDP(addr, rawRequest)
		if err != nil {
			continue
		}
		fmt.Print(rawResponse)
		response, err := parseResponse(rawResponse)
		if err != nil{
			continue
		}
		
		if len(response.Answers) > 0 {
			answerResponse = rawResponse
			break
		} else if len(response.Authorities) > 0 {
			for _, a := range response.Authorities {
				name, err := dataToString(types[a.Type], a.Data)
				if err != nil {
					continue
				}
				stackIPs = stackIPs.Push(name)
			}
		}
	}

	if answerResponse == nil {	
		response := Response{
			Header: request.Header,
			Question: request.Question,
			Answers: make([]*responseData, 0),
			Authorities: make([]*responseData, 0),
			Additionals: make([]*responseData, 0),
		}
		response.Header.QR = 1
		response.Header.RCode = 1
		response.Header.RA = 1
		response.Header.Z = 0
		response.Header.ARCount = 0
		encodedResponse := response.encode()
		conn.WriteToUDP(encodedResponse, clientAddr)
		return
	}

	conn.WriteToUDP(answerResponse, clientAddr)
}


func AskServerUDP(addr string, data []byte) ([]byte, error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при подключении: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при отправке: %w", err)
	}

	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении ответа: %w", err)
	}

	return buffer[:n], nil
}