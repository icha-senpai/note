// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package api

import (
	"fmt"
	"html"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/kernel/util"
)

type ColorScheme struct {
	Primary   string
	Secondary string
}

var colorSchemes = map[string]ColorScheme{
	"red":    {"#d13d51", "#ba2c3f"},
	"blue":   {"#3eb0ea", "#0097e6"},
	"yellow": {"#eec468", "#d89b18"},
	"green":  {"#52E0B8", "#19b37a"},
	"purple": {"#a36cda", "#8952d5"},
	"pink":   {"#f183aa", "#e05b8a"},
	"orange": {"#f3865e", "#ef5e2a"},
	"grey":   {"#576574", "#374a60"},
}

func getColorScheme(color string) ColorScheme {

	color = strings.TrimSpace(color)

	if scheme, ok := colorSchemes[strings.ToLower(color)]; ok {
		return scheme
	}

	if !strings.HasPrefix(color, "#") && len(color) > 0 {
		color = "#" + color
	}

	hexColorPattern := `^#?([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`
	if matched, _ := regexp.MatchString(hexColorPattern, color); matched {

		if !strings.HasPrefix(color, "#") {
			color = "#" + color
		}

		if len(color) == 4 {
			r := string(color[1])
			g := string(color[2])
			b := string(color[3])
			color = "#" + r + r + g + g + b + b
		}

		secondary := darkenColor(color, 0.1)

		return ColorScheme{
			Primary:   color,
			Secondary: secondary,
		}
	}

	return colorSchemes["red"]
}

func darkenColor(hexColor string, factor float64) string {

	hex := hexColor[1:]

	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)

	r = int64(float64(r) * (1 - factor))
	g = int64(float64(g) * (1 - factor))
	b = int64(float64(b) * (1 - factor))

	r = int64(math.Max(0, float64(r)))
	g = int64(math.Max(0, float64(g)))
	b = int64(math.Max(0, float64(b)))

	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func getDynamicIcon(c *gin.Context) {
	// Add internal kernel API `/api/icon/getDynamicIcon`

	iconType := c.Query("type")
	if "" == iconType {
		iconType = "1"
	}
	color := c.Query("color")
	date := c.Query("date")
	lang := c.Query("lang")
	if "" == lang {
		lang = util.Lang
	}
	lang = util.LangToBCP47(lang)
	weekdayType := c.Query("weekdayType")
	if "" == weekdayType {
		weekdayType = "1"
	}

	dateInfo := getDateInfo(date, lang, weekdayType)
	var svg string
	switch iconType {
	case "1":

		svg = generateTypeOneSVG(color, dateInfo)
	case "2":

		svg = generateTypeTwoSVG(color, dateInfo)
	case "3":

		svg = generateTypeThreeSVG(color, dateInfo)
	case "4":

		svg = generateTypeFourSVG(color, dateInfo)
	case "5":

		svg = generateTypeFiveSVG(color, dateInfo)
	case "6":

		svg = generateTypeSixSVG(color, lang, weekdayType, dateInfo)
	case "7":

		svg = generateTypeSevenSVG(color, lang, dateInfo)
	case "8":

		content := c.Query("content")
		id := c.Query("id")
		svg = generateTypeEightSVG(color, content, id)
	default:

		svg = generateTypeOneSVG(color, dateInfo)
	}

	if !model.Conf.Editor.AllowSVGScript {
		var err error
		svg, err = util.SanitizeSVG(svg)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	c.Header("Content-Type", "image/svg+xml")
	c.Header("Content-Security-Policy", "script-src 'none'; object-src 'none'; base-uri 'none'")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "no-cache")
	c.Header("Pragma", "no-cache")
	c.String(http.StatusOK, svg)
}

func getDateInfo(dateStr string, lang string, weekdayType string) map[string]any {

	var date time.Time
	var err error
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			date = time.Now()
		}
	}

	year := date.Year()
	month := date.Format("Jan")
	day := date.Day()
	var weekdayStr string

	switch weekdayType {
	case "1":
		weekdayStr = date.Format("Mon")
	case "2":
		weekdayStr = date.Format("Mon")
		weekdayStr = strings.ToUpper(weekdayStr)
	case "3":
		weekdayStr = date.Format("Monday")
	case "4":
		weekdayStr = date.Format("Monday")
		weekdayStr = strings.ToUpper(weekdayStr)
	default:
		weekdayStr = date.Format("Mon")
	}
	// Calculate week number and ISO year
	isoYear, weekNum := date.ISOWeek()
	weekNumStr := fmt.Sprintf("%dW", weekNum)

	isWeekend := date.Weekday() == time.Saturday || date.Weekday() == time.Sunday
	// Calculate days until today
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	countDown := daysBetween(today, date)

	return map[string]any{
		"year":      year,
		"isoYear":   isoYear,
		"month":     month,
		"day":       day,
		"date":      fmt.Sprintf("%02d-%02d", date.Month(), date.Day()),
		"weekday":   weekdayStr,
		"week":      weekNumStr,
		"countDown": countDown,
		"isWeekend": isWeekend,
	}
}

