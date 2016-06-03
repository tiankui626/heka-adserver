package adserver

import (
	"errors"
	"fmt"
	"github.com/mozilla-services/heka/message"
	"github.com/mozilla-services/heka/pipeline"
	"net/url"
	"regexp"
	"strings"
)

type YdaDecoder struct {
	format   string         //y.da 日志正则字符串
	regexp   *regexp.Regexp //y.da 日志正则
	queryKey string         //y.da query字段
	debug    bool           //debug
}

func (xd *YdaDecoder) Init(config interface{}) (err error) {
	format, _ := getConfString(config, "format")
	queryKey, _ := getConfString(config, "query")
	debug, _ := getConfString(config, "debug")
	if len(format) == 0 {
		err = errors.New("format config is empty")
	} else {
		xd.format = format
		re := regexp.MustCompile(`\\\$([a-z_]+)(\\?(.))`).ReplaceAllString(
			regexp.QuoteMeta(format+" "), "(?P<$1>[^$3]*)$2")
		xd.regexp = regexp.MustCompile(fmt.Sprintf("^%v$", strings.Trim(re, " ")))
	}
	xd.queryKey = queryKey
	xd.debug = (debug == "1")
	fmt.Printf("config, format:%s, queryKey:%s, debug:%s\n", format, queryKey, debug)
	return
}

func (xd *YdaDecoder) Decode(pack *pipeline.PipelinePack) (packs []*pipeline.PipelinePack, err error) {
	line := strings.TrimSpace(pack.Message.GetPayload())

	fields := xd.regexp.FindStringSubmatch(line)
	if fields == nil {
		err = fmt.Errorf("access log line '%v' does not match given format '%v'", line, xd.regexp)
		return
	}

	for i, name := range xd.regexp.SubexpNames() {
		if i == 0 {
			continue
		}
		if xd.debug {
			fmt.Printf("i:%d, name:%s, value:%s\n", fields[i])
		}
		if name == xd.queryKey {
			//parse query
			qs := strings.Split(fields[i], "?")
			var query string
			if len(qs) == 2 {
				//request_path?a=b&c=d
				field := message.NewFieldInit("request_path", message.Field_STRING, "")
				field.AddValue(qs[0])
				pack.Message.AddField(field)
				query = qs[1]
			} else if len(qs) == 1 {
				//a=b&c=d
				query = qs[0]
			}
			values, err := url.ParseQuery(query)
			if err == nil {
				for k, vs := range values {
					//只取相同key的第一个
					field := message.NewFieldInit(k, message.Field_STRING, "")
					field.AddValue(vs[0])
					pack.Message.AddField(field)
				}
			}
		} else {
			field := message.NewFieldInit(name, message.Field_STRING, "")
			field.AddValue(fields[i])
			pack.Message.AddField(field)
		}

	}

	return []*pipeline.PipelinePack{pack}, nil
}

func init() {
	pipeline.RegisterPlugin("YdaDecoder", func() interface{} {
		return new(YdaDecoder)
	})
}
