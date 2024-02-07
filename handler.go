package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jkittell/data/database"
	"net/http"
)

func HandleSegmentCheckInfo(infoDB database.MongoDB[SegmentCheckInfo]) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		id := c.Param("id")
		strId, err := uuid.Parse(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		data, err := infoDB.FindByID(context.TODO(), "segment_check_info", strId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, data)

	}
	return fn
}
