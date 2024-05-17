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

type GetCalibrationQuestionsBody struct {
	SportType string `json:"sport_type"`
	NeedFull  bool   `json:"need_full"` // 是否需要第6题
}

// GetCalibrationQuestions godoc
//
//	@Summary		获取定级问题
//	@Description	获取定级问题详情
//	@Tags			/team_up/user
//	@Accept			json
//	@Produce		json
//	@Param			sport_type	body		string	true	"获取定级问题入参"
//	@Param			need_full	body		bool	true	"是否需要第6题"
//	@Success		200			{object}	GetCalibrationQuestionsResp
//	@Router			/team_up/user/get_calibration_questions [post]
func GetCalibrationQuestions(c *model.TeamUpContext) (interface{}, error) {
	body := &GetCalibrationQuestionsBody{}
	err := c.BindJSON(body)
	if err != nil {
		util.Logger.Printf("[GetCalibrationQuestions] BindJSON failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid req")
	}
	if body.SportType != constant.SportTypePadel && body.SportType != constant.SportTypeTennis && body.SportType != constant.SportTypePickelBall {
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
		// 第6题需要前5题的结果进行额外计算，所以需要判断need_full
		if question.QuestionID == 6 && !body.NeedFull {
			continue
		}
		question.Question = strings.Replace(question.Question, "{sport_type}", body.SportType, -1)
		for k, option := range question.Options {
			question.Options[k] = strings.Replace(option, "{sport_type}", body.SportType, -1)
		}
	}
	util.Logger.Printf("[GetCalibrationQuestions] success, res:%+v", questionnaire)
	return questionnaire, nil
}
