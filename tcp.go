package main

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/David89Lee/pyrolysis_ty_server/common"
	"github.com/David89Lee/pyrolysis_ty_server/common/sse"
	"github.com/David89Lee/pyrolysis_ty_server/model"
	"go.uber.org/zap"
)

// DTU 识别码长度（数据帧前缀，如 AABB558087243440）
const dtuIDLen = 16

// DTU 上报的数据帧格式：
//
//	16 字节识别码 + 原 485 响应帧（地址 + 0x03 + 字节数 + 18 字节数据 + 2 字节 CRC）
//
// startTcpServer 监听 TCP 端口，接收 DTU 主动上报的数据
func startTcpServer(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		zap.L().Error("启动 TCP 服务失败", zap.String("addr", addr), zap.Error(err))
		return
	}
	defer ln.Close()

	zap.L().Info("TCP 服务已启动", zap.String("addr", addr))
	zap.L().Info("DTU 识别码配置",
		zap.String("dtu1", common.DtuSetting.Dtu1),
		zap.String("dtu2", common.DtuSetting.Dtu2),
		zap.String("dtu3", common.DtuSetting.Dtu3),
		zap.String("dtu4", common.DtuSetting.Dtu4))
	for {
		conn, err := ln.Accept()
		if err != nil {
			zap.L().Error("接受 TCP 连接失败", zap.Error(err))
			continue
		}
		zap.L().Info("DTU 已连接", zap.String("remote", conn.RemoteAddr().String()))
		go handleDtuConn(conn)
	}
}

func handleDtuConn(conn net.Conn) {
	defer func() {
		zap.L().Info("DTU 连接断开", zap.String("remote", conn.RemoteAddr().String()))
		_ = conn.Close()
	}()

	buf := make([]byte, 0, 1024)
	chunk := make([]byte, 1024)
	for {
		n, err := conn.Read(chunk)
		if n > 0 {
			buf = append(buf, chunk[:n]...)
			buf = consumeDtuFrames(buf)
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				zap.L().Error("读取 DTU 数据失败",
					zap.String("remote", conn.RemoteAddr().String()), zap.Error(err))
			}
			return
		}
	}
}

// consumeDtuFrames 从缓冲区中解析出完整的 DTU 数据帧，返回未处理的剩余字节
func consumeDtuFrames(buf []byte) []byte {
	zap.L().Info("收到 DTU 数据", zap.Int("len", len(buf)), zap.ByteString("buf", buf))
	for len(buf) >= dtuIDLen+3 {
		dtu, ok := matchDtuID(string(buf[:dtuIDLen]))
		if !ok {
			// 不是已配置的 DTU 识别码，逐字节滑动重新同步
			buf = buf[1:]
			continue
		}
		payload := buf[dtuIDLen:]
		// payload: [从机地址][功能码 0x03][字节数][数据...][CRC]
		if payload[1] != 0x03 || int(payload[2]) != SampleRegCount*2 {
			buf = buf[1:]
			continue
		}
		frameLen := 3 + SampleRegCount*2 + 2
		if len(payload) < frameLen {
			break // 数据不完整，等待后续数据
		}
		sample, err := parseSampleRegisters(payload[3 : 3+SampleRegCount*2])
		if err != nil {
			zap.L().Error("解析 DTU 采样数据失败", zap.Error(err))
		} else {
			handleDtuSample(dtu, sample)
		}
		buf = buf[dtuIDLen+frameLen:]
	}
	return buf
}

// matchDtuID 用识别码匹配配置中的 Dtu1~Dtu4，返回对应编号
func matchDtuID(id string) (uint8, bool) {
	if id == "" {
		return 0, false
	}
	switch id {
	case common.DtuSetting.Dtu1:
		return 1, true
	case common.DtuSetting.Dtu2:
		return 2, true
	case common.DtuSetting.Dtu3:
		return 3, true
	case common.DtuSetting.Dtu4:
		return 4, true
	}
	return 0, false
}

// handleDtuSample 推送 SSE 并写入该 DTU 对应的实时表
func handleDtuSample(dtu uint8, s SampleData) {
	now := time.Now()
	sse.DefaultHub.Publish(s.ToEvent(dtu, now))
	if err := model.CreateDtuSample(dtu, s.ToDtuBase(now)); err != nil {
		zap.L().Error("采样数据落库失败", zap.Uint8("dtu", dtu), zap.Error(err))
		return
	}
	zap.L().Debug("采样数据已落库",
		zap.Uint8("dtu", dtu), zap.String("sample", s.Format()))
}
