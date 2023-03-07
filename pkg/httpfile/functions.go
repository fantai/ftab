package httpfile

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func nowWithOffset(offest, option string) time.Time {
	now := time.Now()
	if offest != "" {
		n, _ := strconv.ParseInt(offest, 10, 32)
		d := option
		switch d {
		case "y":
			now.AddDate(int(n), 0, 0)
			break
		case "M":
			now.AddDate(0, int(n), 0)
			break
		case "d":
			now.AddDate(0, 0, int(n))
			break
		case "h":
			now.Add(time.Duration(n) * time.Hour)
			break
		case "m":
			now.Add(time.Duration(n) * time.Minute)
			break
		case "s":
			now.Add(time.Duration(n) * time.Second)
			break
		case "ms":
			now.Add(time.Duration(n) * time.Millisecond)
			break
		}

	}
	return now
}

func layoutFromTimeFormat(s string) string {
	key, _ := regexp.Compile(`[YMDHmsSZd]+`)
	layout := key.ReplaceAllFunc([]byte(s), func(word []byte) []byte {
		switch string(word) {
		case "YYYY":
			return []byte("2006")
		case "YY":
			return []byte("06")
		case "M":
			return []byte("1")
		case "MM":
			return []byte("01")
		case "MMM":
			return []byte("Jan")
		case "MMMM":
			return []byte("January")
		case "D":
			return []byte("2")
		case "DD":
			return []byte("02")
		case "ddd":
			return []byte("Mon")
		case "dddd":
			return []byte("Monday")
		case "H":
			return []byte("3")
		case "HH":
			return []byte("03")
		case "m":
			return []byte("4")
		case "mm":
			return []byte("04")
		case "s":
			return []byte("5")
		case "ss":
			return []byte("05")
		case "SSS":
			return []byte("000")
		case "Z":
			return []byte("-07:00")
		case "ZZ":
			return []byte("-0700")
		default:
			return word
		}
	})
	return string(layout)
}

func timeLayout(s string) string {
	switch s {
	case "rfc3339":
		return time.RFC3339
	case "rfc1123":
		return time.RFC1123
	case "iso8601":
		return "2006-01-02 03:04:05,.000"
	default:
		return layoutFromTimeFormat(s)
	}
}

func funTimestamap(args []string) string {
	offset := ""
	option := ""
	if len(args) == 3 {
		offset = args[1]
		option = args[2]
	}
	now := nowWithOffset(offset, option)
	return fmt.Sprintf("%d", now.Unix())
}

func funTimestamapms(args []string) string {
	offset := ""
	option := ""
	if len(args) == 3 {
		offset = args[1]
		option = args[2]
	}
	now := nowWithOffset(offset, option)
	return fmt.Sprintf("%d", now.UnixNano()/1e6)
}

func funDateTime(args []string) string {
	offset := ""
	option := ""
	layout := "rfc3339"
	if len(args) == 4 {
		offset = args[2]
		option = args[3]
	}
	if len(args) > 1 {
		layout = args[1]
	}
	now := nowWithOffset(offset, option)
	layout = timeLayout(layout)
	return now.Format(layout)
}

func funLocalDateTime(args []string) string {
	offset := ""
	option := ""
	layout := "rfc3339"
	if len(args) == 4 {
		offset = args[2]
		option = args[3]
	}
	if len(args) > 1 {
		layout = args[1]
	}
	now := nowWithOffset(offset, option).Local()
	layout = timeLayout(layout)
	return now.Format(layout)
}

func funRandomInt(args []string) string {
	n := int64(0)
	if len(args) == 3 {
		minN, _ := strconv.ParseInt(args[1], 10, 64)
		maxN, _ := strconv.ParseInt(args[2], 10, 64)
		n = rand.Int63n(maxN) + minN

	} else {
		n = rand.Int63()
	}

	return fmt.Sprintf("%d", n)
}

var fileListCache = &sync.Map{}

func funRandomFromFile(args []string) string {
	if len(args) < 2 {
		return ""
	}

	filePath := args[1]
	if v, ok := fileListCache.Load(filePath); ok {
		lists, ok := v.([]string)
		if ok && len(lists) > 0 {
			return lists[rand.Intn(len(lists))]
		}
	} else {
		lists, err := readFileLines(filePath)
		if err == nil && len(lists) > 0 {
			fileListCache.Store(filePath, lists)
			return lists[rand.Intn(len(lists))]
		}
	}
	return ""
}

func readFileLines(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
