package helper

import (
	"time"
)

var ShanghaiLoc = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	return loc
}()

// ShanghaiNow 返回上海时区当前时间
func ShanghaiNow() time.Time {
	return time.Now().In(ShanghaiLoc)
}

// ShanghaiDateString 返回上海时区今日日期字符串 "2006-01-02"
func ShanghaiDateString() string {
	return ShanghaiNow().Format("2006-01-02")
}

// ShanghaiDateOnly 将时间转为上海时区日期字符串
func ShanghaiDateOnly(t time.Time) string {
	return t.In(ShanghaiLoc).Format("2006-01-02")
}

// ParseShanghaiDate 解析日期字符串为上海时区时间
func ParseShanghaiDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, ShanghaiLoc)
}

// IsYesterday 判断 last 是否为 today 的前一天
func IsYesterday(today, last time.Time) bool {
	t := today.In(ShanghaiLoc)
	l := last.In(ShanghaiLoc)
	y, m, d := t.Date()
	ly, lm, ld := l.Date()
	// last == today - 1 day
	yesterday := time.Date(y, m, d-1, 0, 0, 0, 0, ShanghaiLoc)
	return ly == yesterday.Year() && lm == yesterday.Month() && ld == yesterday.Day()
}
