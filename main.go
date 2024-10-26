package main

import (
	dns "dnsServer/dns"
)


func main() {
	udpServer := dns.MakeUdpServer(dns.Address{Ip: "127.0.0.1", Port: 2053}, 10)
	udpServer.Run()


	// addr := "8.8.8.8:53"

	// hexStream := "000601000001000000000000096861627261686162720272750000010001"
	// binaryData, err := hex.DecodeString(hexStream)

	// if err != nil {
	// 	fmt.Println("Ошибка декодирования hex-строки:", err)
	// 	return
	// }

	// conn, err := net.Dial("udp", addr)
	// if err != nil {
	// 	fmt.Println("Ошибка при подключении:", err)
	// 	return
	// }
	// defer conn.Close()

	// _, err = conn.Write(binaryData)
	// if err != nil {
	// 	fmt.Println("Ошибка при отправке:", err)
	// 	return
	// }

	// buffer := make([]byte, 1024)

	// n, err := conn.Read(buffer)
	// if err != nil {
	// 	fmt.Println("Ошибка при получении ответа:", err)
	// 	return
	// }

	// fmt.Printf("Получен ответ: %x\n", buffer[:n])
}
