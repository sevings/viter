package neural

import "github.com/tmc/langchaingo/llms"

type History []llms.MessageContent

func NewHistory() *History {
	hst := History(make([]llms.MessageContent, 1, 2))
	hst[0] = llms.MessageContent{
		Role: llms.ChatMessageTypeSystem,
	}
	return &hst
}

func (h *History) AddMessage() {
	role := llms.ChatMessageTypeHuman
	if len(*h)%2 == 0 {
		role = llms.ChatMessageTypeAI
	}
	*h = append(*h, llms.MessageContent{
		Role: role,
	})
}

func (h *History) AddText(text string) {
	(*h)[len(*h)-1].Parts = append((*h)[len(*h)-1].Parts, llms.TextPart(text))
}

func (h *History) Messages() []llms.MessageContent {
	return *h
}
