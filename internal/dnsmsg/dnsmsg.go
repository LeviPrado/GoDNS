package dnsmsg

import (
	"encoding/binary"
	"errors"
	"strings"
)

type Mensagem struct {
	ID      uint16
	Dominio string
	Tipo    uint16
	Raw     []byte
}

// Parse extrai o ID, domínio e tipo da pergunta sem alocações pesadas
func Parse(pkt []byte) (*Mensagem, error) {
	if len(pkt) < 12 {
		return nil, errors.New("pacote DNS curto demais")
	}

	id := binary.BigEndian.Uint16(pkt[0:2])
	offset := 12
	var labels []string

	for {
		if offset >= len(pkt) {
			return nil, errors.New("pacote malformado no nome")
		}
		length := int(pkt[offset])
		offset++

		if length == 0 {
			break
		}
		if offset+length > len(pkt) {
			return nil, errors.New("tamanho de label excede o pacote")
		}

		labels = append(labels, string(pkt[offset:offset+length]))
		offset += length
	}

	if offset+4 > len(pkt) {
		return nil, errors.New("pacote sem seção de tipo/classe")
	}

	tipo := binary.BigEndian.Uint16(pkt[offset : offset+2])
	dominio := strings.Join(labels, ".")

	return &Mensagem{
		ID:      id,
		Dominio: dominio,
		Tipo:    tipo,
		Raw:     pkt,
	}, nil
}

// RespostaBloqueada cria uma resposta DNS A (0.0.0.0) reaproveitando o ID e a Query
func RespostaBloqueada(m *Mensagem) []byte {
	resp := make([]byte, len(m.Raw)+16)
	copy(resp, m.Raw)

	// Flags: QR=1 (resposta), AA=1, RD=1, RA=1, RCODE=0 -> 0x8180
	binary.BigEndian.PutUint16(resp[2:4], 0x8180)
	// ANCOUNT = 1
	binary.BigEndian.PutUint16(resp[6:8], 1)

	// Monta Answer apontando via ponteiro de compressão para a pergunta (0xc00c)
	idx := len(m.Raw)
	resp[idx] = 0xc0
	resp[idx+1] = 0x0c // Ponteiro para o byte 12 (início do nome)
	binary.BigEndian.PutUint16(resp[idx+2:idx+4], 1)   // Tipo A
	binary.BigEndian.PutUint16(resp[idx+4:idx+6], 1)   // Classe IN
	binary.BigEndian.PutUint32(resp[idx+6:idx+10], 60) // TTL: 60s
	binary.BigEndian.PutUint16(resp[idx+10:idx+12], 4) // RDLENGTH = 4 bytes (IPv4)
	copy(resp[idx+12:idx+16], []byte{0, 0, 0, 0})     // IP: 0.0.0.0

	return resp
}