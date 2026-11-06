package model

import (
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

type Info struct {
	gorm.Model
	Sn          string `Gorm:"type:char(16); comment:'sn(区分dtu)'"`
	Description string `Gorm:"type:varchar(255); comment:'描述'"`
}

// DtuBase 各 DTU 实时表与历史表的公共字段
type DtuBase struct {
	ID             uint `gorm:"primary_key"`
	CreatedAt      time.Time `gorm:"index"`
	Year           int16   `gorm:"type:smallint comment '年份';index:idx_date"`
	Month          int8    `gorm:"type:tinyint comment '月份';index:idx_date"`
	Day            int8    `gorm:"type:tinyint comment '日期';index:idx_date"`
	Hour           int8    `gorm:"type:tinyint comment '小时';index:idx_date"`
	Minute         int8    `gorm:"type:tinyint comment '分钟';index:idx_date"`
	Second         int8    `gorm:"type:tinyint comment '秒钟';index:idx_date"`
	Temperature    float32 `Gorm:"type:float; comment '温度 °C'"`
	Humidity       float32 `Gorm:"type:float; comment '湿度 %'"`
	Pm1            int16   `Gorm:"type:smallint; comment:'pm1.0 μg/m³'"`
	Pm2dot5        int16   `Gorm:"type:smallint; comment:'pm2.5 μg/m³'"`
	Pm10           int16   `Gorm:"type:smallint; comment:'pm10 μg/m³'"`
	TvocSampling   int16   `Gorm:"type:smallint; comment:'tvoc采样值 ppb'"`
	TvocBackground int16   `Gorm:"type:smallint; comment:'tvoc本底值 ppb'"`
	Pyrolysis      int16   `Gorm:"type:smallint; comment:'热解粒子值 ppb'"`
	SensorState    bool    `Gorm:" comment:'传感器状态'"`
	Alarm          bool    `Gorm:" comment:'是否报警'"`
}

type Dtu1 struct {
	DtuBase
}
type Dtu2 struct {
	DtuBase
}
type Dtu3 struct {
	DtuBase
}
type Dtu4 struct {
	DtuBase
}

type Dtu1His struct {
	DtuBase
}
type Dtu2His struct {
	DtuBase
}
type Dtu3His struct {
	DtuBase
}
type Dtu4His struct {
	DtuBase
}

// 表名固定为 Dtu1~Dtu4 / Dtu1His~Dtu4His（MySQL 在 Windows 下大小写不敏感）
func (Dtu1) TableName() string    { return "dtu1" }
func (Dtu2) TableName() string    { return "dtu2" }
func (Dtu3) TableName() string    { return "dtu3" }
func (Dtu4) TableName() string    { return "dtu4" }
func (Dtu1His) TableName() string { return "dtu1_his" }
func (Dtu2His) TableName() string { return "dtu2_his" }
func (Dtu3His) TableName() string { return "dtu3_his" }
func (Dtu4His) TableName() string { return "dtu4_his" }

// DtuTable 返回 dtu 编号（1~4）对应的实时表与历史表名
func DtuTable(dtu uint8) (table, hisTable string, err error) {
	switch dtu {
	case 1:
		return "dtu1", "dtu1_his", nil
	case 2:
		return "dtu2", "dtu2_his", nil
	case 3:
		return "dtu3", "dtu3_his", nil
	case 4:
		return "dtu4", "dtu4_his", nil
	}
	return "", "", fmt.Errorf("非法的 dtu 编号: %d", dtu)
}

// DtuQuery 分页 + 时间范围查询条件，Start/End 为 nil 表示不限
type DtuQuery struct {
	Page     int
	PageSize int
	Start    *time.Time
	End      *time.Time
}

// CurvePoint 曲线上的一个点
type CurvePoint struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

func CreateInfo(info *Info) (*Info, error) {
	result := DB.Create(info)
	return info, result.Error
}

func UpdateInfo(info *Info) (*Info, error) {
	result := DB.Model(&Info{}).Where("id = ?", info.ID).
		Select("Sn", "Description").Updates(info)
	return info, result.Error
}

func DeleteInfo(id uint) error {
	return DB.Delete(&Info{}, id).Error
}

func FindInfo(pageNum, pageSize int) ([]*Info, int64, error) {
	db := DB.Model(&Info{})
	var total int64
	db.Count(&total)
	var info []*Info
	err := db.Order("id desc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&info).Error
	return info, total, err
}

// CreateDtuSample 按 dtu 编号把一条采样写入对应的实时表
func CreateDtuSample(dtu uint8, base DtuBase) error {
	table, _, err := DtuTable(dtu)
	if err != nil {
		return err
	}
	return DB.Table(table).Create(&base).Error
}

// FindDtuRecords 分页查询实时表数据（按时间倒序）
func FindDtuRecords(dtu uint8, q DtuQuery) ([]DtuBase, int64, error) {
	table, _, err := DtuTable(dtu)
	if err != nil {
		return nil, 0, err
	}
	return findDtu(table, q)
}

// FindDtuHistory 分页查询历史表数据（按时间倒序）
func FindDtuHistory(dtu uint8, q DtuQuery) ([]DtuBase, int64, error) {
	_, hisTable, err := DtuTable(dtu)
	if err != nil {
		return nil, 0, err
	}
	return findDtu(hisTable, q)
}

func findDtu(table string, q DtuQuery) ([]DtuBase, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 10
	}
	db := DB.Table(table)
	if q.Start != nil {
		db = db.Where("created_at >= ?", *q.Start)
	}
	if q.End != nil {
		db = db.Where("created_at <= ?", *q.End)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]DtuBase, 0, q.PageSize)
	err := db.Order("created_at desc").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&rows).Error
	return rows, total, err
}

