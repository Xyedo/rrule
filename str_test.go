// 2017-2022, Teambition. All rights reserved.

package rrule

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func TestRFCRuleToStr(t *testing.T) {
	nyLoc, _ := time.LoadLocation("America/New_York")
	dtStart := time.Date(2018, 1, 1, 9, 0, 0, 0, nyLoc)

	r, _ := NewRRule(ROption{Freq: MONTHLY, Dtstart: dtStart})
	want := "DTSTART;TZID=America/New_York:20180101T090000\nRRULE:FREQ=MONTHLY"
	if r.String() != want {
		t.Errorf("Expected RFC string %s, got %v", want, r.String())
	}

}

func TestRFCSetToString(t *testing.T) {
	nyLoc, _ := time.LoadLocation("America/New_York")
	dtStart := time.Date(2018, 1, 1, 9, 0, 0, 0, nyLoc)

	r, _ := NewRRule(ROption{Freq: MONTHLY, Dtstart: dtStart})
	want := "DTSTART;TZID=America/New_York:20180101T090000\nRRULE:FREQ=MONTHLY"
	if r.String() != want {
		t.Errorf("Expected RFC string %s, got %v", want, r.String())
	}

	expectedSetStr := "DTSTART;TZID=America/New_York:20180101T090000\nRRULE:FREQ=MONTHLY"

	set := Set{}
	set.RRule(r)
	set.DTStart(dtStart)
	if set.String() != expectedSetStr {
		t.Errorf("Expected RFC Set string %s, got %s", expectedSetStr, set.String())
	}
}

func TestCompatibility(t *testing.T) {
	str := "FREQ=WEEKLY;DTSTART=20120201T093000Z;INTERVAL=5;WKST=TU;COUNT=2;UNTIL=20130130T230000Z;BYSETPOS=2;BYMONTH=3;BYYEARDAY=95;BYWEEKNO=1;BYDAY=MO,+2FR;BYHOUR=9;BYMINUTE=30;BYSECOND=0;BYEASTER=-1"
	r, _ := StrToRRule(str)
	want := "DTSTART:20120201T093000Z\nRRULE:FREQ=WEEKLY;INTERVAL=5;WKST=TU;COUNT=2;UNTIL=20130130T230000Z;BYSETPOS=2;BYMONTH=3;BYYEARDAY=95;BYWEEKNO=1;BYDAY=MO,+2FR;BYHOUR=9;BYMINUTE=30;BYSECOND=0;BYEASTER=-1"
	if s := r.String(); s != want {
		t.Errorf("StrToRRule(%q).String() = %q, want %q", str, s, want)
	}
	r, _ = StrToRRule(want)
	if s := r.String(); s != want {
		t.Errorf("StrToRRule(%q).String() = %q, want %q", want, want, want)
	}

}

func TestInvalidString(t *testing.T) {
	cases := []string{
		"",
		"    ",
		"FREQ",
		"FREQ=HELLO",
		"BYMONTH=",
		"FREQ=WEEKLY;HELLO=WORLD",
		"FREQ=WEEKLY;BYMONTHDAY=I",
		"FREQ=WEEKLY;BYDAY=M",
		"FREQ=WEEKLY;BYDAY=MQ",
		"FREQ=WEEKLY;BYDAY=+MO",
		"BYDAY=MO",
	}
	for _, item := range cases {
		if _, e := StrToRRule(item); e == nil {
			t.Errorf("StrToRRule(%q) = nil, want error", item)
		}
	}
}

func TestSetStr(t *testing.T) {
	setStr := "RRULE:FREQ=DAILY;UNTIL=20180517T235959Z\n" +
		"EXDATE;VALUE=DATE-TIME:20180525T070000Z,20180530T130000Z\n" +
		"RDATE;VALUE=DATE-TIME:20180801T131313Z,20180902T141414Z\n"

	set, err := StrToRRuleSet(setStr)
	if err != nil {
		t.Fatalf("StrToRRuleSet(%s) returned error: %v", setStr, err)
	}

	rule := set.GetRRule()
	if rule == nil {
		t.Errorf("Unexpected rrule parsed")
	}
	if rule.String() != "FREQ=DAILY;UNTIL=20180517T235959Z" {
		t.Errorf("Unexpected rrule: %s", rule.String())
	}

	// matching parsed EXDates
	exDates := set.GetExDate()
	if len(exDates) != 2 {
		t.Errorf("Unexpected number of exDates: %v != 2, %v", len(exDates), exDates)
	}
	if [2]string{timeToStr(exDates[0]), timeToStr(exDates[1])} != [2]string{"20180525T070000Z", "20180530T130000Z"} {
		t.Errorf("Unexpected exDates: %v", exDates)
	}

	// matching parsed RDates
	rDates := set.GetRDate()
	if len(rDates) != 2 {
		t.Errorf("Unexpected number of rDates: %v != 2, %v", len(rDates), rDates)
	}
	if [2]string{timeToStr(rDates[0]), timeToStr(rDates[1])} != [2]string{"20180801T131313Z", "20180902T141414Z"} {
		t.Errorf("Unexpected exDates: %v", exDates)
	}
}

