package check

type Contains struct {
	is string
}

func NewContains(s string) *Contains {
	return &Contains{is: s}
}

func (c *Contains) Do(m *slackapi.Message) bool {
	return true
}
