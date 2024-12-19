package bootstrap

import "sync"

type PreBootstrap struct {
	mu *sync.Mutex
}
