package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/David89Lee/pyrolysis_ty_server/model"
	"github.com/David89Lee/pyrolysis_ty_server/router/response"
	"github.com/David89Lee/pyrolysis_ty_server/service/record_service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const queryTimeFormat = "2006-01-02 15:04:05"

// ===== 设备信息（Info）增删改查 =====

// @Summary 信息
// @Tags 全信息查询
// @Accept json
// @Produce  json
// @Success 200 {object} response.Response{data=[]model.Info} "成功"
// @Failure 500 {object} response.Response
// @Router /api/v1/info [get]
func FindInfo(c *gin.Context) {
	appG := response.Gin{C: c}
	pageNum, pageSizeNum := pagination(c)

	recordService := record_service.RecordService{}
	infos, total, err := recordService.FindInfo(pageNum, pageSizeNum)
	if err != nil {
		zap.L().Error("fail to find info", zap.Error(err))
		appG.Response(http.StatusOK, response.ErrorFindInfo, err)
		return
	}
	appG.PageDataResponse(http.StatusOK, response.SUCCESS, total, pageSizeNum, pageNum, infos)
}

// @Summary 信息
// @Tags 创建信息
// @Accept json
// @Produce  json
// @Param info body model.Info true "信息"
// @Success 200 {object} response.Response{data=model.Info}
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response
// @Router /api/v1/info [post]
func CreateInfo(c *gin.Context) {
	appG := response.Gin{C: c}

	info := &model.Info{}
	err := c.BindJSON(info)
	if err != nil {
		zap.L().Error("BindJSON error", zap.Error(err))
		appG.Response(http.StatusOK, response.InvalidParams, err)
		return
	}
	recordService := record_service.RecordService{}

	info, err = recordService.CreateInfo(info)
	if err != nil {
		zap.L().Error("fail to create info", zap.Error(err))
		appG.Response(http.StatusOK, response.ErrorCreateInfo, err)
		return
	}
	appG.Response(http.StatusOK, response.SUCCESS, info)
}

// @Summary 信息
// @Tags 更新信息
// @Accept json
// @Produce  json
// @Param info body model.Info true "信息"
// @Success 200 {object} response.Response{data=model.Info}
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response
// @Router /api/v1/info [put]
func UpdateInfo(c *gin.Context) {
	appG := response.Gin{C: c}

	info := &model.Info{}
	err := c.BindJSON(info)
	if err != nil {
		zap.L().Error("BindJSON error", zap.Error(err))
		appG.Response(http.StatusOK, response.InvalidParams, err)
		return
	}
	if info.ID == 0 {
		appG.Response(http.StatusOK, response.InvalidParams, "id不能为空")
		return
	}
	recordService := record_service.RecordService{}

	info, err = recordService.UpdateInfo(info)
	if err != nil {
		zap.L().Error("fail to update info", zap.Error(err))
		appG.Response(http.StatusOK, response.ErrorUpdateInfo, err)
		return
	}
	appG.Response(http.StatusOK, response.SUCCESS, info)
}

// @Summary 信息
// @Tags 删除信息
// @Accept json
// @Produce  json
// @Param id path int true "信息id"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "参数错误"
// @Failure 500 {object} response.Response
// @Router /api/v1/info/{id} [delete]
func DeleteInfo(c *gin.Context) {
	appG := response.Gin{C: c}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		zap.L().Error("parse id error")
		appG.Response(response.ServerError, response.InvalidParams, err.Error())
		return
	}

	recordService := record_service.RecordService{}
	err = recordService.DeleteInfo(uint(id))
	if err != nil {
		zap.L().Error("fail to delete info", zap.Error(err))
		appG.Response(response.ServerError, response.ErrorDeleteInfo, err.Error())
		return
	}

	appG.Response(http.StatusOK, response.SUCCESS, nil)
}

// ===== DTU 采样数据查询 =====

// @Summary 记录
// @Tags 获取指定dtu的记录数据（默认当天）
// @Accept json
// @Produce  json
// @Param dtu path int true "dtu编号 1~4"
// @Param start query string false "开始时间 2006-01-02 15:04:05"
// @Param end query string false "结束时间 2006-01-02 15:04:05"
// @Success 200 {object} response.Response
// @Router /api/v1/dtu/{dtu}/records [get]
func FindDtuRecords(c *gin.Context) {
	appG := response.Gin{C: c}

	dtu, ok := dtuParam(c)
	if !ok {
		appG.Response(http.StatusOK, response.InvalidParams, "dtu 编号必须为 1~4")
		return
	}

	now := time.Now()
	start := startOfDay(now)
	end := start.AddDate(0, 0, 1)
	if s, e, ok := timeRange(c); ok {
		start, end = s, e
	}

	pageNum, pageSizeNum := pagination(c)
	recordService := record_service.RecordService{}
	records, total, err := recordService.FindDtuRecords(dtu, model.DtuQuery{
		Page:     pageNum,
		PageSize: pageSizeNum,
		Start:    &start,
		End:      &end,
	})
	if err != nil {
		zap.L().Error("fail to find dtu records", zap.Uint8("dtu", dtu), zap.Error(err))
		appG.Response(http.StatusOK, response.ErrorFindDtuRecords, err)
		return
	}
	appG.PageDataResponse(http.StatusOK, response.SUCCESS, total, pageSizeNum, pageNum, records)
}

