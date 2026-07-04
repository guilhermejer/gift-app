package domain

import (
	"testing"
	"time"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("invalid date %q: %v", s, err)
	}
	return parsed.UTC()
}

func TestNextOccurrence(t *testing.T) {
	now := mustDate(t, "2026-07-04")

	tests := []struct {
		name     string
		rec      ReminderRecurrence
		base     time.Time
		after    time.Time
		want     string
		wantOk   bool
	}{
		{
			name:   "none returns false",
			rec:    ReminderRecurrenceNone,
			base:   mustDate(t, "2026-08-15"),
			after:  now,
			wantOk: false,
		},
		{
			name:   "invalid recurrence returns false",
			rec:    ReminderRecurrence("bimonthly"),
			base:   mustDate(t, "2026-08-15"),
			after:  now,
			wantOk: false,
		},
		{
			name:   "yearly returns next anniversary after now",
			rec:    ReminderRecurrenceYearly,
			base:   mustDate(t, "2020-08-15"),
			after:  now,
			want:   "2026-08-15",
			wantOk: true,
		},
		{
			name:   "yearly with base in future returns base itself",
			rec:    ReminderRecurrenceYearly,
			base:   mustDate(t, "2027-08-15"),
			after:  now,
			want:   "2027-08-15",
			wantOk: true,
		},
		{
			name:   "yearly 29-feb clamps to 28-feb in non-leap year",
			rec:    ReminderRecurrenceYearly,
			base:   mustDate(t, "2020-02-29"),
			after:  mustDate(t, "2026-07-04"),
			want:   "2027-02-28",
			wantOk: true,
		},
		{
			name:   "yearly 29-feb next non-leap clamps to 28-feb",
			rec:    ReminderRecurrenceYearly,
			base:   mustDate(t, "2020-02-29"),
			after:  mustDate(t, "2027-03-01"),
			want:   "2028-02-29",
			wantOk: true,
		},
		{
			name:   "monthly returns next month same day",
			rec:    ReminderRecurrenceMonthly,
			base:   mustDate(t, "2026-06-15"),
			after:  now,
			want:   "2026-07-15",
			wantOk: true,
		},
		{
			name:   "monthly day 31 clamps to last day of short month",
			rec:    ReminderRecurrenceMonthly,
			base:   mustDate(t, "2026-01-31"),
			after:  mustDate(t, "2026-07-04"),
			want:   "2026-07-31",
			wantOk: true,
		},
		{
			name:   "monthly day 31 february clamps to 28",
			rec:    ReminderRecurrenceMonthly,
			base:   mustDate(t, "2026-01-31"),
			after:  mustDate(t, "2026-01-31"),
			want:   "2026-02-28",
			wantOk: true,
		},
		{
			name:   "weekly returns next week",
			rec:    ReminderRecurrenceWeekly,
			base:   mustDate(t, "2026-06-29"),
			after:  now,
			want:   "2026-07-06",
			wantOk: true,
		},
		{
			name:   "daily returns tomorrow",
			rec:    ReminderRecurrenceDaily,
			base:   mustDate(t, "2026-07-01"),
			after:  now,
			want:   "2026-07-05",
			wantOk: true,
		},
		{
			name:   "after equals base returns first step forward",
			rec:    ReminderRecurrenceDaily,
			base:   mustDate(t, "2026-07-04"),
			after:  mustDate(t, "2026-07-04"),
			want:   "2026-07-05",
			wantOk: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := NextOccurrence(tc.rec, tc.base, tc.after)
			if ok != tc.wantOk {
				t.Fatalf("NextOccurrence ok = %v, want %v (got %v)", ok, tc.wantOk, got)
			}
			if !ok {
				return
			}
			gotDate := got.Format("2006-01-02")
			if gotDate != tc.want {
				t.Errorf("NextOccurrence = %s, want %s", gotDate, tc.want)
			}
		})
	}
}

func TestOccurrencesBetween(t *testing.T) {
	tests := []struct {
		name   string
		rec    ReminderRecurrence
		base   time.Time
		from   time.Time
		to     time.Time
		want   []string
	}{
		{
			name: "none in range returns base only",
			rec:  ReminderRecurrenceNone,
			base: mustDate(t, "2026-08-15"),
			from: mustDate(t, "2026-07-01"),
			to:   mustDate(t, "2026-09-30"),
			want: []string{"2026-08-15"},
		},
		{
			name: "none out of range returns empty",
			rec:  ReminderRecurrenceNone,
			base: mustDate(t, "2026-08-15"),
			from: mustDate(t, "2026-09-01"),
			to:   mustDate(t, "2026-09-30"),
			want: []string{},
		},
		{
			name: "daily expands occurrences in window",
			rec:  ReminderRecurrenceDaily,
			base: mustDate(t, "2026-07-01"),
			from: mustDate(t, "2026-07-04"),
			to:   mustDate(t, "2026-07-06"),
			want: []string{"2026-07-04", "2026-07-05", "2026-07-06"},
		},
		{
			name: "yearly skips years outside window",
			rec:  ReminderRecurrenceYearly,
			base: mustDate(t, "2020-08-15"),
			from: mustDate(t, "2026-01-01"),
			to:   mustDate(t, "2027-12-31"),
			want: []string{"2026-08-15", "2027-08-15"},
		},
		{
			name: "monthly in a 3-month window",
			rec:  ReminderRecurrenceMonthly,
			base: mustDate(t, "2026-05-15"),
			from: mustDate(t, "2026-07-01"),
			to:   mustDate(t, "2026-09-30"),
			want: []string{"2026-07-15", "2026-08-15", "2026-09-15"},
		},
		{
			name: "to before from returns empty",
			rec:  ReminderRecurrenceDaily,
			base: mustDate(t, "2026-07-01"),
			from: mustDate(t, "2026-07-10"),
			to:   mustDate(t, "2026-07-05"),
			want: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := OccurrencesBetween(tc.rec, tc.base, tc.from, tc.to)
			gotDates := make([]string, 0, len(got))
			for _, occ := range got {
				gotDates = append(gotDates, occ.Format("2006-01-02"))
			}
			if len(gotDates) != len(tc.want) {
				t.Fatalf("OccurrencesBetween got %d items %v, want %d items %v", len(gotDates), gotDates, len(tc.want), tc.want)
			}
			for i, wantDate := range tc.want {
				if gotDates[i] != wantDate {
					t.Errorf("OccurrencesBetween[%d] = %s, want %s", i, gotDates[i], wantDate)
				}
			}
		})
	}
}

func TestNextOccurrenceLeapYearEdge(t *testing.T) {
	base := mustDate(t, "2020-02-29")

	next, ok := NextOccurrence(ReminderRecurrenceYearly, base, mustDate(t, "2023-01-01"))
	if !ok {
		t.Fatal("expected ok")
	}
	if got, want := next.Format("2006-01-02"), "2023-02-28"; got != want {
		t.Errorf("leap next = %s, want %s", got, want)
	}

	next, ok = NextOccurrence(ReminderRecurrenceYearly, base, mustDate(t, "2024-03-01"))
	if !ok {
		t.Fatal("expected ok")
	}
	if got, want := next.Format("2006-01-02"), "2025-02-28"; got != want {
		t.Errorf("post-leap next = %s, want %s", got, want)
	}
}
