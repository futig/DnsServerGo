package dns

import (
	"context"
	ut "dnsServer/utils"
	"fmt"
	"net"
	"time"
)

type UdpServer struct {
	Address Address
	Cache   *Cache
}

func MakeUdpServer(addr Address, cashSize int) *UdpServer {
	return &UdpServer{
		Address: addr,
		Cache:   NewCache(cashSize),
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
		go handleUDPConnection(conn, clientAddr, buf[:n], s.Cache)
	}
}

func handleUDPConnection(conn *net.UDPConn, clientAddr *net.UDPAddr,
	rawRequest []byte, cache *Cache) {
	request, err := parseRequest(rawRequest)
	if err != nil {
		fmt.Println("Ошибка при обработке DNS запроса:", err)
		return
	}
	requestName := parseNameRecord(request.Question.QName)
	if cachedResponse, ok := cache.Get(requestName, request.Question.QType); ok {
		cachedResponse.Header.AA = 0
		cachedResponse.Header.ID = request.Header.ID
		conn.WriteToUDP(cachedResponse.encode(), clientAddr)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rawResponse, response, err := getAnswer(ctx, rawRequest)
	if err != nil {
		response := Response{
			Header:   request.Header,
			Question: request.Question,
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

	newTime := time.Now().Add(time.Duration(response.Answers[0].TTL) * time.Second)
	cache.Put(requestName, request.Question.QType, newTime, response)
}

func getAnswer(ctx context.Context, rawRequest []byte) ([]byte, *Response, error) {
	stackIPs := make(ut.Stack[string], 0)
	stackIPs = stackIPs.PushRange(rootServersIPs)
	curIp := ""
	for !stackIPs.IsEmpty() {
		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("произошел таймаут при чтении данных")
		default:
		}

		stackIPs, curIp, _ = stackIPs.Pop()
		addr := fmt.Sprintf("%v:53", curIp)
		rawResponse, err := askServerUDP(ctx, addr, rawRequest)
		if err != nil {
			continue
		}
		response, err := parseResponse(rawResponse)
		if err != nil {
			continue
		}

		if len(response.Answers) > 0 {
			return rawResponse, response, nil
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
	return nil, nil, fmt.Errorf("не удалось зарезолвить адрес")
}

func askServerUDP(ctx context.Context, addr string, data []byte) ([]byte, error) {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "udp", addr)
	if err != nil {
		return nil, fmt.Errorf("ошибка при подключении: %w", err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

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