func daysBetween(date1, date2 time.Time) int {

	date1 = time.Date(date1.Year(), date1.Month(), date1.Day(), 0, 0, 0, 0, time.UTC)
	date2 = time.Date(date2.Year(), date2.Month(), date2.Day(), 0, 0, 0, 0, time.UTC)

	swap := false
	if date1.After(date2) {
		date1, date2 = date2, date1
		swap = true
	}

	days := 0
	for y := date1.Year(); y < date2.Year(); y++ {
		if isLeapYear(y) {
			days += 366
		} else {
			days += 365
		}
	}

	days += date2.YearDay() - date1.YearDay()

	if swap {
		return -days
	}
	return days
}

func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func generateTypeOneSVG(color string, dateInfo map[string]any) string {
	colorScheme := getColorScheme(color)

	return fmt.Sprintf(`
    <svg id="dynamic_icon_type1" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">
    <path d="M512,447.5c0,32-25,57-57,57H57c-32,0-57-25-57-57V120.5c0-31,25-57,57-57h398c32,0,57,26,57,57v327Z" style="fill: #ecf2f7;"/>
    <path d="M39,0h434c21.52,0,39,17.48,39,39v146H0V39C0,17.48,17.48,0,39,0Z" style="fill: %s;"/>
    <text transform="translate(22 146.5)" style="fill: #fff; font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; font-size: 100px;">%s</text>
    <text x="50%%" y="392.5" style="fill: #66757f; font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; font-size: 240px; text-anchor: middle">%d</text>
    <text x="50%%" y="472.5" style="fill: #66757f; font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; font-size: 64px; text-anchor: middle">%s</text>
    <text transform="translate(331.03 148.44)" style="fill: #fff; font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; font-size: 71.18px;">%d</text>
    </svg>
    `, colorScheme.Primary, dateInfo["month"], dateInfo["day"], dateInfo["weekday"], dateInfo["year"])
}

func generateTypeTwoSVG(color string, dateInfo map[string]any) string {
	colorScheme := getColorScheme(color)

	return fmt.Sprintf(`
    <svg id="dynamic_icon_type2" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">
    <path d="M512,447.5c0,32-25,57-57,57H57c-32,0-57-25-57-57V120.5c0-31,25-57,57-57h398c32,0,57,26,57,57v327Z" style="fill: #ecf2f7;"/>
    <path d="M39,0h434c21.52,0,39,17.48,39,39v146H0V39C0,17.48,17.48,0,39,0Z" style="fill: %s;"/>
    <text transform="translate(22 146.5)" style="fill: #fff; font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; font-size: 100px;">%s</text>
    <text x="50%%" y="420.5"  style="fill: #66757f; font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; font-size: 256px;text-anchor: middle">%d</text>
    <text transform="translate(331.03 148.44)" style="fill: #fff; font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; font-size: 71.18px;">%d</text>
    </svg>
    `, colorScheme.Primary, dateInfo["month"], dateInfo["day"], dateInfo["year"])
}

func generateTypeThreeSVG(color string, dateInfo map[string]any) string {
	colorScheme := getColorScheme(color)

	return fmt.Sprintf(`
    <svg id="dynamic_icon_type3" xmlns="http://www.w3.org/2000/svg" version="1.1" viewBox="0 0 512 512">
        <path class="cls-6" d="M512,447.5c0,32-25,57-57,57H57c-32,0-57-25-57-57V120.5c0-31,25-57,57-57h398c32,0,57,26,57,57v327Z" style="fill: #ecf2f7;"/>
        <path class="cls-1" d="M39,0h434c21.5,0,39,17.5,39,39v146H0V39C0,17.5,17.5,0,39,0Z" style="fill: %s;"/>
        <g style="fill: %s;">
            <circle  cx="468.5" cy="135" r="14"/>
            <circle  cx="468.5" cy="93" r="14"/>
            <circle  cx="425.5" cy="135" r="14"/>
            <circle  cx="425.5" cy="93" r="14"/>
            <circle  cx="382.5" cy="135" r="14"/>
            <circle  cx="382.5" cy="93" r="14"/>
        </g>
        <text transform="translate(22 146.5)" style="fill: #fff;font-size: 120px; font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; ">%d</text>
        <text x="50%%" y="410.5" style="fill: #66757f;font-size: 200px;text-anchor: middle;font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; ">%s</text>
    </svg>
    `, colorScheme.Primary, colorScheme.Secondary, dateInfo["year"], dateInfo["month"])
}

