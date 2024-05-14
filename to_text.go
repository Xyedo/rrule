package rrule

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type byweekday struct {
	allWeeks   []Weekday
	someWeeks  []Weekday
	isWeekdays bool
	isEveryDay bool
}

type TimeFormatter interface {
	MonthName(int) string
	Format(time.Time) string
	Nth(i int) string
	WeekDayName(Weekday) string
}

type defaultFormatter struct{}

var _ TimeFormatter = defaultFormatter{}

// Format implements TimeFormatter.
func (d defaultFormatter) Format(t time.Time) string {
	return fmt.Sprintf("%s %d, %d", t.Month().String(), t.Day(), t.Year())
}

// MonthName implements TimeFormatter.
func (d defaultFormatter) MonthName(i int) string {
	return [...]string{
		"January",
		"February",
		"March",
		"April",
		"May",
		"June",
		"July",
		"August",
		"September",
		"October",
		"November",
		"December",
	}[i-1]
}

// Nth implements TimeFormatter.
func (d defaultFormatter) Nth(i int) string {
	if i == -1 {
		return "last"
	}

	npos := abs(i)
	var nth strings.Builder
	switch npos {
	case 1, 21, 31:
		nth.WriteString(strconv.Itoa(npos) + "st")
	case 2, 22:
		nth.WriteString(strconv.Itoa(npos) + "nd")
	case 3, 23:
		nth.WriteString(strconv.Itoa(npos) + "rd")
	default:
		nth.WriteString(strconv.Itoa(npos) + "th")
	}

	if i < 0 {
		nth.WriteString(" last")
		return nth.String()
	}

	return nth.String()

}

// WeekDayName implements TimeFormatter.
func (d defaultFormatter) WeekDayName(w Weekday) string {
	weekday := [...]string{
		"Monday",
		"Tuesday",
		"Wednesday",
		"Thursday",
		"Friday",
		"Saturday",
		"Sunday",
	}[w.Day()]

	if n := w.N(); n != 0 {
		return d.Nth(n) + " " + weekday
	}

	return weekday
}

type toText struct {
	bymonthday []int
	byweekday  *byweekday
	option     *ROption
	origOption *ROption
	loc        *i18n.Localizer
	formatter  TimeFormatter
}

var (
	langAnd = &i18n.LocalizeConfig{DefaultMessage: &i18n.Message{
		ID:          "And",
		Description: "Used for final delimiter in list",
		Other:       "and",
	}}

	langOr = &i18n.LocalizeConfig{DefaultMessage: &i18n.Message{
		ID:          "Or",
		Description: "Used for final delimiter in list",
		Other:       "or",
	}}

	langOnThe = &i18n.LocalizeConfig{DefaultMessage: &i18n.Message{
		ID:          "OnThe",
		Description: "Used before list bymonthday, byweekday, byyearday and someWeeks",
		Other:       "on the",
	}}

	langIn = &i18n.LocalizeConfig{DefaultMessage: &i18n.Message{
		ID:    "In",
		Other: "in",
	}}
)

