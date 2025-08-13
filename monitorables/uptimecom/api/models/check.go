package models

type (
	Check struct {
		ID   int
		Name string

		IsUP               bool
		IsPaused           bool
		IsUnderMaintenance bool

		Tags []string
	}
)

func (c Check) MatchOneTag(tags []string) bool {
	for _, t := range c.Tags {
		for _, tag := range tags {
			if t == tag {
				return true
			}
		}
	}
	return false
}
