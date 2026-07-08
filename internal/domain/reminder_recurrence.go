package domain

import "time"

const maxOccurrencesLimit = 366

func startOfUTCDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func lastDayOfMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func clampDayToMonth(year int, month time.Month, day int) int {
	max := lastDayOfMonth(year, month)
	if day > max {
		return max
	}
	return day
}

// addYearsSafe adiciona anos preservando mês/dia, fazendo clamp de 29/fev.
func addYearsSafe(base time.Time, years int) time.Time {
	year := base.Year() + years
	day := clampDayToMonth(year, base.Month(), base.Day())
	return time.Date(year, base.Month(), day, 0, 0, 0, 0, time.UTC)
}

// addMonthsSafe adiciona meses preservando o dia, fazendo clamp para o último dia do mês.
func addMonthsSafe(base time.Time, months int) time.Time {
	total := int(base.Month()) - 1 + months
	year := base.Year() + total/12
	month := time.Month(total%12 + 1)
	if month < 1 {
		month += 12
		year--
	}
	day := clampDayToMonth(year, month, base.Day())
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// NextOccurrence retorna a próxima ocorrência estritamente após `after`, derivada
// de `base` (primeira ocorrência) segundo a cadência `rec`. Para ReminderRecurrenceNone
// ou cadência inválida, retorna (zero,false).
//
// A aritmética é direta (O(1)): calcula o candidato mais próximo e, se não for
// estritamente posterior a `after`, avança um passo. Datas são tratadas como
// data-calendária em UTC (preserva mês/dia, com clamp de bordas de calendário).
func NextOccurrence(rec ReminderRecurrence, base, after time.Time) (time.Time, bool) {
	if !rec.IsValid() || rec == ReminderRecurrenceNone {
		return time.Time{}, false
	}

	baseUTC := startOfUTCDate(base)
	afterUTC := startOfUTCDate(after)
	if afterUTC.Before(baseUTC) {
		return baseUTC, true
	}

	switch rec {
	case ReminderRecurrenceYearly:
		diffYears := afterUTC.Year() - baseUTC.Year()
		cand := addYearsSafe(baseUTC, diffYears)
		if !cand.After(afterUTC) {
			cand = addYearsSafe(baseUTC, diffYears+1)
		}
		return cand, true
	case ReminderRecurrenceMonthly:
		diffMonths := (afterUTC.Year()-baseUTC.Year())*12 + int(afterUTC.Month()) - int(baseUTC.Month())
		cand := addMonthsSafe(baseUTC, diffMonths)
		if !cand.After(afterUTC) {
			cand = addMonthsSafe(baseUTC, diffMonths+1)
		}
		return cand, true
	}
	return time.Time{}, false
}

// OccurrencesBetween expande as ocorrências em [from,to] derivadas de `base` segundo
// `rec`. Para ReminderRecurrenceNone, retorna `base` se estiver em [from,to].
// O resultado é limitado a maxOccurrencesLimit para evitar loops infinitos.
func OccurrencesBetween(rec ReminderRecurrence, base, from, to time.Time) []time.Time {
	baseUTC := startOfUTCDate(base)
	fromUTC := startOfUTCDate(from)
	toUTC := startOfUTCDate(to)

	if toUTC.Before(fromUTC) {
		return nil
	}

	if !rec.IsValid() || rec == ReminderRecurrenceNone {
		if !baseUTC.Before(fromUTC) && !baseUTC.After(toUTC) {
			return []time.Time{baseUTC}
		}
		return nil
	}

	var occurrences []time.Time
	cursor, ok := NextOccurrence(rec, base, fromUTC.Add(-24 * time.Hour))
	if !ok {
		return nil
	}

	for !cursor.After(toUTC) && len(occurrences) < maxOccurrencesLimit {
		if !cursor.Before(fromUTC) {
			occurrences = append(occurrences, cursor)
		}
		next, ok := NextOccurrence(rec, base, cursor)
		if !ok {
			break
		}
		cursor = next
	}
	return occurrences
}