func TestStrToDtStart(t *testing.T) {
	validCases := []string{
		"19970714T133000",
		"19970714T173000Z",
		"TZID=America/New_York:19970714T133000",
	}

	invalidCases := []string{
		"DTSTART;TZID=America/New_York:19970714T133000",
		"19970714T1330000",
		"DTSTART;TZID=:20180101T090000",
		"TZID=:20180101T090000",
		"TZID=notatimezone:20180101T090000",
		"DTSTART:19970714T133000",
		"DTSTART:19970714T133000Z",
		"DTSTART;:19970714T133000Z",
		"DTSTART;:1997:07:14T13:30:00Z",
		";:19970714T133000Z",
		"    ",
		"",
	}

	for _, item := range validCases {
		if _, e := StrToDtStart(item, time.UTC); e != nil {
			t.Errorf("StrToDtStart(%q) error = %s, want nil", item, e.Error())
		}
	}

	for _, item := range invalidCases {
		if _, e := StrToDtStart(item, time.UTC); e == nil {
			t.Errorf("StrToDtStart(%q) err = nil, want not nil", item)
		}
	}
}

func TestStrToDates(t *testing.T) {
	validCases := []string{
		"19970714T133000",
		"19970714T173000Z",
		"VALUE=DATE-TIME:19970714T133000,19980714T133000,19980714T133000",
		"VALUE=DATE-TIME;TZID=America/New_York:19970714T133000,19980714T133000,19980714T133000",
		"VALUE=DATE:19970714T133000,19980714T133000,19980714T133000",
	}

	invalidCases := []string{
		"VALUE:DATE:TIME:19970714T133000,19980714T133000,19980714T133000",
		";:19970714T133000Z",
		"    ",
		"",
		"VALUE=DATE-TIME;TZID=:19970714T133000",
		"VALUE=PERIOD:19970714T133000Z/19980714T133000Z",
	}

	for _, item := range validCases {
		if _, e := StrToDates(item); e != nil {
			t.Errorf("StrToDates(%q) error = %s, want nil", item, e.Error())
		}
		if _, e := StrToDatesInLoc(item, time.Local); e != nil {
			t.Errorf("StrToDates(%q) error = %s, want nil", item, e.Error())
		}
	}

	for _, item := range invalidCases {
		if _, e := StrToDates(item); e == nil {
			t.Errorf("StrToDates(%q) err = nil, want not nil", item)
		}
		if _, e := StrToDatesInLoc(item, time.Local); e == nil {
			t.Errorf("StrToDates(%q) err = nil, want not nil", item)
		}
	}
}

func TestStrToDatesTimeIsCorrect(t *testing.T) {
	nyLoc, _ := time.LoadLocation("America/New_York")
	inputs := []string{
		"VALUE=DATE-TIME:19970714T133000",
		"VALUE=DATE-TIME;TZID=America/New_York:19970714T133000",
	}
	exp := []time.Time{
		time.Date(1997, 7, 14, 13, 30, 0, 0, time.UTC),
		time.Date(1997, 7, 14, 13, 30, 0, 0, nyLoc),
	}

	for i, s := range inputs {
		ts, err := StrToDates(s)
		if err != nil {
			t.Fatalf("StrToDates(%s): error = %s", s, err.Error())
		}
		if len(ts) != 1 {
			t.Fatalf("StrToDates(%s): bad answer: %v", s, ts)
		}
		if !ts[0].Equal(exp[i]) {
			t.Fatalf("StrToDates(%s): bad answer: %v, expected: %v", s, ts[0], exp[i])
		}
	}
}