// @Summary 历史
// @Tags 获取指定dtu的历史数据（默认最近一周）
// @Accept json
// @Produce  json
// @Param dtu path int true "dtu编号 1~4"
// @Param start query string false "开始时间 2006-01-02 15:04:05"
// @Param end query string false "结束时间 2006-01-02 15:04:05"
// @Success 200 {object} response.Response
// @Router /api/v1/dtu/{dtu}/history [get]
func FindDtuHistory(c *gin.Context) {
	appG := response.Gin{C: c}

	dtu, ok := dtuParam(c)
	if !ok {
		appG.Response(http.StatusOK, response.InvalidParams, "dtu 编号必须为 1~4")
		return
	}

	now := time.Now()
	start := now.AddDate(0, 0, -7)
	end := now
	if s, e, ok := timeRange(c); ok {
		start, end = s, e
	}

	pageNum, pageSizeNum := pagination(c)
	recordService := record_service.RecordService{}
	records, total, err := recordService.FindDtuHistory(dtu, model.DtuQuery{
		Page:     pageNum,
		PageSize: pageSizeNum,
		Start:    &start,
		End:      &end,
	})
	if err != nil {
		zap.L().Error("fail to find dtu history", zap.Uint8("dtu", dtu), zap.Error(err))
		appG.Response(http.StatusOK, response.ErrorFindDtuHistory, err)
		return
	}
	appG.PageDataResponse(http.StatusOK, response.SUCCESS, total, pageSizeNum, pageNum, records)
}

// @Summary 曲线
// @Tags 获取指定dtu单个参数的曲线数据（默认当天）
// @Accept json
// @Produce  json
// @Param dtu path int true "dtu编号 1~4"
// @Param field query string true "字段 temperature/humidity/pm1_0/pm2_5/pm10/tvoc_sample/tvoc_background/pyrolysis"
// @Param start query string false "开始时间 2006-01-02 15:04:05"
// @Param end query string false "结束时间 2006-01-02 15:04:05"
// @Success 200 {object} response.Response
// @Router /api/v1/dtu/{dtu}/curve [get]
func FindDtuCurve(c *gin.Context) {
	appG := response.Gin{C: c}

	dtu, ok := dtuParam(c)
	if !ok {
		appG.Response(http.StatusOK, response.InvalidParams, "dtu 编号必须为 1~4")
		return
	}
	field := c.Query("field")
	if field == "" {
		appG.Response(http.StatusOK, response.InvalidParams, "field不能为空")
		return
	}

	now := time.Now()
	start := startOfDay(now)
	end := start.AddDate(0, 0, 1)
	if s, e, ok := timeRange(c); ok {
		start, end = s, e
	}

	recordService := record_service.RecordService{}
	points, err := recordService.FindDtuCurve(dtu, field, start, end)
	if err != nil {
		zap.L().Error("fail to find dtu curve", zap.Uint8("dtu", dtu), zap.String("field", field), zap.Error(err))
		appG.Response(http.StatusOK, response.ErrorFindDtuCurve, err)
		return
	}
	appG.Response(http.StatusOK, response.SUCCESS, points)
}

// dtuParam 解析路径中的 dtu 编号，仅允许 1~4
func dtuParam(c *gin.Context) (uint8, bool) {
	n, err := strconv.Atoi(c.Param("dtu"))
	if err != nil || n < 1 || n > 4 {
		return 0, false
	}
	return uint8(n), true
}

// pagination 解析分页参数，缺省 page=1、pageSize=10，pageSize 上限 200
func pagination(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.Query("page"))
	pageSize, _ = strconv.Atoi(c.Query("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}

// timeRange 解析 start/end 查询参数；未传时 ok=false，由调用方使用默认范围
func timeRange(c *gin.Context) (start, end time.Time, ok bool) {
	s, e := c.Query("start"), c.Query("end")
	if s == "" && e == "" {
		return time.Time{}, time.Time{}, false
	}
	if s != "" {
		t, err := parseQueryTime(s)
		if err != nil {
			return time.Time{}, time.Time{}, false
		}
		start = t
	}
	if e != "" {
		t, err := parseQueryTime(e)
		if err != nil {
			return time.Time{}, time.Time{}, false
		}
		end = t
	}
	if start.IsZero() {
		start = time.Now().AddDate(0, 0, -1)
	}
	if end.IsZero() {
		end = time.Now()
	}
	return start, end, true
}

func parseQueryTime(s string) (time.Time, error) {
	for _, layout := range []string{queryTimeFormat, "2006-01-02", time.RFC3339} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, strconv.ErrSyntax
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
