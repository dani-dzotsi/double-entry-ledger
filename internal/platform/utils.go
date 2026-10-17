package platform

import(
	"github.com/google/uuid"
)


func IdempotencyKeyGenerator() string{
	idempotencykey := uuid.New().String()
	return idempotencykey
}