func TestProcessRRuleName(t *testing.T) {
	validCases := []string{
		"DTSTART;TZID=America/New_York:19970714T133000",
		"RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,TU",
		"EXDATE;VALUE=DATE-TIME:20180525T070000Z,20180530T130000Z",
		"RDATE;TZID=America/New_York;VALUE=DATE-TIME:20180801T131313Z,20180902T141414Z",
	}

	invalidCases := []string{
		"TZID=America/New_York:19970714T133000",
		"19970714T1330000",
		";:19970714T133000Z",
		"FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,TU",
		"    ",
	}

	for _, item := range validCases {
		if _, e := processRRuleName(item); e != nil {
			t.Errorf("processRRuleName(%q) error = %s, want nil", item, e.Error())
		}
	}

	for _, item := range invalidCases {
		if _, e := processRRuleName(item); e == nil {
			t.Errorf("processRRuleName(%q) err = nil, want not nil", item)
		}
	}
}

func TestSetStrCompatibility(t *testing.T) {
	badInputStrs := []string{
		"",
		"FREQ=DAILY;UNTIL=20180517T235959Z",
		"DTSTART:;",
		"RRULE:;",
	}

	for _, badInputStr := range badInputStrs {
		_, err := StrToRRuleSet(badInputStr)
		if err == nil {
			t.Fatalf("StrToRRuleSet(%s) didn't return error", badInputStr)
		}
	}

	inputStr := "DTSTART;TZID=America/New_York:20180101T090000\n" +
		"RRULE:FREQ=DAILY;UNTIL=20180517T235959Z\n" +
		"RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,TU\n" +
		"EXRULE:FREQ=MONTHLY;UNTIL=20180520;BYMONTHDAY=1,2,3\n" +
		"EXDATE;VALUE=DATE-TIME:20180525T070000Z,20180530T130000Z\n" +
		"RDATE;VALUE=DATE-TIME:20180801T131313Z,20180902T141414Z\n"

	set, err := StrToRRuleSet(inputStr)
	if err != nil {
		t.Fatalf("StrToRRuleSet(%s) returned error: %v", inputStr, err)
	}

	nyLoc, _ := time.LoadLocation("America/New_York")
	dtWantTime := time.Date(2018, 1, 1, 9, 0, 0, 0, nyLoc)

	rrule := set.GetRRule()
	if rrule.String() != "DTSTART;TZID=America/New_York:20180101T090000\nRRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=MO,TU" {
		t.Errorf("Unexpected rrule: %s", rrule.String())
	}
	if !dtWantTime.Equal(rrule.dtstart) {
		t.Fatalf("Expected RRule dtstart to be %v got %v", dtWantTime, rrule.dtstart)
	}
	if !dtWantTime.Equal(set.GetDTStart()) {
		t.Fatalf("Expected Set dtstart to be %v got %v", dtWantTime, set.GetDTStart())
	}

	// matching parsed EXDates
	exDates := set.GetExDate()
	if len(exDates) != 2 {
		t.Errorf("Unexpected number of exDates: %v != 2, %v", len(exDates), exDates)
	}
	if [2]string{timeToStr(exDates[0]), timeToStr(exDates[1])} != [2]string{"20180525T070000Z", "20180530T130000Z"} {
		t.Errorf("Unexpected exDates: %v", exDates)
	}

	// matching parsed RDates
	rDates := set.GetRDate()
	if len(rDates) != 2 {
		t.Errorf("Unexpected number of rDates: %v != 2, %v", len(rDates), rDates)
	}
	if [2]string{timeToStr(rDates[0]), timeToStr(rDates[1])} != [2]string{"20180801T131313Z", "20180902T141414Z"} {
		t.Errorf("Unexpected exDates: %v", exDates)
	}

	dtWantAfter := time.Date(2018, 1, 2, 9, 0, 0, 0, nyLoc)
	dtAfter := set.After(dtWantTime, false)
	if !dtWantAfter.Equal(dtAfter) {
		t.Errorf("Next time wrong should be %s but is %s", dtWantAfter, dtAfter)
	}

	// String to set to string comparison
	setStr := set.String()
	setFromSetStr, _ := StrToRRuleSet(setStr)

	if setStr != setFromSetStr.String() {
		t.Errorf("Expected string output\n %s \nbut got\n %s\n", setStr, setFromSetStr.String())
	}
}

