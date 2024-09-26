package atask

import (
	"errors"
	"strconv"
)

const (
	ScheduleTypeCron          ScheduleType = iota // cron
	ScheduleTypeFixedInterval                     // 固定间隔
)

type ScheduleType int

type Schedule struct {
	// 调度类型: cron, fixedInterval
	Type ScheduleType

	// 调度配置
	//   调度类型为cron, 值为 "cron表达式"
	//   调度类型为fixedInterval, 值为 "执行间隔时间, 单位: 秒"
	Conf string
}

// Cron 定时器表达式
func (s Schedule) Cron() (expr string, err error) {
	if s.Type == ScheduleTypeCron {
		expr = s.Conf
	}
	if expr == "" {
		err = errors.New("无效的cron表达式")
		return "", err
	}
	return expr, nil
}

// Interval 执行间隔时间, 单位: 秒
func (s Schedule) Interval() (seconds int64, err error) {
	if s.Type != ScheduleTypeFixedInterval {
		return 0, errors.New("调度类型非固定间隔")
	}
	seconds, err = strconv.ParseInt(s.Conf, 10, 64)
	if err != nil || seconds < 0 {
		return 0, errors.New("无效的间隔时间")
	}
	return
}
