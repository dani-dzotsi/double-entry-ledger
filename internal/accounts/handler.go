package accounts

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handler struct {
	svc Service
}

func NewHandler(svc Service) *handler {
	return &handler{svc: svc}
}

func (h *handler) GetAccountByID(ctx *gin.Context) {
	var id uuid.UUID
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid UUID format"})
		return
	}
	account, err := h.svc.GetAccountByID(ctx.Request.Context(), id)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, account)
}
