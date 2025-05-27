package handler

import (
	"net/http"

	"github.com/datphamcode295/go-lambda-pulumi/internal/core/domain"
	"github.com/datphamcode295/go-lambda-pulumi/internal/core/services"
	"github.com/gin-gonic/gin"
)

type PatientHandler struct {
	svc services.PatientService
}

func NewPatientHandler(PatientService services.PatientService) *PatientHandler {
	return &PatientHandler{
		svc: PatientService,
	}
}

func (h *PatientHandler) PayTransaction(ctx *gin.Context) {
	var data domain.PayTransactionRequest
	if err := ctx.ShouldBindJSON(&data); err != nil {
		HandleError(ctx, http.StatusBadRequest, err)
		return
	}

	rs, err := h.svc.PayTransaction(data)
	if err != nil {
		HandleError(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, rs)
}