func newToText(rule *RRule, loc *i18n.Localizer, formatter TimeFormatter) *toText {
	var byMonthDay []int
	if len(rule.OrigOptions.Bymonthday) > 0 {
		pos := make([]int, 0, len(rule.Options.Bymonthday))
		neg := make([]int, 0, len(rule.Options.Bymonthday))

		for _, monthDay := range rule.OrigOptions.Bymonthday {
			if monthDay > 0 {
				pos = append(pos, monthDay)
			} else {
				neg = append(neg, monthDay)
			}
		}

		sort.Slice(pos, func(i, j int) bool { return pos[i] < pos[j] })
		sort.Slice(neg, func(i, j int) bool { return neg[j] < neg[i] })

		byMonthDay = append(pos, neg...)
	}

	if len(rule.OrigOptions.Byweekday) > 0 {

		allWeeks := make([]Weekday, 0, len(rule.OrigOptions.Byweekday))
		someWeeks := make([]Weekday, 0, len(rule.OrigOptions.Byweekday))
		for _, weekday := range rule.OrigOptions.Byweekday {
			if weekday.N() == 0 {
				allWeeks = append(allWeeks, weekday)
			}

			if weekday.N() != 0 {
				someWeeks = append(someWeeks, weekday)
			}
		}

		isWeekDays :=
			indexOfWeekDay(rule.Options.Byweekday, MO) != -1 &&
				indexOfWeekDay(rule.Options.Byweekday, TU) != -1 &&
				indexOfWeekDay(rule.Options.Byweekday, WE) != -1 &&
				indexOfWeekDay(rule.Options.Byweekday, TH) != -1 &&
				indexOfWeekDay(rule.Options.Byweekday, FR) != -1 &&
				indexOfWeekDay(rule.Options.Byweekday, SA) == -1 &&
				indexOfWeekDay(rule.Options.Byweekday, SU) == -1

		isEveryDay :=
			indexOfWeekDay(rule.Options.Byweekday, MO) != -1 &&
				indexOfWeekDay(rule.Options.Byweekday, TU) != -1 &&
				indexOfWeekDay(rule.Options.Byweekday, WE) != -1 &&
				indexOfWeekDay(rule.Options.Byweekday, TH) != -1 &&
				indexOfWeekDay(rule.Options.Byweekday, FR) != -1 &&
				indexOfWeekDay(rule.Options.Byweekday, SA) != -1 &&
				indexOfWeekDay(rule.Options.Byweekday, SU) != -1

		sort.Sort(weekDays(allWeeks))
		sort.Sort(weekDays(someWeeks))

		return &toText{
			bymonthday: byMonthDay,
			byweekday: &byweekday{
				allWeeks:   allWeeks,
				someWeeks:  someWeeks,
				isWeekdays: isWeekDays,
				isEveryDay: isEveryDay,
			},
			option:     &rule.Options,
			origOption: &rule.OrigOptions,
			loc:        loc,
			formatter:  formatter,
		}
	}

	return &toText{
		bymonthday: byMonthDay,
		byweekday:  nil,
		option:     &rule.Options,
		origOption: &rule.OrigOptions,
		loc:        loc,
		formatter:  formatter,
	}
}

func (t *toText) ToString() string {
	var text strings.Builder
	text.WriteString(
		t.loc.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Every",
				Other: "every",
			},
		}))

	switch t.option.Freq {
	case MINUTELY:
		t.minutely(&text)
	case HOURLY:
		t.hourly(&text)
	case DAILY:
		t.daily(&text)
	case WEEKLY:
		t.weekly(&text)
	case MONTHLY:
		t.monthly(&text)
	case YEARLY:
		t.yearly(&text)
	}
	if !t.option.Until.IsZero() {
		text.WriteString(" ")
		text.WriteString(
			t.loc.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "Until",
					Other: "until",
				},
			}))
		text.WriteString(" ")
		text.WriteString(t.formatter.Format(t.option.Until))

	} else if t.option.Count > 0 {
		text.WriteString(" ")
		text.WriteString(
			t.loc.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "TimeCount",
					One:   "for {{.Count}} time",
					Two:   "for {{.Count}} times",
					Few:   "for {{.Count}} times",
					Many:  "for {{.Count}} times",
					Other: "for {{.Count}} times",
				},
				TemplateData: map[string]interface{}{
					"Count": t.option.Count,
				},
				PluralCount: t.option.Count,
			}),
		)
	}

	return text.String()
}

func (t *toText) hourly(sb *strings.Builder) {
	if t.option.Interval != 1 {
		sb.WriteByte(' ')
		sb.WriteString(strconv.Itoa(t.option.Interval))
	}

	sb.WriteByte(' ')
	sb.WriteString(
		t.loc.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "HourlyHours",
				One:   "hour",
				Two:   "hours",
				Few:   "hours",
				Many:  "hours",
				Other: "hours",
			},
			PluralCount: t.option.Interval,
		}),
	)
}

