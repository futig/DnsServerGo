package dns

import (
	"fmt"
	"net"
	ut "dnsServer/utils"
)

type UdpServer struct {
	Address  Address
	Cache *ut.Cache
}

func MakeUdpServer(addr Address, cashSize int) *UdpServer {
	return &UdpServer{
		Address: addr,
		Cache: ut.NewCache(cashSize),
	}
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
		handleUDPConnection(conn, clientAddr, buf[:n], s.Cache)
	}
}

func handleUDPConnection(conn *net.UDPConn, clientAddr *net.UDPAddr, 
	rawRequest []byte, cache *ut.Cache) {
	request, err := parseRequest(rawRequest)
	if err != nil {
		fmt.Println("Ошибка при обработке DNS запроса:", err)
		return
	}
	requestName := parseNameRecord(request.Question.QName)
	if cachedAnswer, ok := cache.Get(requestName, request.Question.QType); ok {
		conn.WriteToUDP(cachedAnswer, clientAddr)
	}

	rawResponse, response := getAnswer(rawRequest)

	if rawResponse == nil {	
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

	conn.WriteToUDP(rawResponse, clientAddr)

	response.Header.AA = 0
	cache.Put(requestName, request.Question.QType, response.encode())
	fmt.Print(cache)
}


func getAnswer(rawRequest []byte) ([]byte, *Response) {
	stackIPs := make(ut.Stack[string], 0)
	stackIPs  = stackIPs.PushRange(rootServersIPs)
	curIp := ""
	for !stackIPs.IsEmpty() {
		stackIPs, curIp, _ = stackIPs.Pop()
		addr := fmt.Sprintf("%v:53", curIp)
		rawResponse, err := askServerUDP(addr, rawRequest)
		if err != nil {
			continue
		}
		response, err := parseResponse(rawResponse)
		if err != nil{
			continue
		}
		
		if len(response.Answers) > 0 {
			return rawResponse, response
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
	return nil, nil
}


func askServerUDP(addr string, data []byte) ([]byte, error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		return nil, fmt.Errorf("ошибка при отправке: %w", err)
	}

	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении ответа: %w", err)
	}

	return buffer[:n], nil
}