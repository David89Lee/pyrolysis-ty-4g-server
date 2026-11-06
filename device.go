package main

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/David89Lee/pyrolysis_ty_server/common/sse"
	"github.com/David89Lee/pyrolysis_ty_server/model"
)

// 一次采样包含的寄存器数量：温度、湿度、PM1.0、PM2.5、PM10、
// TVOC 采样值、TVOC 本底值、热解粒子、状态码，共 9 个（18 字节）
const SampleRegCount = 9

// 状态码位定义
const (
	StatusAlarmMask       uint16 = 1 << 0 // Bit0: 0=未报警,   1=报警
	StatusSensorErrorMask uint16 = 1 << 1 // Bit1: 0=传感器正常, 1=传感器异常
)

// SampleData 保存一次完整采样的解析结果
type SampleData struct {
	Temperature    float64 // °C
	Humidity       float64 // %
	PM1_0          uint16  // μg/m³
	PM2_5          uint16  // μg/m³
	PM10           uint16  // μg/m³
	TVOCSample     uint16  // ppb
	TVOCBackground uint16  // ppb
	Pyrolysis      uint16  // ppb
	StatusCode     uint16  // 原始状态码
}

// SensorNormal 返回传感器是否正常
func (s SampleData) SensorNormal() bool {
	return s.StatusCode&StatusSensorErrorMask == 0
}

// Alarming 返回是否处于报警状态
func (s SampleData) Alarming() bool {
	return s.StatusCode&StatusAlarmMask != 0
}

// parseSampleRegisters 解析 9 个采样寄存器（18 字节，大端）。
// 数据即原 485 从机响应帧中的数据段，例如:
//
//	02 C6 02 60 00 11 00 11 00 18 00 00 00 00 00 00 00 00
//	温度=0x02C6=710 → (710-500)/10=21.0°C，湿度=0x0260=608 → 60.8%
func parseSampleRegisters(data []byte) (SampleData, error) {
	if len(data) < SampleRegCount*2 {
		return SampleData{}, fmt.Errorf("采样数据长度不足: %d", len(data))
	}
	regs := make([]uint16, SampleRegCount)
	for i := 0; i < SampleRegCount; i++ {
		regs[i] = binary.BigEndian.Uint16(data[i*2 : i*2+2])
	}
	// 温度: (寄存器值 - 500) / 10  °C
	temp := float64(int16(regs[0])) - 500
	temp /= 10
	return SampleData{
		Temperature:    temp,
		Humidity:       float64(int16(regs[1])) / 10,
		PM1_0:          regs[2],
		PM2_5:          regs[3],
		PM10:           regs[4],
		TVOCSample:     regs[5],
		TVOCBackground: regs[6],
		Pyrolysis:      regs[7],
		StatusCode:     regs[8],
	}, nil
}

// ToEvent 把采样结果转换成供 SSE 推送给前端的事件负载
func (s SampleData) ToEvent(dtu uint8, ts time.Time) sse.SampleEvent {
	return sse.SampleEvent{
		Dtu:            dtu,
		Timestamp:      ts.Format(time.RFC3339),
		Temperature:    s.Temperature,
		Humidity:       s.Humidity,
		PM1_0:          s.PM1_0,
		PM2_5:          s.PM2_5,
		PM10:           s.PM10,
		TVOCSample:     s.TVOCSample,
		TVOCBackground: s.TVOCBackground,
		Pyrolysis:      s.Pyrolysis,
		StatusCode:     s.StatusCode,
		SensorNormal:   s.SensorNormal(),
		Alarming:       s.Alarming(),
	}
}

// ToDtuBase 把采样结果转换成待落库的记录
func (s SampleData) ToDtuBase(ts time.Time) model.DtuBase {
	return model.DtuBase{
		CreatedAt:      ts,
		Year:           int16(ts.Year()),
		Month:          int8(ts.Month()),
		Day:            int8(ts.Day()),
		Hour:           int8(ts.Hour()),
		Minute:         int8(ts.Minute()),
		Second:         int8(ts.Second()),
		Temperature:    float32(s.Temperature),
		Humidity:       float32(s.Humidity),
		Pm1:            int16(s.PM1_0),
		Pm2dot5:        int16(s.PM2_5),
		Pm10:           int16(s.PM10),
		TvocSampling:   int16(s.TVOCSample),
		TvocBackground: int16(s.TVOCBackground),
		Pyrolysis:      int16(s.Pyrolysis),
		SensorState:    s.SensorNormal(),
		Alarm:          s.Alarming(),
	}
}

// Format 返回一行带单位的人类可读字符串
func (s SampleData) Format() string {
	sensor := "正常"
	if !s.SensorNormal() {
		sensor = "异常"
	}
	alarm := "未报警"
	if s.Alarming() {
		alarm = "报警"
	}
	return fmt.Sprintf(
		"温度=%.1f°C  湿度=%.1f%%  PM1.0=%dμg/m³  PM2.5=%dμg/m³  PM10=%dμg/m³  "+
			"TVOC采样=%dppb  TVOC本底=%dppb  热解粒子=%dppb  状态码=0x%04X(传感器%s,%s)",
		s.Temperature, s.Humidity, s.PM1_0, s.PM2_5, s.PM10,
		s.TVOCSample, s.TVOCBackground, s.Pyrolysis, s.StatusCode, sensor, alarm,
	)
}