func TestSetParseLocalTimes(t *testing.T) {
	moscow, _ := time.LoadLocation("Europe/Moscow")

	t.Run("DtstartTimeZoneIsUsed", func(t *testing.T) {
		input := []string{
			"DTSTART;TZID=Europe/Moscow:20180220T090000",
			"RDATE;VALUE=DATE-TIME:20180223T100000",
		}
		s, err := StrSliceToRRuleSet(input)
		if err != nil {
			t.Error(err)
		}
		d := s.GetRDate()[0]
		if !d.Equal(time.Date(2018, 02, 23, 10, 0, 0, 0, moscow)) {
			t.Error("Bad time parsed: ", d)
		}
	})

	t.Run("DtstartTimeZoneValidOutput", func(t *testing.T) {
		input := []string{
			"DTSTART;TZID=Europe/Moscow:20180220T090000",
			"RDATE;VALUE=DATE-TIME:20180223T100000",
		}
		expected := "DTSTART;TZID=Europe/Moscow:20180220T090000\nRDATE;TZID=Europe/Moscow:20180223T100000"
		s, err := StrSliceToRRuleSet(input)
		if err != nil {
			t.Error(err)
		}

		sRRule := s.String()

		if sRRule != expected {
			t.Errorf("DTSTART output not valid. Expected: \n%s \n Got: \n%s", expected, sRRule)
		}
	})

	t.Run("DtstartUTCValidOutput", func(t *testing.T) {
		input := []string{
			"DTSTART:20180220T090000Z",
			"RDATE;VALUE=DATE-TIME:20180223T100000",
		}
		expected := "DTSTART:20180220T090000Z\nRDATE:20180223T100000Z"
		s, err := StrSliceToRRuleSet(input)
		if err != nil {
			t.Error(err)
		}

		sRRule := s.String()

		if sRRule != expected {
			t.Errorf("DTSTART output not valid. Expected: \n%s \n Got: \n%s", expected, sRRule)
		}
	})

	t.Run("SpecifiedDefaultZoneIsUsed", func(t *testing.T) {
		input := []string{
			"RDATE;VALUE=DATE-TIME:20180223T100000",
		}
		s, err := StrSliceToRRuleSetInLoc(input, moscow)
		if err != nil {
			t.Error(err)
		}
		d := s.GetRDate()[0]
		if !d.Equal(time.Date(2018, 02, 23, 10, 0, 0, 0, moscow)) {
			t.Error("Bad time parsed: ", d)
		}
	})
}

func TestRDateValueDateStr(t *testing.T) {
	t.Run("DefaultToUTC", func(t *testing.T) {
		input := []string{
			"RDATE;VALUE=DATE:20180223",
		}
		s, err := StrSliceToRRuleSet(input)
		if err != nil {
			t.Error(err)
		}
		d := s.GetRDate()[0]
		if !d.Equal(time.Date(2018, 02, 23, 0, 0, 0, 0, time.UTC)) {
			t.Error("Bad time parsed: ", d)
		}
	})

	t.Run("PreserveExplicitTimezone", func(t *testing.T) {
		denver, _ := time.LoadLocation("America/Denver")
		input := []string{
			"RDATE;VALUE=DATE;TZID=America/Denver:20180223",
		}
		s, err := StrSliceToRRuleSet(input)
		if err != nil {
			t.Error(err)
		}
		d := s.GetRDate()[0]
		if !d.Equal(time.Date(2018, 02, 23, 0, 0, 0, 0, denver)) {
			t.Error("Bad time parsed: ", d)
		}
	})
}

func TestStrSetEmptySliceParse(t *testing.T) {
	s, err := StrSliceToRRuleSet([]string{})
	if err != nil {
		t.Error(err)
	}
	if s == nil {
		t.Error("Empty set should not be nil")
	}
}

func TestStrSetParseErrors(t *testing.T) {
	inputs := [][]string{
		{"RRULE:XXX"},
		{"RDATE;TZD=X:1"},
	}

	for _, ss := range inputs {
		if _, err := StrSliceToRRuleSet(ss); err == nil {
			t.Error("Expected parse error for rules: ", ss)
		}
	}
}

