package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"godns/internal/blocklist"
	"godns/internal/cache"
	"godns/internal/dnsmsg"
)

var upstreamDNS = "1.1.1.1:53"

func main() {
	porta := flag.Int("port", 8053, "Porta UDP para escutar")
	flag.Parse()

	bl := blocklist.Nova()
	dnsCache := cache.Novo()

	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: *porta})
	if err != nil {
		log.Fatalf("Erro ao abrir porta UDP %d: %v", *porta, err)
	}
	defer conn.Close()

	fmt.Printf("GoDNS rodando na porta :%d. Upstream: %s\n", *porta, upstreamDNS)

	buf := make([]byte, 4096)
	for {
		n, cliente, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		// Copia o buffer para isolamento por goroutine
		pkt := make([]byte, n)
		copy(pkt, buf[:n])

		// Requisito: Uma goroutine por consulta
		go func(p []byte, addr *net.UDPAddr) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recuperado de pânico no pacote: %v", r)
				}
			}()

			msg, err := dnsmsg.Parse(p)
			if err != nil {
				return
			}

			// 1. Checagem de Blocklist
			if bl.Bloqueado(msg.Dominio) {
				conn.WriteToUDP(dnsmsg.RespostaBloqueada(msg), addr)
				return
			}

			// 2. Checagem de Cache
			if resp, hit := dnsCache.Obter(msg.Dominio, msg.Tipo, msg.ID); hit {
				conn.WriteToUDP(resp, addr)
				return
			}

			// 3. Encaminhamento para Upstream
			respUpstream, err := resolverUpstream(p)
			if err == nil && len(respUpstream) > 0 {
				dnsCache.Salvar(msg.Dominio, msg.Tipo, respUpstream, 60)
				conn.WriteToUDP(respUpstream, addr)
			}
		}(pkt, cliente)
	}
}

func resolverUpstream(pkt []byte) ([]byte, error) {
	conn, err := net.Dial("udp", upstreamDNS)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if _, err := conn.Write(pkt); err != nil {
		return nil, err
	}

	resp := make([]byte, 4096)
	n, err := conn.Read(resp)
	if err != nil {
		return nil, err
	}
	return resp[:n], nil
}
