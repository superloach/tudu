package tudu_test

import (
	"testing"
	"time"

	"superloach.xyz/tudu"
)

func TestNext(t *testing.T) {
	for _, test := range []struct {
		name string
		rep  tudu.Repeat
		from time.Time
		want time.Time
	}{{
		name: "2023-09-12_days_1",
		rep: tudu.Repeat{
			Start: time.Date(2023, 9, 12, 0, 0, 0, 0, time.UTC),
			Type:  tudu.RepeatTypeDays,
			Int:   1,
		},
		want: time.Date(2023, 9, 13, 0, 0, 0, 0, time.UTC),
	}, {
		name: "2023-09-12_days_3",
		rep: tudu.Repeat{
			Start: time.Date(2023, 9, 12, 0, 0, 0, 0, time.UTC),
			Type:  tudu.RepeatTypeDays,
			Int:   3,
		},
		want: time.Date(2023, 9, 15, 0, 0, 0, 0, time.UTC),
	}, {
		name: "2023-09-12_days_3_from_2023-09-14",
		rep: tudu.Repeat{
			Start: time.Date(2023, 9, 12, 0, 0, 0, 0, time.UTC),
			Type:  tudu.RepeatTypeDays,
			Int:   3,
		},
		from: time.Date(2023, 9, 14, 0, 0, 0, 0, time.UTC),
		want: time.Date(2023, 9, 15, 0, 0, 0, 0, time.UTC),
	}, {
		name: "2023-09-12_days_3_from_2023-09-15",
		rep: tudu.Repeat{
			Start: time.Date(2023, 9, 12, 0, 0, 0, 0, time.UTC),
			Type:  tudu.RepeatTypeDays,
			Int:   3,
		},
		from: time.Date(2023, 9, 15, 0, 0, 0, 0, time.UTC),
		want: time.Date(2023, 9, 18, 0, 0, 0, 0, time.UTC),
	}, {
		name: "2023-09-12_days_7",
		rep: tudu.Repeat{
			Start: time.Date(2023, 9, 12, 0, 0, 0, 0, time.UTC),
			Type:  tudu.RepeatTypeDays,
			Int:   7,
		},
		want: time.Date(2023, 9, 19, 0, 0, 0, 0, time.UTC),
	}, {
		name: "2023-09-12_weeks_1",
		rep: tudu.Repeat{
			Start: time.Date(2023, 9, 12, 0, 0, 0, 0, time.UTC),
			Type:  tudu.RepeatTypeWeeks,
			Int:   1,
		},
		want: time.Date(2023, 9, 19, 0, 0, 0, 0, time.UTC),
	}, {
		name: "2023-09-12_days_1_from_2023-09-10",
		rep: tudu.Repeat{
			Start: time.Date(2023, 9, 12, 0, 0, 0, 0, time.UTC),
			Type:  tudu.RepeatTypeDays,
			Int:   1,
		},
		from: time.Date(2023, 9, 10, 0, 0, 0, 0, time.UTC),
		want: time.Date(2023, 9, 12, 0, 0, 0, 0, time.UTC),
	}, {
		name: "2023-09-12_days_1_until_2023-09-14_from_2023-09-14",
		rep: tudu.Repeat{
			Start: time.Date(2023, 9, 12, 0, 0, 0, 0, time.UTC),
			Until: time.Date(2023, 9, 14, 0, 0, 0, 0, time.UTC),
			Type:  tudu.RepeatTypeDays,
			Int:   1,
		},
		from: time.Date(2023, 9, 14, 0, 0, 0, 0, time.UTC),
		want: time.Time{},
	}, {
		name: "2023-09-12_days_1_until_2023-09-14_from_2023-09-15",
		rep: tudu.Repeat{
			Start: time.Date(2023, 9, 12, 0, 0, 0, 0, time.UTC),
			Until: time.Date(2023, 9, 14, 0, 0, 0, 0, time.UTC),
			Type:  tudu.RepeatTypeDays,
			Int:   1,
		},
		from: time.Date(2023, 9, 15, 0, 0, 0, 0, time.UTC),
		want: time.Time{},
	}} {
		t.Run(test.name, func(t *testing.T) {
			if test.from == (time.Time{}) {
				test.from = test.rep.Start
			}

			got := test.rep.Next(test.from)
			if !test.want.Equal(got) {
				t.Errorf("want %v, got %v", test.want, got)
			}
		})
	}
}