func TestStrToOption(t *testing.T) {
	t.Run("ValidParsing", func(t *testing.T) {
		str := `
	{
		"rule":"DTSTART:20241102T090000Z\nRRULE:FREQ=MONTHLY;BYDAY=21MO;BYMONTH=08;UNTIL=20270306T175200Z;INTERVAL=9"
	}
	`

		var input struct {
			Rule string `json:"rule"`
		}
		err := json.Unmarshal([]byte(str), &input)
		if err != nil {
			t.Fatalf("json.Unmarshal error: %v", err)
		}

		log.Println(input)

		rrule, err := StrToRRule(input.Rule)
		if err != nil {
			t.Fatalf("StrToRRule(%q) returned error: %v", str, err)
		}

		z := rrule.After(time.Now(), false)
		if !z.IsZero() {
			t.Fatalf("Expected zero time got %v", z)
		}

		z = rrule.Before(time.Now(), false)
		if !z.IsZero() {
			t.Fatalf("Expected zero time got %v", z)
		}
	})

	t.Run("ValidWithTZID", func(t *testing.T) {
		input := "DTSTART;TZID=Asia/Bangkok:20240521T114100\nRRULE:FREQ=WEEKLY;UNTIL=20240521T114200;WKST=MO"
		rrule, err := StrToRRule(input)
		if err != nil {
			t.Fatalf("StrToRRule(%q) returned error: %v", input, err)
		}

		expectedDtstart := time.Date(2024, 5, 21, 11, 41, 0, 0, time.FixedZone("Asia/Bangkok", 7*3600))
		if !rrule.GetDTStart().Equal(expectedDtstart) {
			t.Errorf("Expected dtstart %v, got %v", expectedDtstart, rrule.GetDTStart())
		}

		expectedUntil := time.Date(2024, 5, 21, 11, 42, 0, 0, time.FixedZone("Asia/Bangkok", 7*3600))
		if !rrule.GetUntil().Equal(expectedUntil) {
			t.Errorf("Expected until %v, got %v", expectedUntil, rrule.GetUntil())
		}
	})
}
func TestToText(t *testing.T) {
	var tests = []struct {
		rule     string
		expected string
	}{
		{rule: "DTSTART:20171101T010000Z\nRRULE:UNTIL=20171214T013000Z;FREQ=DAILY;INTERVAL=2;WKST=MO;BYHOUR=11,12;BYMINUTE=30;BYSECOND=0", expected: "every 2 days at 11 and 12 until December 14, 2017"},
		{rule: "DTSTART:20171101T010000Z\nRRULE:UNTIL=20171214T013000Z;FREQ=DAILY;INTERVAL=2;WKST=MO;BYHOUR=11;BYMINUTE=30;BYSECOND=0", expected: "every 2 days at 11 until December 14, 2017"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			r, err := StrToRRule(tt.rule)
			if err != nil {
				t.Fatalf("StrToRRule(%q) returned error: %v", tt.rule, err)
			}

			if r.ToText() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, r.ToText())
			}
		})
	}
}

type indonesianFormatter struct{}

// Format implements TimeFormatter.
func (i indonesianFormatter) Format(t time.Time) string {
	month := [...]string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}[t.Month()-1]

	return strconv.Itoa(t.Day()) + " " + month + " " + strconv.Itoa(t.Year())
}

// MonthName implements TimeFormatter.
func (i indonesianFormatter) MonthName(m int) string {
	return [...]string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}[m-1]
}

// Nth implements TimeFormatter.
func (indonesianFormatter) Nth(i int) string {
	if i == -1 {
		return "terakhir"
	}

	npos := abs(i)
	n := strings.Builder{}
	switch npos {
	case 1:
		return "pertama"
	default:
		n.WriteString("ke-" + strconv.Itoa(npos))
	}
	if i < 0 {
		return n.String() + " terakhir"
	}

	return n.String()
}

// WeekDayName implements TimeFormatter.
func (i indonesianFormatter) WeekDayName(w Weekday) string {
	weekday := [...]string{
		"Senin",
		"Selasa",
		"Rabu",
		"Kamis",
		"Jumat",
		"Sabtu",
		"Minggu",
	}[w.Day()]

	if n := w.N(); n != 0 {
		return weekday + " " + i.Nth(n)
	}

	return weekday
}

var _ TimeFormatter = indonesianFormatter{}

