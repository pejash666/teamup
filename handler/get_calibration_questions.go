package handler

import (
	"strings"
	"teamup/constant"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

type GetCalibrationQuestionsResp struct {
	ErrNo   int32                `json:"err_no"`
	ErrTips string               `json:"err_tips"`
	Data    *model.Questionnaire `json:"data"`
}

// GetCalibrationQuestions godoc
// @Summary      获取定级问题
// @Description  获取定级问题详情
// @Tags         /teamup/user
// @Accept       json
// @Produce      json
// @Param        sport_type  body    string  true  "运动类型"
// @Success      200  {object}  GetCalibrationQuestionsResp
// @Router       /teamup/user/get_calibration_questions [post]
func GetCalibrationQuestions(c *model.TeamUpContext) (interface{}, error) {
	type Body struct {
		SportType string `json:"sport_type"`
	}
	body := &Body{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[GetCalibrationQuestions] BindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid req")
	}
	if body.SportType != constant.SportTypePedal && body.SportType != constant.SportTypeTennis && body.SportType != constant.SportTypePickelBall {
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid sport_type")
	}
	questionnaire := model.Questionnaire{}
	// 根据语言返回不同语言的问卷问题
	if c.Language == "zh_US" {
		questionnaire = util.QuestionnaireEn
	} else {
		// 默认展示中文
		questionnaire = util.QuestionnaireCn
	}
	// 根据入参的sport_type做问题和选项的替换
	for _, question := range questionnaire {
		question.Question = strings.Replace(question.Question, "{sport_type}", body.SportType, -1)
		for k, option := range question.Options {
			question.Options[k] = strings.Replace(option, "{sport_type}", body.SportType, -1)
		}
	}
	util.Logger.Printf("[GetCalibrationQuestions] success, res:%+v", questionnaire)
	return questionnaire, nil
}