func (t *toText) minutely(sb *strings.Builder) {
	if t.option.Interval != 1 {
		sb.WriteByte(' ')
		sb.WriteString(strconv.Itoa(t.option.Interval))
	}

	sb.WriteByte(' ')
	sb.WriteString(
		t.loc.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "MinutelyMinutes",
				One:   "minute",
				Two:   "minutes",
				Few:   "minutes",
				Many:  "minutes",
				Other: "minutes",
			},
			PluralCount: t.option.Interval,
		}),
	)
}

func (t *toText) yearly(sb *strings.Builder) {
	if len(t.origOption.Bymonth) > 0 {
		if t.option.Interval != 1 {
			sb.WriteByte(' ')
			sb.WriteString(strconv.Itoa(t.option.Interval))

			sb.WriteByte(' ')
			sb.WriteString(t.loc.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "YearlyYears",
					One:   "year",
					Two:   "years",
					Few:   "years",
					Many:  "years",
					Other: "years",
				},
				PluralCount: t.option.Interval,
			}))
		}

		t.byMonth(sb)
	} else {
		if t.option.Interval != 1 {
			sb.WriteByte(' ')
			sb.WriteString(strconv.Itoa(t.option.Interval))
		}

		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YearlyYears",
				One:   "year",
				Two:   "years",
				Few:   "years",
				Many:  "years",
				Other: "years",
			},
			PluralCount: t.option.Interval,
		}))
	}

	if len(t.bymonthday) > 0 {
		t.byMonthDay(sb)
	} else if t.byweekday != nil {
		t.byWeekDay(sb)
	}

	if len(t.option.Byyearday) > 0 {
		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(langOnThe))

		sb.WriteByte(' ')
		sb.WriteString(
			t.list(
				t.option.Byyearday,
				t.formatter.Nth,
				t.loc.MustLocalize(langAnd),
				",",
			),
		)

		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YearlyDays",
				Other: "day",
			},
		}))
	}

	if len(t.option.Byweekno) > 0 {
		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(langIn))

		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YearlyWeeks",
				One:   "week",
				Two:   "weeks",
				Few:   "weeks",
				Many:  "weeks",
				Other: "weeks",
			},
			PluralCount: len(t.option.Byweekno),
		}))

		sb.WriteByte(' ')
		sb.WriteString(
			t.list(
				t.option.Byweekno,
				func(i int) string {
					return strconv.Itoa(i)
				},
				t.loc.MustLocalize(langAnd),
				",",
			),
		)
	}
}
func (t *toText) monthly(sb *strings.Builder) {
	if len(t.origOption.Bymonth) > 0 {
		if t.option.Interval != 1 {
			sb.WriteByte(' ')
			sb.WriteString(strconv.Itoa(t.option.Interval))

			sb.WriteByte(' ')
			sb.WriteString(t.loc.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "MonthlyMonths",
					One:   "month",
					Two:   "months",
					Few:   "months",
					Many:  "months",
					Other: "months",
				},
				PluralCount: len(t.origOption.Bymonth),
			}))

			if t.option.Interval > 1 {
				sb.WriteByte(' ')
				sb.WriteString(t.loc.MustLocalize(langIn))
			}
		}

		t.byMonth(sb)

	} else {
		if t.option.Interval != 1 {
			sb.WriteByte(' ')
			sb.WriteString(strconv.Itoa(t.option.Interval))
		}

		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "MonthlyMonths",
				One:   "month",
				Two:   "months",
				Few:   "months",
				Many:  "months",
				Other: "months",
			},
			PluralCount: t.option.Interval,
		}))
	}

	if len(t.bymonthday) > 0 {
		t.byMonthDay(sb)
	} else if t.byweekday != nil && t.byweekday.isWeekdays {
		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "OnMonthly",
				Other: "on",
			},
		}))

		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "MonthlyWeekdays",
				Other: "weekdays",
			},
		}))
	} else if t.byweekday != nil {
		t.byWeekDay(sb)
	}
}
func (t *toText) weekly(sb *strings.Builder) {
	if t.option.Interval != 1 {
		sb.WriteByte(' ')
		sb.WriteString(strconv.Itoa(t.option.Interval))

		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "WeeklyWeeks",
				One:   "week",
				Two:   "weeks",
				Few:   "weeks",
				Many:  "weeks",
				Other: "weeks",
			},
			PluralCount: t.option.Interval,
		}))
	}

	if t.byweekday != nil && t.byweekday.isWeekdays {
		if t.option.Interval == 1 {
			sb.WriteByte(' ')
			sb.WriteString(
				t.loc.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "Weekdays",
						One:   "weekday",
						Two:   "weekdays",
						Few:   "weekdays",
						Many:  "weekdays",
						Other: "weekdays",
					},
					PluralCount: t.option.Interval,
				}),
			)
		} else {
			sb.WriteByte(' ')
			sb.WriteString(
				t.loc.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "OnWeekly",
						Other: "on",
					},
				}),
			)

			sb.WriteByte(' ')
			sb.WriteString(
				t.loc.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "Weekdays",
						One:   "weekday",
						Two:   "weekdays",
						Few:   "weekdays",
						Many:  "weekdays",
						Other: "weekdays",
					},
					PluralCount: t.option.Interval,
				}),
			)

		}
	} else if t.byweekday != nil && t.byweekday.isEveryDay {
		sb.WriteByte(' ')
		sb.WriteString(
			t.loc.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "WeeklyDay",
					One:   "day",
					Two:   "days",
					Few:   "days",
					Many:  "days",
					Other: "days",
				},
				PluralCount: t.option.Interval,
			}),
		)
	} else {
		if t.option.Interval == 1 {
			sb.WriteByte(' ')
			sb.WriteString(
				t.loc.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "WeeklyWeeks",
						One:   "week",
						Two:   "weeks",
						Few:   "weeks",
						Many:  "weeks",
						Other: "weeks",
					},
					PluralCount: t.option.Interval,
				}),
			)
		}

		if len(t.origOption.Bymonth) > 0 {
			sb.WriteByte(' ')
			sb.WriteString(t.loc.MustLocalize(langIn))

			t.byMonth(sb)
		}

		if len(t.bymonthday) > 0 {
			t.byMonthDay(sb)
		} else if t.byweekday != nil {
			t.byWeekDay(sb)
		}

		if len(t.origOption.Byhour) > 0 {
			t.byHour(sb)
		}
	}

}
func (t *toText) daily(sb *strings.Builder) {
	if t.option.Interval != 1 {
		sb.WriteByte(' ')
		sb.WriteString(strconv.Itoa(t.option.Interval))
	}

	if t.byweekday != nil && t.byweekday.isWeekdays {
		sb.WriteByte(' ')
		sb.WriteString(
			t.loc.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "Weekdays",
					One:   "weekday",
					Two:   "weekdays",
					Few:   "weekdays",
					Many:  "weekdays",
					Other: "weekdays",
				},
				PluralCount: t.option.Interval,
			}),
		)

	} else {
		sb.WriteByte(' ')
		sb.WriteString(
			t.loc.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "DailyDay",
					One:   "day",
					Two:   "days",
					Few:   "days",
					Many:  "days",
					Other: "days",
				},
				PluralCount: t.option.Interval,
			}),
		)
	}

	if len(t.origOption.Bymonth) > 0 {
		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(langIn))

		t.byMonth(sb)
	}

	if len(t.bymonthday) > 0 {
		t.byMonthDay(sb)
	} else if t.byweekday != nil {
		t.byWeekDay(sb)
	} else if len(t.origOption.Byhour) > 0 {
		t.byHour(sb)
	}
}

