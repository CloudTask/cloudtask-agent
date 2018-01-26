package driver

import "github.com/cloudtask/common/models"

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

const ZERO_TICK float64 = 0.0000000

var (
	ErrScheduleInvalid = errors.New("schedule invalid.") //任务计划无效
	ErrScheduleExpired = errors.New("schedule expired.") //任务计划过期
)

var (
	monthsMapping = map[string]time.Month{
		"1":  time.January,
		"2":  time.February,
		"3":  time.March,
		"4":  time.April,
		"5":  time.May,
		"6":  time.June,
		"7":  time.July,
		"8":  time.August,
		"9":  time.September,
		"10": time.October,
		"11": time.November,
		"12": time.December,
	}
	weeksMapping = map[string]time.Weekday{
		"0": time.Sunday,
		"1": time.Monday,
		"2": time.Tuesday,
		"3": time.Wednesday,
		"4": time.Thursday,
		"5": time.Friday,
		"6": time.Saturday,
	}
)

func CalcSchedule(schedule *models.Schedule, seed time.Time) (time.Time, error) {

	switch schedule.TurnMode {
	case models.TURNMODE_SECONDS:
		return CalcInterval(schedule, seed)
	case models.TURNMODE_MINUTES:
		return CalcInterval(schedule, seed)
	case models.TURNMODE_HOURLY:
		return CalcInterval(schedule, seed)
	case models.TURNMODE_DAILY:
		return CalcDaily(schedule, seed)
	case models.TURNMODE_WEEKLY:
		return CalcWeekly(schedule, seed)
	case models.TURNMODE_MONTHLY:
		return CalcMonthly(schedule, seed)
	}
	return time.Time{}, ErrScheduleInvalid
}

func getZoneLocation(t time.Time) *time.Location {

	name, offset := t.Zone()
	return time.FixedZone(name, offset)
}

func getSecondsTail(turnmode int) string {

	var sectail_str string
	if turnmode == models.TURNMODE_SECONDS {
		sectail_str = ":00" //按秒轮询，endtime拼接到:00，超过指定endtime即不能运行
	} else {
		sectail_str = ":59" //按分钟、小时或天轮询, endtime拼接到:59，在00:00~23:59情况时保证59分也可运行
	}
	return sectail_str
}

func getIntervalDuration(turnmode int, interval int) (time.Duration, error) {

	var d int
	switch turnmode {
	case models.TURNMODE_SECONDS:
		d = interval
	case models.TURNMODE_MINUTES:
		d = interval * 60
	case models.TURNMODE_HOURLY:
		d = interval * 3600
	}
	dur, err := time.ParseDuration("+" + strconv.Itoa(d) + "s")
	if err != nil {
		return 0, err
	}
	return dur, nil
}

func isExpired(schedule *models.Schedule, seed time.Time) (bool, error) {

	local := getZoneLocation(seed)
	enddate := strings.TrimSpace(schedule.EndDate)
	if enddate == "" { //无过期时间，永不过期
		return false, nil
	}

	var err error
	var expired time.Time
	if strings.TrimSpace(schedule.EndTime) != "" {
		sectail := getSecondsTail(schedule.TurnMode)
		expired, err = time.ParseInLocation("01/02/2006 15:04:05", schedule.EndDate+" "+schedule.EndTime+sectail, local)
		if err != nil {
			return false, err
		}
	} else {
		expired, err = time.ParseInLocation("01/02/2006 15:04:05", schedule.EndDate+" "+schedule.StartTime+":59", local)
		if err != nil {
			return false, err
		}
	}

	if seed.Sub(expired).Seconds() >= ZERO_TICK {
		return true, nil //计划已过期
	}
	return false, nil
}
