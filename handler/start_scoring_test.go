package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"math/rand"
	"reflect"
	"teamup/constant"
	"teamup/db/mysql"
	"teamup/model"
	"teamup/util"
	"testing"
	"time"
)

func Test_dividePickleBall(t *testing.T) {
	// 初始化Logger
	util.InitLogger()
	eventMeta := &mysql.EventMeta{
		Model:            gorm.Model{},
		Status:           "",
		Creator:          "",
		IsHost:           0,
		OrganizationID:   0,
		SportType:        "",
		MatchType:        "",
		GameType:         constant.EventGameTypeDuo,
		ScoreRule:        constant.PedalScoreRuleAmericano,
		Scorers:          "haosha",
		LowestLevel:      0,
		HighestLevel:     0,
		IsPublic:         0,
		IsBooked:         0,
		FieldType:        "",
		Date:             "",
		Weekday:          "",
		City:             "",
		Name:             "",
		Desc:             "",
		StartTime:        0,
		StartTimeStr:     "",
		EndTime:          0,
		EndTimeStr:       "",
		FieldName:        "",
		MaxPlayerNum:     0,
		CurrentPlayerNum: 0,
		CurrentPlayer:    "",
		Price:            0,
	}

	players := []*model.Player{
		{
			OpenID: "ADA",
			Level:  1.0,
		},
		{
			OpenID: "BDB",
			Level:  1.2,
		},
		{
			OpenID: "CDC",
			Level:  3.0,
		},
		{
			OpenID: "DDD",
			Level:  2.5,
		},
		{
			OpenID: "EDE",
			Level:  2.0,
		},
		{
			OpenID: "FDF",
			Level:  1.7,
		},
		{
			OpenID: "GDG",
			Level:  1.8,
		},
		{
			OpenID: "HDH",
			Level:  2.3,
		},
	}

	c := &model.TeamUpContext{
		Context: &gin.Context{
			Request:  nil,
			Writer:   nil,
			Params:   nil,
			Keys:     nil,
			Errors:   nil,
			Accepted: nil,
		},
		AppInfo:     nil,
		BasicUser:   nil,
		AccessToken: "",
		ID:          0,
		Rand:        rand.New(rand.NewSource(time.Now().UnixNano())),
		Timestamp:   0,
		Language:    "",
	}

	res := dividePedal(c, eventMeta, players)

	fmt.Println(ToReadable(res))

}

func ToReadable(value interface{}) string {
	if value == nil {
		return "nil"
	}
	var str string
	switch vt := value.(type) {
	case string:
		str = vt
	case []byte:
		str = string(vt)
	default:
		kind := reflect.TypeOf(value).Kind()
		switch kind {
		case reflect.Struct, reflect.Interface, reflect.Map, reflect.Slice:
			bytes, err := json.MarshalIndent(value, "", " ")
			if err != nil {
				return fmt.Sprintf("format error: %v", err)
			}
			str = string(bytes)
		default:
			str = fmt.Sprint(value)
		}
	}
	return str
}