func (t *toText) byHour(sb *strings.Builder) {
	sb.WriteByte(' ')
	sb.WriteString(
		t.loc.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "AtHour",
				Other: "at",
			},
		}))

	sb.WriteByte(' ')
	sb.WriteString(
		t.list(
			t.origOption.Byhour,
			func(i int) string {
				return strconv.Itoa(i)
			},
			t.loc.MustLocalize(langAnd),
			",",
		),
	)

}
func (t *toText) byMonthDay(sb *strings.Builder) {
	if t.byweekday != nil && len(t.byweekday.allWeeks) > 0 {
		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "OnAllWeeks",
				Other: "on",
			},
		}))

		sb.WriteByte(' ')
		sb.WriteString(
			t.listWeekDay(
				t.byweekday.allWeeks,
				t.formatter.WeekDayName,
				t.loc.MustLocalize(langOr),
				","),
		)

		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "TheAllWeeks",
				Other: "the",
			},
		}))

		sb.WriteByte(' ')
		sb.WriteString(
			t.list(t.bymonthday, t.formatter.Nth,
				t.loc.MustLocalize(langOr),
				","),
		)
	} else {
		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(langOnThe))

		sb.WriteByte(' ')
		sb.WriteString(
			t.list(t.bymonthday, t.formatter.Nth,
				t.loc.MustLocalize(langAnd),
				","),
		)
	}
}

