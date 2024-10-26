package dns

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

func parseRequest(bufPointer *[]byte, n int) (*header, *question, error) {
	header, err := readHeader(bufPointer, n)
	if err != nil {
		return nil, nil, err
	}
	if header.QDCount == 0 || header.QDCount > 1 {
		return nil, nil, fmt.Errorf("Недопустимое число вопросов")
	}
	if header.OPCode != 0 {
		return nil, nil, fmt.Errorf("Недопустимый тип запроса")
	}
	question, _ := readQuestion(bufPointer, 12)
	return header, question, nil
}

func parseResponse(bufPointer *[]byte, n int) (
	header *header, answers []*responseData, authorities []*responseData, additionals []*responseData, err error) {
	header, err = readHeader(bufPointer, n)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if header.RCode != 0 {
		return nil, nil, nil, nil, fmt.Errorf("Сервер не смог ответить на вопрос: %d", header.RCode)
	}
	_, pos := readQuestion(bufPointer, 12)
	for range header.ANCount {
		data, ind, err := readResponseData(bufPointer, pos)
		if err != nil {
			continue
		}
		pos = ind
		answers = append(answers, data)
	}
	for range header.NSCount {
		data, ind, err := readResponseData(bufPointer, pos)
		if err != nil {
			continue
		}
		pos = ind
		authorities = append(answers, data)
	}
	for range header.ARCount {
		data, ind, err := readResponseData(bufPointer, pos)
		if err != nil {
			continue
		}
		pos = ind
		additionals = append(answers, data)
	}
	err = nil
	return
}

func readHeader(bufPointer *[]byte, n int) (*header, error) {
	if n < 12 {
		return nil, fmt.Errorf("Заголовок DNS должен состоять из 12 байт")
	}
	buf := *bufPointer
	h := &header{
		ID:      uint16(buf[0])<<8 | uint16(buf[1]),
		QR:      uint16(buf[2] >> 7),
		OPCode:  uint16((buf[2] << 1) >> 4),
		AA:      uint16((buf[2] << 5) >> 7),
		TC:      uint16((buf[2] << 6) >> 7),
		RD:      uint16((buf[2] << 7) >> 7),
		RA:      uint16(buf[3] >> 7),
		Z:       uint16((buf[3] << 1) >> 5),
		RCode:   uint16((buf[3] << 4) >> 4),
		QDCount: uint16(buf[4])<<8 | uint16(buf[5]),
		ANCount: uint16(buf[6])<<8 | uint16(buf[7]),
		NSCount: uint16(buf[8])<<8 | uint16(buf[9]),
		ARCount: uint16(buf[10])<<8 | uint16(buf[11]),
	}
	return h, nil
}

func (h *header) encode() *[]byte {
	dnsHeader := make([]byte, 12)

	var flags uint16 = 0
	flags = h.QR<<15 | h.OPCode<<11 | h.AA<<10 | h.TC<<9 | h.RD<<8 | h.RA<<7 | h.Z<<4 | h.RCode

	binary.BigEndian.PutUint16(dnsHeader[0:2], h.ID)
	binary.BigEndian.PutUint16(dnsHeader[2:4], flags)
	binary.BigEndian.PutUint16(dnsHeader[4:6], h.QDCount)
	binary.BigEndian.PutUint16(dnsHeader[6:8], h.ANCount)
	binary.BigEndian.PutUint16(dnsHeader[8:10], h.NSCount)
	binary.BigEndian.PutUint16(dnsHeader[10:12], h.ARCount)

	return &dnsHeader
}

func readQuestion(bufPointer *[]byte, start int) (*question, int) {
	buf := *bufPointer
	questionName, ind := readNameRecord(bufPointer, start)

	questionType := binary.BigEndian.Uint16(buf[start : start+2])
	questionClass := binary.BigEndian.Uint16(buf[start+2 : start+4])

	q := question{
		QName:  parseNameRecord(questionName),
		QType:  questionType,
		QClass: questionClass,
	}

	return &q, ind + 4
}

func (q *question) Encode() []byte {
	domain := q.QName
	parts := strings.Split(domain, ".")

	var buf bytes.Buffer

	for _, label := range parts {
		if len(label) > 0 {
			buf.WriteByte(byte(len(label)))
			buf.WriteString(label)
		}
	}
	buf.WriteByte(0x00)
	buf.Write(intToBytes(uint16(q.QType)))
	buf.Write(intToBytes(uint16(q.QClass)))

	return buf.Bytes()
}

