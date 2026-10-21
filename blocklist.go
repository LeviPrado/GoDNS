package blocklist

import (
	"bufio"
	"bytes"
	_ "embed"
	"strings"
)

var listaPadrao []byte

type Blocklist struct {
	dominios map[string]struct{}
}

func Nova() *Blocklist {
	b := &Blocklist{
		dominios: make(map[string]struct{}),
	}
	b.carregarLista(listaPadrao)
	return b
}

func (b *Blocklist) carregarLista(dados []byte) {
	scanner := bufio.NewScanner(bytes.NewReader(dados))

	for scanner.Scan() {
		linha := strings.TrimSpace(scanner.Text())

		if linha != "" && !strings.HasPrefix(linha, "#") {
			b.dominios[strings.ToLower(linha)] = struct{}{}
		}
	}
}

func (b *Blocklist) Bloqueado(dominio string) bool {
	dominio = strings.ToLower(strings.TrimSuffix(dominio, "."))

	for {
		if _, existe := b.dominios[dominio]; existe {
			return true
		}

		ponto := strings.Index(dominio, ".")
		if ponto == -1 {
			break
		}

		dominio = dominio[ponto+1:]
	}

	return false
}
