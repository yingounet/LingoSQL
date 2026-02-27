package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"lingosql/internal/models"
	"lingosql/internal/service"
	"lingosql/internal/utils"
)

type FavoriteHandler struct {
	favoriteService *service.FavoriteService
}

func NewFavoriteHandler(favoriteService *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favoriteService: favoriteService}
}

// GetFavorites 获取收藏列表
// 查询参数: connection_id (可选), database (可选), sort (created_at | last_used_at，默认 created_at)
func (h *FavoriteHandler) GetFavorites(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	var connectionID *int
	if s := c.Query("connection_id"); s != "" {
		if id, err := strconv.Atoi(s); err == nil {
			connectionID = &id
		}
	}
	var database *string
	if s := c.Query("database"); s != "" {
		database = &s
	}
	sort := c.DefaultQuery("sort", "created_at")
	if sort != "last_used_at" {
		sort = "created_at"
	}

	favorites, err := h.favoriteService.GetByUserID(userID, connectionID, database, sort)
	if err != nil {
		utils.InternalServerError(c, err.Error())
		return
	}

	response := make([]models.FavoriteResponse, len(favorites))
	for i, fav := range favorites {
		connName := ""
		if conn, err := h.favoriteService.GetConnectionName(fav.ConnectionID); err == nil {
			connName = conn
		}
		response[i] = models.FavoriteResponse{
			ID:             fav.ID,
			ConnectionID:   fav.ConnectionID,
			ConnectionName: connName,
			Database:       fav.Database,
			Name:           fav.Name,
			SQLQuery:       fav.SQLQuery,
			Description:    fav.Description,
			CreatedAt:      fav.CreatedAt,
			LastUsedAt:     fav.LastUsedAt,
		}
	}
	utils.Success(c, response)
}

// CreateFavorite 创建收藏
func (h *FavoriteHandler) CreateFavorite(c *gin.Context) {
	var req models.FavoriteCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	fav, err := h.favoriteService.Create(userID, &req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	connName, _ := h.favoriteService.GetConnectionName(fav.ConnectionID)
	utils.SuccessWithMessage(c, "收藏添加成功", models.FavoriteResponse{
		ID:             fav.ID,
		ConnectionID:   fav.ConnectionID,
		ConnectionName: connName,
		Database:       fav.Database,
		Name:           fav.Name,
		SQLQuery:       fav.SQLQuery,
		Description:    fav.Description,
		CreatedAt:      fav.CreatedAt,
		LastUsedAt:     fav.LastUsedAt,
	})
}

// UpdateFavorite 更新收藏
func (h *FavoriteHandler) UpdateFavorite(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的收藏 ID")
		return
	}

	var req models.FavoriteUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	fav, err := h.favoriteService.Update(id, userID, &req)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	connName, _ := h.favoriteService.GetConnectionName(fav.ConnectionID)
	utils.SuccessWithMessage(c, "收藏更新成功", models.FavoriteResponse{
		ID:             fav.ID,
		ConnectionID:   fav.ConnectionID,
		ConnectionName: connName,
		Database:       fav.Database,
		Name:           fav.Name,
		SQLQuery:       fav.SQLQuery,
		Description:    fav.Description,
		CreatedAt:      fav.CreatedAt,
		LastUsedAt:     fav.LastUsedAt,
	})
}

// RecordFavoriteUse 记录收藏使用，更新 last_used_at
func (h *FavoriteHandler) RecordFavoriteUse(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的收藏 ID")
		return
	}
	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.favoriteService.RecordUse(id, userID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	utils.SuccessWithMessage(c, "已记录", nil)
}

// DeleteFavorite 删除收藏
func (h *FavoriteHandler) DeleteFavorite(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的收藏 ID")
		return
	}

	userID, ok := GetUserID(c)
	if !ok {
		return
	}
	if err := h.favoriteService.Delete(id, userID); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "收藏删除成功", nil)
}
