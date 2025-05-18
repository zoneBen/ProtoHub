package pinyin

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

var codeCompareMap = map[string]string{
	"开关机":      "air_switch_machine",
	"进风温度":     "air_inlet_air_temperature",
	"进风湿度":     "air_inlet_air_humidity",
	"出风温度":     "air_wind_temperature",
	"出风湿度":     "air_air_humidity",
	"总有功功率":    "el_total_active_power",
	"总功率因数":    "el_total_power_factor",
	"频率":       "el_frequency",
	"平均相电压":    "el_average_phase_voltage",
	"平均线电压":    "el_average_line_voltage",
	"平均电流":     "el_average_current",
	"总无功功率":    "el_total_reactive_power",
	"A-B线电压":   "el_a_to_b_line_voltage",
	"C-B线电压":   "el_c_to_b_line_voltage",
	"A-C线电压":   "el_a_to_c_line_voltage",
	"A相电流":     "el_a_phase_current",
	"B相电流":     "el_b_phase_current",
	"C相电流":     "el_c_phase_current",
	"A相有功功率":   "el_a_phase_active_power",
	"B相有功功率":   "el_b_phase_active_power",
	"C相有功功率":   "el_c_phase_active_power",
	"总视在功率":    "el_total_apparent_power",
	"有功功率(单相)": "pdu_active_power_single_phase",
	"功率因数(单相)": "pdu_power_factor_single_phase",
	"无功功率(单相)": "pdu_reactive_power_single_phase",
	"机组状态":     "air_machine_running_state",
	"压缩机":      "air_compressor_working_condition",
	"温度":       "device_temperature",
	"湿度":       "device_humidity",
	"烟感状态":     "device_smoking",
	"水浸状态":     "device_flooding",
	"单体电压":     "battery_voltage",
	"单体内阻":     "battery_resistance",
	"单体温度":     "battery_temperature",
	"组电压":      "battery_group_voltage",
	"组电流":      "battery_group_discharge_current",
	"平均温度":     "average_temperature_of_battery_pack",
	"单体平均电压":   "average_voltage_of_battery_pack",
	"平均单体内阻":   "pack_avg_resistance",
}

var replaceMap = map[string]string{
	"#":  "",
	"（":  "_",
	"(":  "_",
	"）":  "",
	")":  "",
	"-":  "_",
	"+":  "",
	"*":  "",
	"/":  "",
	"@":  "",
	"$":  "",
	"%":  "",
	"<":  "",
	">":  "",
	"?":  "",
	":":  "",
	"{":  "_",
	"}":  "",
	"!":  "",
	".":  "",
	"”":  "",
	"'":  "",
	";":  "",
	"^":  "",
	"~":  "",
	"&":  "",
	" ":  "",
	"\t": "",
	"\n": "",
	"\r": "",
	//０１２３４５６７８９
	"０": "0",
	"１": "1",
	"２": "2",
	"３": "3",
	"４": "4",
	"５": "5",
	"６": "6",
	"７": "7",
	"８": "8",
	"９": "9",
}

func GetCode(hans string) string {
	hans = strings.Replace(hans, " ", "", -1)
	hans = strings.Replace(hans, "\t", "", -1)
	hans = strings.Replace(hans, "\r", "", -1)
	hans = strings.Replace(hans, "\n", "", -1)
	c, ok := codeCompareMap[hans]
	if ok {
		return c
	}
	var out string
	hasher := sha256.New()
	hasher.Write([]byte(hans))
	hashBytes := hasher.Sum(nil)
	hashStr := hex.EncodeToString(hashBytes)
	shortHash := hashStr[:8]
	oo := LazyConvert(hans, nil)
	var y string
	var isSohrt bool
	if len(oo) > 9 {
		isSohrt = true
		for _, o := range oo {
			y = y + string(o[0])
		}
	} else {
		y = strings.Join(oo, "")
	}
	y = extractAndMoveNumberToEnd(y)
	if hasHashCode(hans) || isSohrt {
		out = fmt.Sprintf("%s_%s", y, shortHash)
	} else {
		out = y
	}
	return replaceStr(strings.ToLower(out))
}

func replaceStr(has string) string {
	for k, v := range replaceMap {
		has = strings.Replace(has, k, v, -1)
	}
	return has
}

func sd(hans string) []string {
	a := NewArgs()
	a.Style = Tone2
	pys := Pinyin(hans, a)
	var oo []string
	for _, i := range pys {
		oo = append(oo, i...)
	}
	return oo
}

func hasHashCode(has string) bool {
	for k := range replaceMap {
		if strings.Contains(has, k) {
			return true
		}
	}
	return false
}

// extractAndMoveNumberToEnd 提取字符串开头的数字，并将其放到字符串的末尾
func extractAndMoveNumberToEnd(s string) string {
	re := regexp.MustCompile(`^(\d+)(.*)`)
	match := re.FindStringSubmatch(s)
	if len(match) == 3 {
		return match[2] + match[1]
	}
	return s
}
