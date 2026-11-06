package record_service

import (
	"time"

	"github.com/David89Lee/pyrolysis_ty_server/model"
)

type RecordService struct {
}

func (t *RecordService) CreateInfo(info *model.Info) (*model.Info, error) {
	return model.CreateInfo(info)
}

func (t *RecordService) UpdateInfo(info *model.Info) (*model.Info, error) {
	return model.UpdateInfo(info)
}

func (t *RecordService) DeleteInfo(id uint) error {
	return model.DeleteInfo(id)
}

func (t *RecordService) FindInfo(pageNum, pageSize int) ([]*model.Info, int64, error) {
	return model.FindInfo(pageNum, pageSize)
}

// CreateDtuSample 按 dtu 编号把一条采样写入对应的实时表
func (t *RecordService) CreateDtuSample(dtu uint8, base model.DtuBase) error {
	return model.CreateDtuSample(dtu, base)
}

// FindDtuRecords 分页查询某个 DTU 的实时数据
func (t *RecordService) FindDtuRecords(dtu uint8, q model.DtuQuery) ([]model.DtuBase, int64, error) {
	return model.FindDtuRecords(dtu, q)
}

// FindDtuHistory 分页查询某个 DTU 的历史数据
func (t *RecordService) FindDtuHistory(dtu uint8, q model.DtuQuery) ([]model.DtuBase, int64, error) {
	return model.FindDtuHistory(dtu, q)
}

// FindDtuCurve 查询某个 DTU 某个参数在时间范围内的曲线数据
func (t *RecordService) FindDtuCurve(dtu uint8, field string, start, end time.Time) ([]model.CurvePoint, error) {
	return model.FindDtuCurve(dtu, field, start, end)
}