// FindDtuCurve 查询某个传感器参数在时间范围内的曲线数据（实时表，按时间升序）
func FindDtuCurve(dtu uint8, field string, start, end time.Time) ([]CurvePoint, error) {
	table, _, err := DtuTable(dtu)
	if err != nil {
		return nil, err
	}
	var rows []DtuBase
	err = DB.Table(table).
		Where("created_at >= ? AND created_at <= ?", start, end).
		Order("created_at asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	points := make([]CurvePoint, 0, len(rows))
	for _, r := range rows {
		v, ok := dtuFieldValue(r, field)
		if !ok {
			return nil, fmt.Errorf("不支持的曲线字段: %s", field)
		}
		points = append(points, CurvePoint{
			Time:  r.CreatedAt.Format("2006-01-02 15:04:05"),
			Value: v,
		})
	}
	return points, nil
}

func dtuFieldValue(r DtuBase, field string) (float64, bool) {
	switch field {
	case "temperature":
		return float64(r.Temperature), true
	case "humidity":
		return float64(r.Humidity), true
	case "pm1_0":
		return float64(r.Pm1), true
	case "pm2_5":
		return float64(r.Pm2dot5), true
	case "pm10":
		return float64(r.Pm10), true
	case "tvoc_sample":
		return float64(r.TvocSampling), true
	case "tvoc_background":
		return float64(r.TvocBackground), true
	case "pyrolysis":
		return float64(r.Pyrolysis), true
	}
	return 0, false
}

// AggregateDtuHistory 把实时表中 cutoff 之前的数据抽稀为 10 分钟一条写入历史表，
// 然后删除实时表中这些数据。返回写入历史表的条数和删除的条数。
func AggregateDtuHistory(dtu uint8, cutoff time.Time) (inserted int, deleted int64, err error) {
	table, hisTable, err := DtuTable(dtu)
	if err != nil {
		return 0, 0, err
	}

	const batchSize = 5000
	// 每个 10 分钟桶保留最早的一条
	buckets := make(map[string]DtuBase)
	var lastID uint
	for {
		var rows []DtuBase
		err = DB.Table(table).
			Where("created_at < ? AND id > ?", cutoff, lastID).
			Order("id asc").
			Limit(batchSize).
			Find(&rows).Error
		if err != nil {
			return 0, 0, err
		}
		if len(rows) == 0 {
			break
		}
		for _, r := range rows {
			key := r.CreatedAt.Format("2006010215") + fmt.Sprintf("%02d", r.CreatedAt.Minute()/10)
			if _, ok := buckets[key]; !ok {
				b := r
				b.ID = 0 // 历史表自增主键，避免与实时表 id 冲突
				b.CreatedAt = time.Date(r.CreatedAt.Year(), r.CreatedAt.Month(), r.CreatedAt.Day(),
					r.CreatedAt.Hour(), (r.CreatedAt.Minute()/10)*10, 0, 0, r.CreatedAt.Location())
				b.Year = int16(b.CreatedAt.Year())
				b.Month = int8(b.CreatedAt.Month())
				b.Day = int8(b.CreatedAt.Day())
				b.Hour = int8(b.CreatedAt.Hour())
				b.Minute = int8(b.CreatedAt.Minute())
				b.Second = int8(b.CreatedAt.Second())
				buckets[key] = b
			}
		}
		lastID = rows[len(rows)-1].ID
	}

	if len(buckets) > 0 {
		list := make([]DtuBase, 0, len(buckets))
		for _, b := range buckets {
			list = append(list, b)
		}
		sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.Before(list[j].CreatedAt) })
		if err = DB.Table(hisTable).CreateInBatches(&list, 200).Error; err != nil {
			return 0, 0, err
		}
		inserted = len(list)
	}

	res := DB.Exec(fmt.Sprintf("DELETE FROM %s WHERE created_at < ?", table), cutoff)
	if res.Error != nil {
		return inserted, 0, res.Error
	}
	return inserted, res.RowsAffected, nil
}
