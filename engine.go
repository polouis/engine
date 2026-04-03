package engine

import (
	"github.com/polouis/engine/types"
)

type Context struct {
	W *World
	B types.Backend
}
