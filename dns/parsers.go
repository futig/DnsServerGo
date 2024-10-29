package dns

import (
	"encoding/binary"
	"fmt"
	"strings"
)

func parseRequest(buf []byte) (*Request, error) {
	header, err := readHeader(buf)
	if err != nil {
		return nil, err
	}
	if header.QDCount == 0 || header.QDCount > 1 {
		return nil, fmt.Errorf("Недопустимое число вопросов: %w", header.QDCount)
	}
	if header.OPCode != 0 {
		return nil, fmt.Errorf("Недопустимый тип запроса: %w", header.OPCode)
	}
	question, _ := readQuestion(buf, 12)
	request := &Request{
		Header:   *header,
		Question: *question,
	}
	return request, nil
}

func parseResponse(buf []byte) (*Response, error) {
	header, err := readHeader(buf)
	if err != nil {
		return nil, err
	}
	question, pos := readQuestion(buf, 12)

	parts := [...][]*responseData{
		make([]*responseData, header.ANCount),
		make([]*responseData, header.NSCount),
		make([]*responseData, header.ARCount),
	}

	for _, values := range parts {
		n := len(values)
		values = values[:0]
		for range n {
			data, ind, err := readResponseData(buf, pos)
			if err != nil {
				continue
			}
			pos = ind
			values = append(values, data)
		}
	}

	response := &Response{
		Header:      *header,
		Question:    *question,
		Answers:     parts[0],
		Authorities: parts[1],
		Additionals: parts[2],
	}
	return response, nil
}

func (r *Request) encode() []byte {
	var request []byte
	var names map[string]uint16
	request = append(request, r.Header.encode()...)
	request = append(request, r.Question.encode(12, &names)...)
	return request
}

func (r *Response) encode() []byte {
	var response []byte
	var names map[string]uint16
	response = append(response, r.Header.encode()...)
	response = append(response, r.Question.encode(len(response), &names)...)

	parts := [...][]*responseData{
		r.Answers,
		r.Authorities,
		r.Additionals,
	}

	for _, values := range parts {
		for _, dataPtr := range values {
			data := *dataPtr
			response = append(response, data.encode(len(response), &names)...)
		}
	}
	
	return response
}

