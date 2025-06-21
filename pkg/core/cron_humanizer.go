package core

type CronHumanizer struct{}

func NewCronHumanizer() *CronHumanizer {
	return &CronHumanizer{}
}

func (h *CronHumanizer) Humanize(expr string) string {
	if expr == "* * * * *" {
		return "Every minute"
	}
	if expr == "0 * * * *" {
		return "Every hour"
	}
	if expr == "0 0 * * *" {
		return "Every day at midnight"
	}
	return expr
}
