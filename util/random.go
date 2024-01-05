package util

import "teamup/model"

// ShuffleSlice 对任意切片进行随机打散
func ShuffleSlice(c *model.TeamUpContext, slice []interface{}) {
	c.Rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}
