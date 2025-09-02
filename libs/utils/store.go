package utils

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Store is struct for persist values
type Store struct {
	store *sync.Map
}

// NewStore return instance of store
func NewStore() *Store {
	return &Store{
		store: &sync.Map{},
	}
}

// Get return value the key
func (s *Store) Get(key string) any {
	if value, ok := s.store.Load(key); ok {
		return value
	}
	return nil
}

// Set execute store of key
func (s *Store) Set(key string, value any) error {
	if len(key) == 0 {
		return errors.New("invalid key length")
	}
	s.store.Store(key, value)
	return nil
}

// StartCleanupGoroutine execute cleanning
func (s *Store) StartCleanupGoroutine(excludePrefixes []string, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			fmt.Println("Execute cleanup store keys ...")
			s.store.Range(func(key, value any) bool {
				k, ok := key.(string)
				if !ok{
					return true
				}

				// Verificar se a chave deve ser excluída da limpeza
				shouldExclude := false
				for _, prefix := range excludePrefixes {
					if strings.HasPrefix(k, prefix) {
						shouldExclude = true
						break // Sai do loop de prefixos assim que um for encontrado
					}
				}

				// Se a chave não estiver na lista de exclusão, a removemos.
				if !shouldExclude {
					fmt.Printf("Removing key from store: %s\n", k)
					s.store.Delete(k)
				}

				return true
			})
		}
	}()
}