func (t *toText) byWeekDay(sb *strings.Builder) {
	if len(t.byweekday.allWeeks) > 0 && !t.byweekday.isWeekdays {
		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "OnWeekDays",
				Other: "on",
			},
		}))

		sb.WriteByte(' ')
		sb.WriteString(
			t.listWeekDay(
				t.byweekday.allWeeks,
				t.formatter.WeekDayName,
				"",
				","),
		)
	}

	if len(t.byweekday.someWeeks) > 0 {
		if len(t.byweekday.allWeeks) > 0 {
			sb.WriteByte(' ')
			sb.WriteString(t.loc.MustLocalize(langAnd))
		}

		sb.WriteByte(' ')
		sb.WriteString(t.loc.MustLocalize(langOnThe))

		sb.WriteByte(' ')
		sb.WriteString(
			t.listWeekDay(
				t.byweekday.someWeeks,
				t.formatter.WeekDayName,
				t.loc.MustLocalize(langAnd),
				",",
			),
		)
	}
}
func (t *toText) byMonth(sb *strings.Builder) {
	sb.WriteByte(' ')
	sb.WriteString(
		t.list(
			t.option.Bymonth,
			t.formatter.MonthName,
			t.loc.MustLocalize(langAnd),
			",",
		),
	)

}

func (*toText) list(numbers []int, callback func(int) string, finalDelim, delim string) string {
	delimJoin := func(arr []string, delimiter, finalDelimiter string) string {
		sb := strings.Builder{}
		for i := 0; i < len(arr); i++ {
			if i != 0 {
				if i == len(arr)-1 {
					sb.WriteByte(' ')
					sb.WriteString(finalDelimiter)
					sb.WriteByte(' ')
				} else {
					sb.WriteString(delimiter)
					sb.WriteByte(' ')
				}
			}
			sb.WriteString(arr[i])
		}

		return sb.String()
	}

	cbRes := make([]string, 0, len(numbers))
	for _, num := range numbers {
		cbRes = append(cbRes, callback(num))
	}

	if finalDelim != "" {
		return delimJoin(cbRes, delim, finalDelim)
	}

	return strings.Join(cbRes, delim+" ")

}

func (*toText) listWeekDay(weekdays []Weekday, callback func(Weekday) string, finalDelim, delim string) string {
	delimJoin := func(arr []string, delimiter, finalDelimiter string) string {
		sb := strings.Builder{}
		for i := 0; i < len(arr); i++ {
			if i != 0 {
				if i == len(arr)-1 {
					sb.WriteByte(' ')
					sb.WriteString(finalDelimiter)
					sb.WriteByte(' ')
				} else {
					sb.WriteString(delimiter)
					sb.WriteByte(' ')
				}
			}
			sb.WriteString(arr[i])
		}

		return sb.String()
	}

	cbRes := make([]string, 0, len(weekdays))
	for _, weekday := range weekdays {
		cbRes = append(cbRes, callback(weekday))
	}

	if finalDelim != "" {
		return delimJoin(cbRes, delim, finalDelim)
	}

	return strings.Join(cbRes, delim+" ")

}
func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