func TestToTextWithCustomFormatter(t *testing.T) {
	var tests = []struct {
		rule     string
		expected string
	}{
		{rule: "DTSTART:20171101T010000Z\nRRULE:UNTIL=20171214T013000Z;FREQ=DAILY;INTERVAL=2;WKST=MO;BYHOUR=11,12;BYMINUTE=30;BYSECOND=0", expected: "setiap 2 hari pada pukul 11 dan 12 sampai 14 Desember 2017"},
		{rule: "DTSTART:20171101T010000Z\nRRULE:UNTIL=20171214T013000Z;FREQ=DAILY;INTERVAL=2;WKST=MO;BYHOUR=11;BYMINUTE=30;BYSECOND=0", expected: "setiap 2 hari pada pukul 11 sampai 14 Desember 2017"},
		{expected: "setiap hari", rule: "RRULE:FREQ=DAILY"},
		{expected: "setiap hari pada pukul 10, 12 dan 17", rule: "RRULE:FREQ=DAILY;BYHOUR=10,12,17"},
		{expected: "setiap minggu pada hari Minggu pada pukul 10, 12 dan 17", rule: "RRULE:FREQ=WEEKLY;BYDAY=SU;BYHOUR=10,12,17"},
		{expected: "setiap minggu", rule: "RRULE:FREQ=WEEKLY"},
		{expected: "setiap jam", rule: "RRULE:FREQ=HOURLY"},
		{expected: "setiap 4 jam", rule: "RRULE:INTERVAL=4;FREQ=HOURLY"},
		{expected: "setiap minggu pada hari Selasa", rule: "RRULE:FREQ=WEEKLY;BYDAY=TU"},
		{expected: "setiap minggu pada hari Senin, Rabu", rule: "RRULE:FREQ=WEEKLY;BYDAY=MO,WE"},
		{expected: "setiap hari kerja", rule: "RRULE:FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR"},
		{expected: "setiap 2 minggu", rule: "RRULE:INTERVAL=2;FREQ=WEEKLY"},
		{expected: "setiap bulan", rule: "RRULE:FREQ=MONTHLY"},
		{expected: "setiap 6 bulan", rule: "RRULE:INTERVAL=6;FREQ=MONTHLY"},
		{expected: "setiap tahun", rule: "RRULE:FREQ=YEARLY"},
		{expected: "setiap tahun pada hari Jumat pertama", rule: "RRULE:FREQ=YEARLY;BYDAY=+1FR"},
		{expected: "setiap tahun pada hari Jumat ke-13", rule: "RRULE:FREQ=YEARLY;BYDAY=+13FR"},
		{expected: "setiap bulan pada hari ke-4", rule: "RRULE:FREQ=MONTHLY;BYMONTHDAY=4"},
		{expected: "setiap bulan pada hari ke-4 terakhir", rule: "RRULE:FREQ=MONTHLY;BYMONTHDAY=-4"},
		{expected: "setiap bulan pada hari Selasa ke-3", rule: "RRULE:FREQ=MONTHLY;BYDAY=+3TU"},
		{expected: "setiap bulan pada hari Selasa ke-3 terakhir", rule: "RRULE:FREQ=MONTHLY;BYDAY=-3TU"},
		{expected: "setiap bulan pada hari Senin terakhir", rule: "RRULE:FREQ=MONTHLY;BYDAY=-1MO"},
		{expected: "setiap bulan pada hari Jumat ke-2 terakhir", rule: "RRULE:FREQ=MONTHLY;BYDAY=-2FR"},
		{expected: "setiap minggu sampai 1 Januari 2007", rule: "RRULE:FREQ=WEEKLY;UNTIL=20070101T080000Z"},
		{expected: "setiap minggu sebanyak 20 kali", rule: "RRULE:FREQ=WEEKLY;COUNT=20"},
	}

	bun := i18n.NewBundle(language.English)
	bun.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bun.MustLoadMessageFile("active.en.toml")

	bun.MustLoadMessageFile("example/active.id.toml")
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {

			r, err := StrToRRuleWithi18n(tt.rule, bun)
			if err != nil {
				t.Fatalf("StrToRRule(%q) returned error: %v", tt.rule, err)
			}

			got, err := r.ToTextWithCustomFormatter(indonesianFormatter{}, "id")
			if err != nil {
				t.Fatalf("ToTextWithCustomFormatter error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, got)
			}
		})
	}
}