func readHeader(buf []byte) (*header, error) {
	if len(buf) < 12 {
		return nil, fmt.Errorf("Заголовок DNS должен состоять из 12 байт")
	}
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

func (h *header) encode() []byte {
	header := make([]byte, 12)

	var flags uint16 = 0
	flags = h.QR<<15 | h.OPCode<<11 | h.AA<<10 | h.TC<<9 | h.RD<<8 | h.RA<<7 | h.Z<<4 | h.RCode

	binary.BigEndian.PutUint16(header[0:2], h.ID)
	binary.BigEndian.PutUint16(header[2:4], flags)
	binary.BigEndian.PutUint16(header[4:6], h.QDCount)
	binary.BigEndian.PutUint16(header[6:8], h.ANCount)
	binary.BigEndian.PutUint16(header[8:10], h.NSCount)
	binary.BigEndian.PutUint16(header[10:12], h.ARCount)

	return header
}

func readQuestion(buf []byte, start int) (*question, int) {
	questionName, ind := readNameRecord(buf, start)

	questionType := binary.BigEndian.Uint16(buf[start : start+2])
	questionClass := binary.BigEndian.Uint16(buf[start+2 : start+4])

	q := question{
		QName:  questionName,
		QType:  questionType,
		QClass: questionClass,
	}

	return &q, ind + 4
}

func (q *question) encode(start int, namesPtr *map[string]uint16) []byte {
	var question []byte
	question = append(question, writeName(q.QName, start, namesPtr)...)
	question = append(question, uint16ToBytes(q.QType)...)
	question = append(question, uint16ToBytes(q.QClass)...)

	return question
}

func readResponseData(buf []byte, start int) (*responseData, int, error) {
	name, ind := readNameRecord(buf, start)

	rType := binary.BigEndian.Uint16(buf[ind : ind+2])
	if _, ok := types[rType]; !ok {
		return nil, 0, fmt.Errorf("Недопустимый тип записи: %d", rType)
	}

	rClass := binary.BigEndian.Uint16(buf[ind+2 : ind+4])
	timeToLive := binary.BigEndian.Uint32(buf[ind+4 : ind+8])
	dataLength := binary.BigEndian.Uint16(buf[ind+8 : ind+10])

	var data []byte
	switch types[rType] {
	case "A":
		data, ind = readIpv4(buf, ind+10)
	case "AAAA":
		data, ind = readIpv6(buf, ind+10)
	case "MX":
		data, ind = readMxRecord(buf, ind+10)
	case "NS", "CNAME":
		data, ind = readNameRecord(buf, ind+10)
	}

	d := responseData{
		Name:       name,
		Type:       rType,
		Class:      rClass,
		TTL:        timeToLive,
		DataLength: dataLength,
		Data:       data,
	}

	return &d, ind, nil
}

func (r *responseData) encode(start int, namesPtr *map[string]uint16) []byte {
	var response []byte

	response = append(response, writeName(r.Name, start, namesPtr)...)
	response = append(response, uint16ToBytes(r.Type)...)
	response = append(response, uint16ToBytes(r.Class)...)

	time := make([]byte, 4)
	binary.BigEndian.PutUint32(time, r.TTL)

	response = append(response, time...)
	response = append(response, uint16ToBytes(r.DataLength)...)

	switch types[r.Type] {
	case "A", "AAAA":
		response = append(response, r.Data...)
	case "MX":
		response = append(response, r.Data[:2]...)
		response = append(response, writeName(r.Data[2:], start+len(response), namesPtr)...)
	case "NS", "CNAME":
		response = append(response, writeName(r.Data, start+len(response), namesPtr)...)
	}

	return response
}

func writeName(name []byte, start int, namesPtr *map[string]uint16) []byte {
	names := *namesPtr
	var data []byte
	noLinks := true
	for len(name) > 0 {
		nameKey := parseNameRecord(name)
		if address, ok := names[nameKey]; ok {
			bytesPos := uint16ToBytes(address)
			data = append(data, byte(3)<<6|bytesPos[0])
			data = append(data, bytesPos[1])
			noLinks = false
			break
		} else {
			length := int(name[0])
			data = append(data, name[0])
			data = append(data, name[:length]...)
			names[nameKey] = uint16(start)
			start += length + 1
			name = name[length+1:]
		}
	}
	if noLinks {
		data = append(data, 0x00)
	}
	return data
}

func readIpv4(buf []byte, start int) ([]byte, int) {
	return buf[start : start+4], start + 4
}

func readIpv6(buf []byte, start int) ([]byte, int) {
	return buf[start : start+16], start + 16
}

func readMxRecord(buf []byte, start int) ([]byte, int) {
	var data []byte
	data = append(data, buf[start:start+2]...)
	name, ind := readNameRecord(buf, start+2)
	data = append(data, name...)
	return data, ind
}

func readNameRecord(buf []byte, pos int) ([]byte, int) {
	var record []byte
	var old int = -1
	for {
		if mark := buf[pos] >> 6; mark != 0 {
			if old == -1 {
				old = pos + 2
			}
			pos = int(buf[pos]<<2)<<6 | int(buf[pos+1])
			continue
		}
		length := int((buf[pos] << 2) >> 2)
		if length == 0 {
			break
		}
		pos++
		record = append(record, byte(length))
		record = append(record, buf[pos:pos+length]...)
		pos += length
	}
	var nextPos int
	if old != -1 {
		nextPos = old
	} else {
		nextPos = pos + 1
	}

	return record, nextPos
}

func parseIpv4(buf []byte) string {
	res := make([]string, 4)
	for i, el := range buf {
		res[i] = string(el)
	}
	return strings.Join(res, ".")
}

func parseIpv6(buf []byte) string {
	res := make([]string, 8)
	for i := 0; i < 16; i += 2 {
		part := uint16(buf[i])<<8 | uint16(buf[i+1])
		res[i] = string(part)
	}
	return strings.Join(res, ".")
}

func parseMxRecord(buf []byte) (int, string) {
	priority := uint16(buf[0])<<8 | uint16(buf[1])
	bufName := buf[2:]
	name := parseNameRecord(bufName)
	return int(priority), name
}

func parseNameRecord(buf []byte) string {
	var nameParts []string
	pos := 0
	for pos < len(buf) {
		length := int(buf[pos])
		nameParts = append(nameParts, string(buf[pos:pos+length]))
		pos += length
	}
	return strings.Join(nameParts, ".")
}

func uint16ToBytes(u uint16) []byte {
	bytes := make([]byte, 2)
	bytes[0] = byte(u >> 8)
	bytes[1] = byte((u << 8) >> 8)
	return bytes
}
