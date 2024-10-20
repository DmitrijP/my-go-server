package handlers

import (
	"sync/atomic"

	"github.com/DmitrijP/my-go-server/internal/database"
)

type ApiConfig struct {
	Jwt_secret     string
	FileserverHits atomic.Int32
	Db             database.Queries
	PolkaKey       string
}
