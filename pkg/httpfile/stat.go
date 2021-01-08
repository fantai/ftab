package httpfile

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"time"
)

// Stat is the file execute time
type Stat struct {
	Requests      int
	TimeConsuming float64
	BytesSend     int
	BytesReceived int
	Successed     int
	Failed        int
}

// Report is the statatics of results
type Report struct {
	TotalRequests    int
	Currency         int
	Successed        int
	Failed           int
	TotalSend        int
	TotalRecv        int
	TotalTimeUsed    float64
	SendSpeed        float64
	RecvSpeed        float64
	AvgTimeUsed      float64
	RequestPerSecond int
	MaxTimeUsed      float64
	MinTimeUsed      float64
	P50TimeUsed      float64
	P75TimeUsed      float64
	P90TimeUsed      float64
	P95TimeUsed      float64
	P99TimeUsed      float64
	Stats            []Stat
}

func pos(n int, p float64) int {
	pn := int(float64(n) * p)
	if pn >= n {
		pn = n - 1
	}
	return pn
}

func removeFailed(stats []Stat) []Stat {
	i := 0
	for _, s := range stats {
		if s.Successed > 0 {
			stats[i] = s
			i++
		}
	}
	return stats[0:i]
}

// ReportStat generate report for stats
func ReportStat(stats []Stat, totalTimeUsed float64) Report {
	var report Report

	if len(stats) == 0 {
		return report
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].TimeConsuming < stats[j].TimeConsuming
	})

	sumTimeUsed := 0.0
	report.TotalRequests = len(stats)
	for _, s := range stats {
		report.TotalRecv = report.TotalRecv + s.BytesReceived
		report.TotalSend = report.TotalSend + s.BytesSend
		report.Successed += s.Successed
		report.Failed += s.Failed
		sumTimeUsed = sumTimeUsed + s.TimeConsuming
	}
	report.TotalTimeUsed = totalTimeUsed
	report.RecvSpeed = float64(report.TotalRecv) / float64(report.TotalTimeUsed)
	report.SendSpeed = float64(report.TotalSend) / float64(report.TotalTimeUsed)

	report.AvgTimeUsed = sumTimeUsed / float64(report.Successed)
	report.RequestPerSecond = int((1.0 / report.TotalTimeUsed) * float64(report.Successed))

	ss := removeFailed(stats)
	report.MinTimeUsed = ss[0].TimeConsuming
	report.MaxTimeUsed = ss[len(ss)-1].TimeConsuming
	report.P50TimeUsed = ss[pos(len(ss), 0.50)].TimeConsuming
	report.P75TimeUsed = ss[pos(len(ss), 0.75)].TimeConsuming
	report.P90TimeUsed = ss[pos(len(ss), 0.90)].TimeConsuming
	report.P95TimeUsed = ss[pos(len(ss), 0.95)].TimeConsuming
	report.P99TimeUsed = ss[pos(len(ss), 0.99)].TimeConsuming

	report.Stats = stats

	return report
}

// PlainOutput is plain output of report
func PlainOutput(report *Report, w io.Writer) {

	format := "%-20v: %v\n"

	fmt.Fprintf(w, format, "Total Requests", report.TotalRequests)
	fmt.Fprintf(w, format, "Currency", report.Currency)
	fmt.Fprintf(w, format, "Successed", report.Successed)
	fmt.Fprintf(w, format, "Failed", report.Failed)
	fmt.Fprintf(w, format, "Time Used", report.TotalTimeUsed)
	fmt.Fprintf(w, format, "Reqeusts Per Second", report.RequestPerSecond)
	fmt.Fprintf(w, format, "Send Speed", report.SendSpeed)
	fmt.Fprintf(w, format, "Recv Speed", report.RecvSpeed)

	fmt.Fprintln(w)

	fmt.Fprintf(w, format, "Avg Time Used", report.AvgTimeUsed)
	fmt.Fprintf(w, format, "Min Time Used", report.MinTimeUsed)
	fmt.Fprintf(w, format, "Max Time Used", report.MaxTimeUsed)

	fmt.Fprintln(w)

	fmt.Fprintf(w, format, "P50 Time Used", report.P50TimeUsed)
	fmt.Fprintf(w, format, "P75 Time Used", report.P75TimeUsed)
	fmt.Fprintf(w, format, "P90 Time Used", report.P90TimeUsed)
	fmt.Fprintf(w, format, "P95 Time Used", report.P95TimeUsed)
	fmt.Fprintf(w, format, "P99 Time Used", report.P99TimeUsed)
}

func thoundsNumber(n int) string {
	in := strconv.Itoa(n)
	numOfDigits := len(in)
	if n < 0 {
		numOfDigits-- // First character is the - sign (not a digit)
	}
	numOfCommas := (numOfDigits - 1) / 3

	out := make([]byte, len(in)+numOfCommas)
	if n < 0 {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}

func humanDuration(val float64) string {
	d := time.Duration(val * float64(time.Second))
	return d.String()
}

func bytesNumber(val float64) string {
	unit := "B"
	if val < 1024*1024 {
		unit = "K"
		val = val / (1024)
	} else if val < 1024*1024*1024 {
		unit = "M"
		val = val / (1024 * 1024)
	} else {
		unit = "G"
		val = val / (1024 * 1024 * 1024)
	}
	return fmt.Sprintf("%.3f%s", val, unit)
}

// HumanOutput is plain output of report
func HumanOutput(report *Report, w io.Writer) {

	format := "%-20v: %v%v\n"

	fmt.Fprintf(w, format, "Total Requests", thoundsNumber(report.TotalRequests), "")
	fmt.Fprintf(w, format, "Currency", thoundsNumber(report.Currency), "")
	fmt.Fprintf(w, format, "Successed", thoundsNumber(report.Successed), "")
	fmt.Fprintf(w, format, "Failed", thoundsNumber(report.Failed), "")

	fmt.Fprintln(w)

	fmt.Fprintf(w, format, "Time Used", humanDuration(report.TotalTimeUsed), "")
	fmt.Fprintf(w, format, "Reqeusts Per Second", thoundsNumber(report.RequestPerSecond), "/S")
	fmt.Fprintf(w, format, "Send Speed", bytesNumber(report.SendSpeed), "/S")
	fmt.Fprintf(w, format, "Recv Speed", bytesNumber(report.RecvSpeed), "/S")

	fmt.Fprintln(w)

	fmt.Fprintf(w, format, "Avg Time Used", humanDuration(report.AvgTimeUsed), "")
	fmt.Fprintf(w, format, "Min Time Used", humanDuration(report.MinTimeUsed), "")
	fmt.Fprintf(w, format, "Max Time Used", humanDuration(report.MaxTimeUsed), "")

	fmt.Fprintln(w)

	fmt.Fprintf(w, format, "P50 Time Used", humanDuration(report.P50TimeUsed), "")
	fmt.Fprintf(w, format, "P75 Time Used", humanDuration(report.P75TimeUsed), "")
	fmt.Fprintf(w, format, "P90 Time Used", humanDuration(report.P90TimeUsed), "")
	fmt.Fprintf(w, format, "P95 Time Used", humanDuration(report.P95TimeUsed), "")
	fmt.Fprintf(w, format, "P99 Time Used", humanDuration(report.P99TimeUsed), "")
}
