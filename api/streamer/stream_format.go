package streamer

import (
	"github.com/miekg/dns"
	"strings"
)

type StreamRequestLog struct {
	Name       string `json:"name"`
	Typ        string `json:"type"`
	SourceHost string `json:"src_host"`
	SourceIP   string `json:"src_ip"`
}

type StreamRecordLog struct {
	Typ     string `json:"type"`
	TTL     int    `json:"ttl"`
	Content string `json:"content"`
}

type StreamResponseLog struct {
	Code    string            `json:"code"`
	Records []StreamRecordLog `json:"records"`
}

type StreamLog struct {
	Created  int64             `json:"created_at"`
	Request  StreamRequestLog  `json:"request"`
	Response StreamResponseLog `json:"response"`
}

// dns response to stream log
func responseToStreamLog(created_at int64, r *dns.Msg, src_host string, src_ip string) StreamLog {
	var streamLog StreamLog
	streamLog.Response.Code = dns.RcodeToString[r.Rcode]
	return StreamLog{
		Created: created_at,
		Request: StreamRequestLog{
			Name:       r.Question[0].Name,
			Typ:        dns.TypeToString[r.Question[0].Qtype],
			SourceHost: src_host,
			SourceIP:   src_ip,
		},
		Response: StreamResponseLog{
			Code:    dns.RcodeToString[r.Rcode],
			Records: recordLog(r.Answer),
		},
	}
}

func answerToLog(answer dns.RR) StreamRecordLog {
	content := answer.String()
	// this is kind of silly but I think it's an ok way to do it
	content = strings.Join(strings.Fields(content)[4:], " ")
	return StreamRecordLog{
		Typ:     dns.TypeToString[answer.Header().Rrtype],
		TTL:     int(answer.Header().Ttl),
		Content: content,
	}
}

func recordLog(answers []dns.RR) []StreamRecordLog {
	answerLogs := []StreamRecordLog{}
	for _, answer := range answers {
		answerLogs = append(answerLogs, answerToLog(answer))
	}
	return answerLogs
}
