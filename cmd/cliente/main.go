package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	// Pacote DNS simplificado perguntando por "ads.site.com" tipo A (IPv4)
	query := []byte{
		0xaa, 0xbb, // ID da transação
		0x01, 0x00, // Flags: consulta padrão recursiva
		0x00, 0x01, // 1 Pergunta
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		// ads.site.com
		0x03, 'a', 'd', 's',
		0x04, 's', 'i', 't', 'e',
		0x03, 'c', 'o', 'm',
		0x00,       // fim do nome
		0x00, 0x01, // Tipo A
		0x00, 0x01, // Classe IN
	}

	inicio := time.Now()
	conn, err := net.Dial("udp", "127.0.0.1:8053")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	conn.Write(query)

	resp := make([]byte, 512)
	n, err := conn.Read(resp)
	if err != nil {
		panic(err)
	}

	duracao := time.Since(inicio)
	fmt.Printf("Recebidos %d bytes em %v\n", n, duracao)
	if n >= 16 {
		ip := resp[n-4 : n]
		fmt.Printf("IP devolvido: %d.%d.%d.%d\n", ip[0], ip[1], ip[2], ip[3])
	}
}
