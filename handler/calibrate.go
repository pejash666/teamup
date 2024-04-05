package handler

import (
	"github.com/bytedance/sonic"
	"sort"
	"strconv"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/iface"
	"teamup/model"
	"teamup/util"
)

type Question struct {
	QID    int    `json:"q_id"`
	Option string `json:"option"`
}

var MaxScoreMap = map[string]float32{
	"A": 1.5,
	"B": 3.5,
	"C": 5.0,
	"D": 6.0,
	"E": 7.0,
}

var CalculatingMap = map[string]map[string]float32{
	"1": {"A": 0, "B": 1.5, "C": 4.0, "D": 5.5, "E": 7.0},
	"2": {"A": 0, "B": 0.5, "C": 0.5, "D": 1},
	"3": {"A": 0, "B": 0, "C": 0.5, "D": 0.5},
	"4": {"A": 0, "B": 0, "C": 0.5, "D": 0.5},
	"5": {"A": 0.5, "B": 0.5, "C": 0, "D": -0.5, "E": -0.5},
	"6": {"A": 0, "B": 0.5, "C": 0, "D": -0.5},
}

func getCalibrationScore(qid int, option string) float32 {
	return CalculatingMap[strconv.FormatInt(int64(qid), 10)][option]
}

type CalibrateResp struct {
	ErrNo   int32         `json:"err_no"`
	ErrTips string        `json:"err_tips"`
	Data    *CalibrateRes `json:"data"`
}

type CalibrateRes struct {
	SportType  string  `json:"sport_type"`
	Level      float32 `json:"level"`
	ProofImage string  `json:"proof_image"` // 只有定位Pro的人才需要
}

// Calibrate godoc
//
//	@Summary		用户定级
//	@Description	获取定级问题详情
//	@Tags			/team_up/user
//	@Accept			json
//	@Produce		json
//	@Param			sport_type		formData	string	true	"运动类型"
//	@Param			questionnaire	formData	string	true	"问卷结构体"
//	@Success		200				{object}	GetCalibrationQuestionsResp
//	@Router			/team_up/user/calibrate [post]
func Calibrate(c *model.TeamUpContext) (interface{}, error) {
	sportType := c.PostForm("sport_type")
	if sportType != constant.SportTypePedal && sportType != constant.SportTypeTennis && sportType != constant.SportTypePickelBall {
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid sport_type")
	}
	questionnaireParam := c.PostForm("questionnaire")
	questionnaire := make([]*Question, 0)
	err := sonic.UnmarshalString(questionnaireParam, &questionnaire)
	if err != nil {
		util.Logger.Printf("[Calibrate] unmarshal questionnaire failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid questionnaire")
	}
	file := c.PostForm("proof")
	//err := c.BindJSON(body)
	//if err != nil {
	//	util.Logger.Printf("[Calibrate] BindJSON failed, err:%v", err)
	//	return nil, iface.NewBackEndError(iface.ParamsError, "invalid req")
	//}
	//if body.SportType != constant.SportTypePedal && body.SportType != constant.SportTypeTennis && body.SportType != constant.SportTypePickelBall {
	//	return nil, iface.NewBackEndError(iface.ParamsError, "invalid sport_type")
	//}
	if len(questionnaire) < 1 {
		return nil, iface.NewBackEndError(iface.ParamsError, "invalid questionnaire")
	}
	// 按照qid重新sort一下
	sort.SliceStable(questionnaire, func(i, j int) bool {
		return questionnaire[i].QID < questionnaire[j].QID
	})
	isPro := false
	totalScore := float32(0)
	maxScore := float32(0)
	for _, q := range questionnaire {
		// 特殊处理选择7.0的人,需要保存上传的文件
		if q.QID == 1 {
			maxScore = MaxScoreMap[q.Option]
			if q.Option == "E" {
				if file == "" {
					return nil, iface.NewBackEndError(iface.ParamsError, "pro need proof file")
				}
				//	// 不能大于1mb
				//	if file.Size > 1<<20 {
				//		util.Logger.Printf("[Calibrate] file size is too big")
				//		return nil, iface.NewBackEndError(iface.ParamsError, "file too big")
				//	}
				//	fileName := strings.Split(file.Filename, ".")
				//	if fileName[len(fileName)-1] != "png" && fileName[len(fileName)-1] != "jpeg" {
				//		util.Logger.Printf("[Calibrate] invalid file, should either png or jpeg, now:%v", fileName[len(fileName)-1])
				//		return nil, iface.NewBackEndError(iface.ParamsError, "invalid filename")
				//	}
				//	filePath := path.Join("./user_calibration_proof", c.BasicUser.OpenID+"."+fileName[len(fileName)-1])
				//	// todo: 是否要将这个存起来？
				//	err = c.SaveUploadedFile(file, filePath)
				//	if err != nil {
				//		util.Logger.Printf("[Calibrate] iSaveUploadedFile failed, err:%v", err)
				//		return nil, iface.NewBackEndError(iface.ParamsError, "save file failed")
				//	}
				isPro = true
			}
		}
		totalScore += getCalibrationScore(q.QID, q.Option)
		if isPro {
			break
		}
	}

	// 根据第一个问题的答案，有一个分数的上限
	if totalScore > maxScore {
		totalScore = maxScore
	}
	// 如果最后 < 0; 则展示0
	if totalScore < 0 {
		totalScore = 0
	}

	// 更新用户表
	user := &mysql.WechatUserInfo{}
	err = util.DB().Where("open_id = ? AND sport_type = ?", c.BasicUser.OpenID, sportType).Take(user).Error
	if err != nil {
		util.Logger.Printf("[Calibrate] query record failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.MysqlError, err.Error())
	}
	// level保存的时候 是 x 1000的整数
	user.Level = int(totalScore * 1000)
	// 非pro直接更新calibration状态，pro需要等待人工审批
	if !isPro {
		user.IsCalibrated = 1
	} else {
		user.CalibrationProof = file
	}
	err = util.DB().Save(user).Error
	if err != nil {
		util.Logger.Printf("[Calibrate] save user failed, err:%v", err)
		return nil, iface.NewBackEndError(iface.InternalError, err.Error())
	}
	res := &CalibrateRes{
		SportType: sportType,
		Level:     totalScore,
	}
	//baseImg := ""
	// 如果是上传了图片，则返回一个base64的字符串
	//if proofPath != "" && isPro {
	//	file, err := os.Open(proofPath)
	//	if err != nil {
	//		return nil, iface.NewBackEndError(iface.InternalError, "open proof_path failed")
	//	}
	//	defer file.Close()
	//	imgByte, _ := io.ReadAll(file)
	//	mimeType := http.DetectContentType(imgByte)
	//	switch mimeType {
	//	case "image/jpeg":
	//		baseImg = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imgByte)
	//	case "image/png":
	//		baseImg = "data:image/png;base64," + base64.StdEncoding.EncodeToString(imgByte)
	//	}
	//	res["proof_image"] = baseImg
	//}
	if file != "" && isPro {
		res.ProofImage = file
	}
	util.Logger.Printf("[Calibrate] success")
	return res, nil
}
