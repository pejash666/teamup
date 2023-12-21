package model

type Questionnaire []*Question

type Question struct {
	QuestionID int               `json:"q_id"`
	Question   string            `json:"q_text"`
	Options    map[string]string `json:"options"`
}
