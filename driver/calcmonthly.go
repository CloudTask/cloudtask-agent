package driver

import "github.com/cloudtask/common/models"

import (
	"strconv"
	"strings"
	"time"
)

func CalcMonthly(schedule *models.Schedule, seed time.Time) (time.Time, error) {

	local := getZoneLocation(seed)
	selectat := checkMonthlySelectAt(schedule.SelectAt)
	if len(selectat) == 0 {
		return time.Time{}, ErrScheduleInvalid
	}

	start, err := time.ParseInLocation("01/02/2006 15:04:05", schedule.StartDate+" "+schedule.StartTime+":00", local)
	if err != nil {
		return time.Time{}, ErrScheduleInvalid
	}

	if seed.Sub(start).Seconds() > ZERO_TICK {
		start, _ = time.ParseInLocation("01/02/2006 15:04:05", seed.Format("01/02/2006")+" "+schedule.StartTime+":00", local)
		if seed.Sub(start).Seconds() > ZERO_TICK {
			start = start.AddDate(0, 0, 1)
		}
	}

	var next time.Time
	if schedule.MonthlyOf.Day == 0 {
		next = calcMonthlyOfWeek(schedule.MonthlyOf.Week, selectat, start, local)
	} else {
		next = calcMonthlyOfDay(schedule.MonthlyOf.Day, selectat, start, local)
	}
	if next.IsZero() {
		return time.Time{}, ErrScheduleInvalid
	}
	return next, nil
}

func checkMonthlySelectAt(selectat string) []time.Month {

	months := make([]time.Month, 0)
	selectat = strings.TrimSpace(selectat)
	items := strings.Split(selectat, ",")
	for i := 0; i < len(items); i++ {
		if month, ret := monthsMapping[items[i]]; ret {
			months = append(months, month)
		}
	}
	return months
}

func isIncludeMonth(t time.Time, selectat []time.Month) bool {

	for i := 0; i < len(selectat); i++ {
		if t.Month() == selectat[i] {
			return true
		}
	}
	return false
}

func calcMonthlyOfDay(day int, selectat []time.Month, start time.Time, local *time.Location) time.Time {

	next := start
	for {
		if day > 0 {
			if next.Day() == day && isIncludeMonth(next, selectat) {
				break
			}
		} else {
			nextmonth := next.AddDate(0, 1, 0)
			at := time.Date(nextmonth.Year(), nextmonth.Month(), 1, 0, 0, 0, 0, local).AddDate(0, 0, day)
			if at.Day() == next.Day() && isIncludeMonth(next, selectat) {
				break
			}
		}
		next = next.AddDate(0, 0, 1)
	}
	return next
}

func calcMonthlyOfWeek(week string, selectat []time.Month, start time.Time, local *time.Location) time.Time {

	oweek := strings.SplitN(week, ":", 2)
	if len(oweek) != 2 {
		return time.Time{}
	}

	which, err := strconv.Atoi(oweek[0])
	if err != nil {
		return time.Time{}
	}

	w, ret := weeksMapping[oweek[1]]
	if !ret {
		return time.Time{}
	}

	weekday := w
	next := start
	for {
		if !isWeekDay(next, which, weekday, local) || !isIncludeMonth(next, selectat) {
			next = next.AddDate(0, 0, 1)
			continue
		}
		break
	}
	return next
}

func isWeekDay(next time.Time, which int, weekday time.Weekday, local *time.Location) bool {

	if weekday == next.Weekday() {
		if which == 0 {
			qty1 := dayOfWeekQty(next.Year(), next.Month(), weekday, local)
			qty2 := weekQty(next, weekday, local)
			if qty1 == qty2 {
				return true
			}
		} else {
			qty := weekQty(next, weekday, local)
			if qty == which {
				return true
			}
		}
	}
	return false
}

func weekQty(day time.Time, weekday time.Weekday, local *time.Location) int {

	qty := 0
	for i := 1; i <= day.Day(); i++ {
		if time.Date(day.Year(), day.Month(), i, 0, 0, 0, 0, local).Weekday() == weekday {
			qty++
		}
	}
	return qty
}

func dayOfWeekQty(year int, month time.Month, weekday time.Weekday, local *time.Location) int {

	qty := 0
	for i := 1; i <= daysInMonth(year, month); i++ {
		if time.Date(year, month, i, 0, 0, 0, 0, local).Weekday() == weekday {
			qty++
		}
	}
	return qty
}

func daysInMonth(year int, month time.Month) (days int) {

	if month != time.February {
		if month == time.April ||
			month == time.June ||
			month == time.September ||
			month == time.November {
			days = 30

		} else {
			days = 31
		}
	} else {
		if ((year%4) == 0 && (year%100) != 0) || (year%400) == 0 {
			days = 29
		} else {
			days = 28
		}
	}
	return
}
