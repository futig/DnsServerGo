package main

import (
	dns "dnsServer/dns"
)


func main() {
	udpServer := dns.MakeUdpServer(dns.Address{Ip: "127.0.0.1", Port: 2053}, 10)
	go udpServer.Run()
	select {}
}
