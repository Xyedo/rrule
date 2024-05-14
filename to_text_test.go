package rrule

import (
	"strings"
	"testing"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func TestToString(t *testing.T) {
	var tests = []struct {
		expected string
		rule     string
	}{
		{expected: "Every day", rule: "RRULE:FREQ=DAILY"},
		{expected: "Every day at 10, 12 and 17", rule: "RRULE:FREQ=DAILY;BYHOUR=10,12,17"},
		{expected: "Every week on Sunday at 10, 12 and 17", rule: "RRULE:FREQ=WEEKLY;BYDAY=SU;BYHOUR=10,12,17"},
		{expected: "Every week", rule: "RRULE:FREQ=WEEKLY"},
		{expected: "Every hour", rule: "RRULE:FREQ=HOURLY"},
		{expected: "Every 4 hours", rule: "RRULE:INTERVAL=4;FREQ=HOURLY"},
		{expected: "Every week on Tuesday", rule: "RRULE:FREQ=WEEKLY;BYDAY=TU"},
		{expected: "Every week on Monday, Wednesday", rule: "RRULE:FREQ=WEEKLY;BYDAY=MO,WE"},
		{expected: "Every weekday", rule: "RRULE:FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR"},
		{expected: "Every 2 weeks", rule: "RRULE:INTERVAL=2;FREQ=WEEKLY"},
		{expected: "Every month", rule: "RRULE:FREQ=MONTHLY"},
		{expected: "Every 6 months", rule: "RRULE:INTERVAL=6;FREQ=MONTHLY"},
		{expected: "Every year", rule: "RRULE:FREQ=YEARLY"},
		{expected: "Every year on the 1st Friday", rule: "RRULE:FREQ=YEARLY;BYDAY=+1FR"},
		{expected: "Every year on the 13th Friday", rule: "RRULE:FREQ=YEARLY;BYDAY=+13FR"},
		{expected: "Every month on the 4th", rule: "RRULE:FREQ=MONTHLY;BYMONTHDAY=4"},
		{expected: "Every month on the 4th last", rule: "RRULE:FREQ=MONTHLY;BYMONTHDAY=-4"},
		{expected: "Every month on the 3rd Tuesday", rule: "RRULE:FREQ=MONTHLY;BYDAY=+3TU"},
		{expected: "Every month on the 3rd last Tuesday", rule: "RRULE:FREQ=MONTHLY;BYDAY=-3TU"},
		{expected: "Every month on the last Monday", rule: "RRULE:FREQ=MONTHLY;BYDAY=-1MO"},
		{expected: "Every month on the 2nd last Friday", rule: "RRULE:FREQ=MONTHLY;BYDAY=-2FR"},
		{expected: "Every week until January 1, 2007", rule: "RRULE:FREQ=WEEKLY;UNTIL=20070101T080000Z"},
		{expected: "Every week for 20 times", rule: "RRULE:FREQ=WEEKLY;COUNT=20"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			r, err := StrToRRule(tt.rule)
			if err != nil {
				t.Fatalf("failed to parse rrule: %v", err)
			}

			bundle := i18n.NewBundle(language.English)
			loc := i18n.NewLocalizer(bundle, "en")
			got := newToText(r, loc, defaultFormatter{}).ToString()
			if !strings.EqualFold(got, tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
