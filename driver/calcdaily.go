package driver

import "github.com/cloudtask/common/models"

import (
	"time"
)

func CalcDaily(schedule *models.Schedule, seed time.Time) (time.Time, error) {

	local := getZoneLocation(seed)
	start, err := time.ParseInLocation("01/02/2006 15:04:05", schedule.StartDate+" "+schedule.StartTime+":00", local)
	if err != nil {
		return time.Time{}, ErrScheduleInvalid //返回无效
	}

	if seed.Sub(start).Seconds() < ZERO_TICK { //计算时间差(相差多少秒，未到开始时间为负数)
		return start, nil //未到开始时间，就返回开始时间
	}

	seed_start, err := time.ParseInLocation("01/02/2006 15:04:05", seed.Format("01/02/2006")+" "+schedule.StartTime+":00", local)
	if err != nil {
		return time.Time{}, ErrScheduleInvalid //返回无效
	}

	var next time.Time
	diffdays := (int)(seed_start.Sub(start).Hours()) / 24
	mod := (diffdays) % schedule.Interval //取模，检查是否在轮询天
	if mod != 0 {
		next = nextTime(start, schedule.Interval, diffdays)
	} else {
		next = seed_start
		if seed.Sub(next).Seconds() > ZERO_TICK { //当前时间已超过，计算到下一个轮询
			next = nextTime(start, schedule.Interval, diffdays)
		}
	}
	return next, nil
}

func nextTime(start time.Time, interval int, diffdays int) time.Time {

	n := diffdays / interval
	days := n*interval + interval
	return start.AddDate(0, 0, days)
}