func generateTypeFourSVG(color string, dateInfo map[string]any) string {
	colorScheme := getColorScheme(color)

	return fmt.Sprintf(`
    <svg id="dynamic_icon_type4" xmlns="http://www.w3.org/2000/svg" version="1.1" viewBox="0 0 512 512">
        <path class="cls-6" d="M512,447.5c0,32-25,57-57,57H57c-32,0-57-25-57-57V120.5c0-31,25-57,57-57h398c32,0,57,26,57,57v327Z" style="fill: #ecf2f7;"/>
        <path class="cls-1" d="M39,0h434c21.5,0,39,17.5,39,39v146H0V39C0,17.5,17.5,0,39,0Z" style="fill: %s;"/>
        <g style="fill: %s;">
            <circle  cx="468.5" cy="135" r="14"/>
            <circle  cx="468.5" cy="93" r="14"/>
            <circle  cx="425.5" cy="135" r="14"/>
            <circle  cx="425.5" cy="93" r="14"/>
            <circle  cx="382.5" cy="135" r="14"/>
            <circle  cx="382.5" cy="93" r="14"/>
        </g>
        <text x="50%%" y="410.5" style="fill: #66757f;font-size: 200px;text-anchor: middle;font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; ">%d</text>
    </svg>
    `, colorScheme.Primary, colorScheme.Secondary, dateInfo["year"])
}

func generateTypeFiveSVG(color string, dateInfo map[string]any) string {
	colorScheme := getColorScheme(color)

	return fmt.Sprintf(`
    <svg id="dynamic_icon_type5" xmlns="http://www.w3.org/2000/svg" version="1.1" viewBox="0 0 512 512">
        <path class="cls-6" d="M512,447.5c0,32-25,57-57,57H57c-32,0-57-25-57-57V120.5c0-31,25-57,57-57h398c32,0,57,26,57,57v327Z" style="fill: #ecf2f7;"/>
        <path class="cls-1" d="M39,0h434c21.5,0,39,17.5,39,39v146H0V39C0,17.5,17.5,0,39,0Z" style="fill: %s;"/>
        <g style="fill: %s;">
            <circle  cx="468.5" cy="135" r="14"/>
            <circle  cx="468.5" cy="93" r="14"/>
            <circle  cx="425.5" cy="135" r="14"/>
            <circle  cx="425.5" cy="93" r="14"/>
            <circle  cx="382.5" cy="135" r="14"/>
            <circle  cx="382.5" cy="93" r="14"/>
        </g>
        <text transform="translate(22 146.5)" style="fill: #fff;font-size: 120px; font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; ">%d</text>
        <text x="50%%" y="410.5" style="fill: #66757f;font-size: 200px;text-anchor: middle;font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; ">%s</text>
    </svg>
    `, colorScheme.Primary, colorScheme.Secondary, dateInfo["isoYear"], dateInfo["week"])
}

func generateTypeSixSVG(color string, lang string, weekdayType string, dateInfo map[string]any) string {

	weekday := dateInfo["weekday"].(string)
	isWeekend := dateInfo["isWeekend"].(bool)

	var colorScheme ColorScheme
	if color == "" {
		if isWeekend {
			colorScheme = colorSchemes["blue"]
		} else {
			colorScheme = colorSchemes["red"]
		}
	} else {
		colorScheme = getColorScheme(color)
	}

	var fontSize float64
	switch weekdayType {
	case "1":
		fontSize = 690 / float64(len([]rune(weekday)))
	case "2":
		fontSize = 600 / float64(len([]rune(weekday)))
	case "3":
		fontSize = 720 / float64(len([]rune(weekday)))
	case "4":
		fontSize = 630 / float64(len([]rune(weekday)))
	default:
		fontSize = 750 / float64(len([]rune(weekday)))
	}

	return fmt.Sprintf(`
    <svg id="dynamic_icon_type6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">
    <path id="center" d="M512,447.5c0,32-25,57-57,57H57c-32,0-57-25-57-57V120.5c0-31,25-57,57-57h398c32,0,57,26,57,57v327Z" style="fill: #ecf2f7;"/>
    <path id="top" d="M39,0h434c21.5,0,39,14,39,31.2v116.8H0V31.2C0,14,17.5,0,39,0Z" style="fill: %s;"/>
    <g id="cirle" style="fill: %s;">
        <circle cx="468.5" cy="113.5" r="14"/>
        <circle cx="468.5" cy="71.5" r="14"/>
        <circle cx="425.5" cy="113.5" r="14"/>
        <circle cx="425.5" cy="71.5" r="14"/>
        <circle cx="382.5" cy="113.5" r="14"/>
        <circle cx="382.5" cy="71.5" r="14"/>
    </g>
    <text id="weekday" x="50%%"  y="65%%" style="fill: %s; font-size: %.2fpx; text-anchor: middle; dominant-baseline:middle; font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei';">%s</text>
    </svg>`, colorScheme.Primary, colorScheme.Secondary, colorScheme.Primary, fontSize, weekday)
}

