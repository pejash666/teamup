package iface

import "teamup/model"

type HandlerFunc func(c *model.TeamUpContext) (interface{}, error)
