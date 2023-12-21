package util

import (
	"github.com/bytedance/sonic"
	"os"
	"sync"
	"teamup/model"
)

var (
	onceLoader sync.Once

	QuestionnaireEn = model.Questionnaire{}
	QuestionnaireCn = model.Questionnaire{}
)

func LoadFile() {
	onceLoader.Do(func() {
		dataEn, err := os.ReadFile("./questionnaire/questionnaire_en_US.json")
		if err != nil {
			panic(err)
		}
		err = sonic.Unmarshal(dataEn, &QuestionnaireEn)
		if err != nil {
			panic(err)
		}
		dataCn, err := os.ReadFile("./questionnaire/questionnaire_zh_CN.json")
		if err != nil {
			panic(err)
		}
		err = sonic.Unmarshal(dataCn, &QuestionnaireCn)
		if err != nil {
			panic(err)
		}
	})
}
