package cache

import (
	"bytes"
	"encoding/binary"
	"sync"
	"time"
)

type itemCache struct {
	resposta  []byte
	expiraEm time.Time
}

type Cache struct {
	mu    sync.RWMutex
	itens map[string]itemCache
}

func Novo() *Cache {
	return &Cache{itens: make(map[string]itemCache)}
}

func (c *Cache) Obter(dominio string, tipo uint16, novoID uint16) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	chave := string(rune(tipo)) + ":" + dominio
	item, ok := c.itens[chave]
	if !ok || time.Now().After(item.expiraEm) {
		return nil, false
	}

	// Atalho: clona a resposta e substitui os 2 primeiros bytes pelo ID da nova requisição
	resp := bytes.Clone(item.resposta)
	binary.BigEndian.PutUint16(resp[0:2], novoID)
	return resp, true
}

func (c *Cache) Salvar(dominio string, tipo uint16, resposta []byte, ttlSegundos uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	chave := string(rune(tipo)) + ":" + dominio
	c.itens[chave] = itemCache{
		resposta:  bytes.Clone(resposta),
		expiraEm: time.Now().Add(time.Duration(ttlSegundos) * time.Second),
	}
}