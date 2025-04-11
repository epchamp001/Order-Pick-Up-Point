package mapper

import "time"

const timeLayout = time.RFC3339

func ParseTime(timeStr string) (*time.Time, error) {
	if timeStr == "" {
		return nil, nil
	}
	t, err := time.Parse(timeLayout, timeStr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func FormatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(timeLayout)
}
