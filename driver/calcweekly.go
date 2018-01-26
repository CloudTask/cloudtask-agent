package driver

import "github.com/cloudtask/common/models"

import (
	"strings"
	"time"
)

func CalcWeekly(schedule *models.Schedule, seed time.Time) (time.Time, error) {

	local := getZoneLocation(seed)
	selectat := checkWeeklySelectAt(schedule.SelectAt)
	if len(selectat) == 0 {
		return time.Time{}, ErrScheduleInvalid //返回无效
	}

	start, err := time.ParseInLocation("01/02/2006 15:04:05", schedule.StartDate+" "+schedule.StartTime+":00", local)
	if err != nil {
		return time.Time{}, ErrScheduleInvalid //返回无效
	}

	ret, t := checkIsStart(schedule, selectat, seed, start)
	if !ret {
		return t, nil
	}

	var next time.Time
	t1 := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, local) //开始日期
	r1 := t1.AddDate(0, 0, ^(int)(t1.Weekday())+1)                               //开始周第一天(周日)
	t2 := time.Date(seed.Year(), seed.Month(), seed.Day(), 0, 0, 0, 0, local)    //当前日期
	r2 := t2.AddDate(0, 0, ^(int)(t2.Weekday())+1)                               //当前周第一天(周日)
	diff := r2.Sub(r1).Hours() / 24 / 7                                          //相差了多少周
	mod := (int)(diff) % schedule.Interval                                       //取模检查是否在有效周
	if mod != 0 {
		next = calcNextWeekly(schedule, selectat, int(diff), r1, local) //不在有效周，需要跳到下一个间隔
	} else {
		next = calcCurrWeekly(schedule, selectat, seed, r2, local) //在有效间隔周，计算当周
	}
	return next, nil
}

func checkWeeklySelectAt(selectat string) []int {

	weekdays := make([]int, 0)
	selectat = strings.TrimSpace(selectat)
	items := strings.Split(selectat, ",")
	for i := 0; i < len(items); i++ {
		if weekday, ret := weeksMapping[items[i]]; ret {
			weekdays = append(weekdays, (int)(weekday))
		}
	}
	return weekdays
}

func checkIsStart(schedule *models.Schedule, selectat []int, seed time.Time, start time.Time) (bool, time.Time) {

	start_weekday := (int)(start.Weekday())
	for i := 0; i < len(selectat); i++ {
		if selectat[i] == start_weekday { //检查开始时间是否在selectat中
			if seed.Sub(start).Seconds() < ZERO_TICK { //开始时间还未到，返回开始时间
				return false, start
			}
			return true, time.Time{}
		}
		if selectat[i] > start_weekday { //如果selectat中没有,就找到selectat中第一个大于开始日期(周)
			t := start.AddDate(0, 0, selectat[i]-start_weekday)
			if seed.Sub(t).Seconds() < ZERO_TICK { //开始时间还未到，返回开始时间
				return false, t
			}
			return true, time.Time{}
		}
	}

	start = start.AddDate(0, 0, ^(int)(start.Weekday())+1)
	start = start.AddDate(0, 0, schedule.Interval*7+selectat[0])
	if seed.Sub(start).Seconds() < ZERO_TICK {
		return false, start
	}
	return true, time.Time{}
}

func calcCurrWeekly(schedule *models.Schedule, selectat []int, seed time.Time, start time.Time, local *time.Location) time.Time {

	weekday := (int)(seed.Weekday())
	for i := 0; i < len(selectat); i++ {
		if selectat[i] == weekday {
			seed_start, _ := time.ParseInLocation("01/02/2006 15:04:05", seed.Format("01/02/2006")+" "+schedule.StartTime+":00", local)
			if seed.Sub(seed_start).Seconds() < ZERO_TICK { //当天执行时间还未到，返回下次执行时间
				return seed_start
			}
			break
		}
	}

	var nextday int = -1
	for i := 0; i < len(selectat); i++ {
		if selectat[i] > weekday {
			nextday = selectat[i]
			break
		}
	}

	var next time.Time
	if nextday == -1 { //未找到，跳到下一个周期
		next = start.AddDate(0, 0, schedule.Interval*7+selectat[0])
	} else { //找到，返回下次时间
		next = start.AddDate(0, 0, nextday)
	}
	next, _ = time.ParseInLocation("01/02/2006 15:04:05", next.Format("01/02/2006")+" "+schedule.StartTime+":00", local)
	return next
}

func calcNextWeekly(schedule *models.Schedule, selectdays []int, diff int, start time.Time, local *time.Location) time.Time {

	diff = diff + 1
	p := ((int(diff) / schedule.Interval) * schedule.Interval)
	if p != int(diff) {
		p = p + schedule.Interval + 1
	} else { //当前间隔的最后一周，只需加1周跳到下一个间隔
		p = p + 1
	}

	var next time.Time
	days := ((p - 1) * 7)
	weekday := start.AddDate(0, 0, days)
	weekday = weekday.AddDate(0, 0, ^(int)(weekday.Weekday())+1)
	next = weekday.AddDate(0, 0, selectdays[0])
	next, _ = time.ParseInLocation("01/02/2006 15:04:05", next.Format("01/02/2006")+" "+schedule.StartTime+":00", local)
	return next
}
