package driver

import "github.com/cloudtask/common/models"

import (
	"strconv"
	"time"
)

func CalcInterval(schedule *models.Schedule, seed time.Time) (time.Time, error) {

	local := getZoneLocation(seed)
	start, err := time.ParseInLocation("01/02/2006 15:04:05", schedule.StartDate+" "+schedule.StartTime+":00", local)
	if err != nil {
		return time.Time{}, ErrScheduleInvalid //返回无效
	}

	diffsec := seed.Sub(start).Seconds() //计算时间差(相差多少秒，未到开始时间为负数)
	if diffsec < ZERO_TICK {
		return start, nil //未到开始时间，就返回开始时间
	}

	dur, err := getIntervalDuration(schedule.TurnMode, schedule.Interval)
	if err != nil {
		return time.Time{}, ErrScheduleInvalid //返回无效
	}

	var next time.Time
	diffcount := (int)(diffsec) / (int)(dur.Seconds()) //相差轮询次数
	totalsec := diffcount * (int)(dur.Seconds())       //上次执行总秒数
	totaldur, _ := time.ParseDuration("+" + strconv.Itoa(totalsec) + "s")
	prev := start.Add(totaldur) //上次执行时间
	next = prev.Add(dur)        //下次执行时间

	seed_start, err := time.ParseInLocation("01/02/2006 15:04:05", seed.Format("01/02/2006")+" "+schedule.StartTime+":00", local)
	if err != nil {
		return time.Time{}, ErrScheduleInvalid //返回无效
	}

	if next.Sub(seed_start).Seconds() < ZERO_TICK {
		return seed_start, nil //当天开始时间还未到，返回当天开始时间
	}

	sectail := getSecondsTail(schedule.TurnMode)
	seed_end, err := time.ParseInLocation("01/02/2006 15:04:05", seed.Format("01/02/2006")+" "+schedule.EndTime+sectail, local)
	if err != nil {
		return time.Time{}, ErrScheduleInvalid
	}

	if next.Sub(seed_end).Seconds() >= ZERO_TICK { //下次执行时间若超过了当天的结束时间，则返回第二天的开始
		next = seed_start.AddDate(0, 0, 1) //当天开始时间向后加一天
	}
	return next, nil
}
