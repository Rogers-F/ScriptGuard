package services

import "time"

// SG-023: 全链路使用北京时间（UTC+8）
// BeijingLocation 北京时区
var BeijingLocation = time.FixedZone("Asia/Shanghai", 8*3600)

// NowBeijing 获取当前北京时间
func NowBeijing() time.Time {
	return time.Now().In(BeijingLocation)
}