func generateTypeSevenSVG(color string, lang string, dateInfo map[string]any) string {
	colorScheme := getColorScheme(color)

	diffDays := dateInfo["countDown"].(int)

	var tipText, diffDaysText string

	switch {
	case diffDays == 0:
		tipText = "Today"
		diffDaysText = "--"
	case diffDays > 0:
		tipText = "Left"
		diffDaysText = fmt.Sprintf("%d", diffDays)
	default:
		tipText = "Past"
		absDiffDays := -diffDays
		diffDaysText = fmt.Sprintf("%d", absDiffDays)
	}

	dayStr := "days"

	var fontSize float64
	switch {
	case len(diffDaysText) <= 3:
		fontSize = 240
	case len(diffDaysText) == 4:
		fontSize = 190
	case len(diffDaysText) == 5:
		fontSize = 140
	case len(diffDaysText) >= 6:
		fontSize = 780 / float64(len(diffDaysText))
	}

	return fmt.Sprintf(`
    <svg id="dynamic_icon_type7" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">
        <path id="bottom" d="M512,447.5c0,32-25,57-57,57H57c-32,0-57-25-57-57V120.5c0-31,25-57,57-57h398c32,0,57,26,57,57v327Z" style="fill: #ecf2f7;"/>
        <path id="top" d="M39,0h434c21.52,0,39,17.48,39,39v146H0V39C0,17.48,17.48,0,39,0Z" style="fill: %s;"/>
        <text id="year" transform="translate(46.1 78.92)" style="fill: #fff; font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; font-size: 60px;">%d</text>
        <text id="day" transform="translate(43.58 148.44)" style="fill: #fff; font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; font-size: 60px;">%s</text>
        <text id="passStr" transform="translate(400 148.44)" style="fill: #fff; text-anchor: middle;font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; font-size: 71.18px;">%s</text>
        <text id="diffDays" x="50%%" y="65%%" style="font-size: %.0fpx; fill: #66757f; text-anchor: middle; dominant-baseline:middle;font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; ">%s</text>
        <text id="dayStr" x="50%%" y="472.5" style="font-size: 64px; text-anchor: middle; fill: #66757f; font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei';">%s</text>
    </svg>`, colorScheme.Primary, dateInfo["year"], dateInfo["date"], tipText, fontSize, diffDaysText, dayStr)
}

func generateTypeEightSVG(color, content, id string) string {
	if strings.Contains(content, ".action{") {
		content = model.RenderDynamicIconContentTemplate(content, id)
	}

	colorScheme := getColorScheme(color)

	isChinese := regexp.MustCompile(`[\p{Han}]`).MatchString(content)
	var fontSize float64
	if isChinese {
		switch {
		case len([]rune(content)) == 1:
			fontSize = 320
		default:
			fontSize = 480 / float64(len([]rune(content)))
		}
	} else {
		switch {
		case len([]rune(content)) == 1:
			fontSize = 480
		case len([]rune(content)) == 2:
			fontSize = 300
		case len([]rune(content)) == 3:
			fontSize = 240
		default:
			fontSize = 750 / float64(len([]rune(content)))
		}
	}

	dy := "0%"
	if len([]rune(content)) == 1 {
		switch content {
		case "g", "p", "y", "q":
			dy = "-10%"
		case "j":
			dy = "-5%"
		default:
			dy = "0%"
		}
	}

	escapedContent := html.EscapeString(content)
	return fmt.Sprintf(`
    <svg id="dynamic_icon_type8" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">
        <path d="M39,0h434c20.97,0,38,17.03,38,38v412c0,33.11-26.89,60-60,60H60c-32.56,0-59-26.44-59-59V38C1,17.03,18.03,0,39,0Z" style="fill: %s;"/>
        <text x="50%%" y="55%%" dy="%s" style="font-size: %.2fpx; fill: #fff; text-anchor: middle; dominant-baseline:middle;font-family: -apple-system, BlinkMacSystemFont, 'Noto Sans', 'Noto Sans CJK SC', 'Microsoft YaHei'; ">%s</text>
	</svg>
    `, colorScheme.Primary, dy, fontSize, escapedContent)
}