func intToBytes(u uint16) []byte {
	bytes := make([]byte, 2)
	bytes[0] = byte(u >> 8)
	bytes[1] = byte((u << 8) >> 8)
	return bytes
}

func readResponseData(bufPointer *[]byte, start int) (*responseData, int, error) {
	buf := *bufPointer
	name, ind := readNameRecord(bufPointer, start)

	rType := binary.BigEndian.Uint16(buf[ind : ind+2])
	rClass := binary.BigEndian.Uint16(buf[ind+2 : ind+4])
	timeToLive := binary.BigEndian.Uint32(buf[ind+4 : ind+8])
	dataLength := binary.BigEndian.Uint32(buf[ind+8 : ind+10])
	if _, ok := types[rType]; !ok {
		return nil, 0, fmt.Errorf("Недопустимый тип записи: %d", rType)
	}

	var data *[]byte
	switch types[rType] {
	case "A":
		data, ind = readIpv4(bufPointer, ind+10)
	case "AAAA":
		data, ind = readIpv6(bufPointer, ind+10)
	case "MX":
		data, ind = readMxRecord(bufPointer, ind+10)
	case "NS", "CNAME":
		data, ind = readNameRecord(bufPointer, ind+10)
	}

	d := responseData{
		Name:       parseNameRecord(name),
		Type:       rType,
		Class:      rClass,
		TTL:        timeToLive,
		DataLength: dataLength,
		Data:       data,
	}

	return &d, ind, nil
}

func readIpv4(bufPointer *[]byte, start int) (*[]byte, int) {
	buf := *bufPointer
	ip := buf[start : start+4]
	return &ip, start + 4
}

func readIpv6(bufPointer *[]byte, start int) (*[]byte, int) {
	buf := *bufPointer
	ip := buf[start : start+16]
	return &ip, start + 16
}

func readMxRecord(bufPointer *[]byte, start int) (*[]byte, int) {
	buf := *bufPointer
	var record []byte
	record = append(record, buf[start:start+2]...)
	name, ind := readNameRecord(bufPointer, start+2)
	record = append(record, *name...)
	return &record, ind
}

func readNameRecord(bufPointer *[]byte, pos int) (*[]byte, int) {
	buf := *bufPointer
	var record []byte
	var old int = 0
	for {
		if mark := uint16(buf[pos] >> 6); mark != 0 {
			if old == 0 {
				old = pos + 2
			}
			pos = int((buf[pos]<<2)>>2)<<8 | int(buf[pos+1])
			continue
		}
		lengthByte := (buf[pos] << 2) >> 2
		length := int(lengthByte)
		if length == 0 {
			break
		}
		pos++
		record = append(record, lengthByte)
		record = append(record, buf[pos:pos+length]...)
		pos += length
	}
	var nextPos int
	if old != 0 {
		nextPos = old
	} else {
		nextPos = pos + 1
	}

	return &record, nextPos
}

func parseIpv4(bufPointer *[]byte) string {

}

func parseIpv6(bufPointer *[]byte) string {

}

func parseMxRecord(bufPointer *[]byte) string {

}

func parseNameRecord(bufPointer *[]byte) string {
	buf := *bufPointer
	var nameParts []string
	pos := 0
	for pos < len(buf) {
		length := int(buf[pos])
		nameParts = append(nameParts, string(buf[pos:pos+length]))
		pos += length
	}
	return strings.Join(nameParts, ".")
}

func (a *answer) Encode() []byte {
	var rrBytes []byte

	domain := a.Name
	parts := strings.Split(domain, ".")

	for _, label := range parts {
		if len(label) > 0 {
			rrBytes = append(rrBytes, byte(len(label)))
			rrBytes = append(rrBytes, []byte(label)...)
		}
	}
	rrBytes = append(rrBytes, 0x00)

	rrBytes = append(rrBytes, intToBytes(uint16(a.Type))...)
	rrBytes = append(rrBytes, intToBytes(uint16(a.Class))...)

	time := make([]byte, 4)
	binary.BigEndian.PutUint32(time, a.TTL)

	rrBytes = append(rrBytes, time...)
	rrBytes = append(rrBytes, intToBytes(a.Length)...)

	ipBytes, err := net.IPv4(a.Data[0], a.Data[1], a.Data[2], a.Data[3]).MarshalText()
	if err != nil {
		return nil
	}

	rrBytes = append(rrBytes, ipBytes...)

	return rrBytes
}
