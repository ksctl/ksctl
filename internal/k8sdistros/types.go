package k8sdistros

import "sync"

type PreBootstrap struct {
	mu *sync.Mutex
}
