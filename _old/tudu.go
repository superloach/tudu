package tudu

import (
	"strconv"
	"time"
)

type ID uint64

const IDNone = ^ID(0)

// Tudu is a task that can be completed.
type Tudu struct {
	ID       ID        `json:"id"`
	Title    string    `json:"title"`
	Done     bool      `json:"done"`
	Children []ID      `json:"children"`
	When     time.Time `json:"when"`
	Repeat   Repeat    `json:"repeat"`
}

// Repeat describes how a task should be repeated.
//
// Given a start date, a number of repeats, and a RepeatType, a repeat will return the
// next date that the task should be repeated on.
//
// For example, given a start date of January 1st, 2020, a type of RepeatTypeDays,
// and an interval of 3, the next date will be January 4th, 2020.
type Repeat struct {
	Start time.Time `json:"start"`

	Type RepeatType `json:"type"`
	Int  int        `json:"int"`

	Until time.Time `json:"until"`
}

type RepeatType uint8

const (
	// RepeatTypeDays adds a given number of days to the date. Given January 31st, 2020,
	// adding 1 day will result in February 1st, 2020.
	RepeatTypeDays RepeatType = iota

	// RepeatTypeWeeks adds a given number of weeks to the date. Given January 31st, 2020,
	// adding 1 week will result in February 7th, 2020.
	RepeatTypeWeeks

	// RepeatTypeMonths adds a given number of months to the date. Given January 31st, 2020,
	// adding 1 month will result in February 29th, 2020.
	RepeatTypeMonths

	// RepeatTypeYears adds a given number of years to the date. Given January 31st, 2020,
	// adding 1 year will result in January 31st, 2021.
	RepeatTypeYears

	// RepeatTypeDaysOfMonth adds a given number of days to the date, skipping months
	// with fewer days. Given January 31st, 2020, adding 1 day will result in February 29th,
	// 2020.
	RepeatTypeDaysOfMonth
)

// LeapYear returns 1 if the given year is a leap year, 0 otherwise.
//
//go:inline
func LeapYear(year int) int {
	switch {
	case year%400 == 0:
		return 1
	case year%100 == 0:
		return 0
	case year%4 == 0:
		return 1
	default:
		return 0
	}
}

// DaysInMonth returns the number of days in the given month.
//
//go:inline
func DaysInMonth(month int, year int) int {
	switch month {
	case 2:
		return 28 + LeapYear(year)
	case 4, 6, 9, 11:
		return 30
	default:
		return 31
	}
}

// Next returns the next date after `today` which falls on an `r.Int` interval.
func (r Repeat) Next(today time.Time) (next time.Time) {
	// truncate `today` to the day
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	// if `today` is before `r.Start`, the next date is `r.Start`
	if today.Before(r.Start) {
		return r.Start
	}

	// if `today` is on or after `r.Until`, there is no next date
	if r.Until != (time.Time{}) && !today.Before(r.Until) {
		return time.Time{}
	}

	// store the interval in a variable so we can modify it later
	interval := r.Int

	switch r.Type {
	case RepeatTypeWeeks:
		interval *= 7
		fallthrough
	case RepeatTypeDays:
		// how many days have passed since `r.Start`?
		days := int(today.Sub(r.Start).Hours() / 24)

		// how many intervals have passed?
		passed := days / interval

		// how many days until the next interval?
		daysToNext := (passed + 1) * interval

		// add that many days to `r.Start`
		return r.Start.AddDate(0, 0, daysToNext)
	default:
		panic("not implemented: unit " + strconv.Itoa(int(r.Type)))
	}
}

func (t Tudu) Next() (Tudu, bool) {
	return Tudu{
		ID:    IDNone,
		Title: t.Title,
		Done:  t.Done,
	}, true
}
