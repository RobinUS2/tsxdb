package strategy

import (
	"../../backend"
)

type Abstract interface {
	GetBackend() backend.Abstract
}
