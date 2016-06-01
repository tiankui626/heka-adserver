package adserver

import (
	"errors"
	"fmt"
	"github.com/mozilla-services/heka/message"
	"github.com/mozilla-services/heka/pipeline"
	"regexp"
	"strings"
)

type YdaDecoder struct {
	format   string         //y.da 日志正则字符串
	regexp   *regexp.Regexp //y.da 日志正则
	queryKey string         //y.da query字段
}

func (xd *YdaDecoder) Init(config interface{}) (err error) {
	format, _ := getConfString(config, "format")
	queryKey, _ := getConfString(config, "query")
	if len(format) == 0 {
		err = errors.New("format config is empty")
	} else {
		xd.format = format
		re := regexp.MustCompile(`\\\$([a-z_]+)(\\?(.))`).ReplaceAllString(
			regexp.QuoteMeta(format+" "), "(?P<$1>[^$3]*)$2")
		xd.regexp = regexp.MustCompile(fmt.Sprintf("^%v$", strings.Trim(re, " ")))
	}
	xd.queryKey = queryKey
	return
}

func (xd *YdaDecoder) Decode(pack *pipeline.PipelinePack) (packs []*pipeline.PipelinePack, err error) {
	line := pack.Message.GetPayload()

	fields := xd.regexp.FindStringSubmatch(line)
	if fields == nil {
		err = fmt.Errorf("access log line '%v' does not match given format '%v'", line, xd.regexp)
		return
	}

	for i, name := range xd.regexp.SubexpNames() {
		if i == 0 {
			continue
		}
		field := message.NewFieldInit(name, message.Field_STRING, "")
		field.AddValue(fields[i])
		pack.Message.AddField(field)
	}

	return []*pipeline.PipelinePack{pack}, nil
}

func init() {
	pipeline.RegisterPlugin("YdaDecoder", func() interface{} {
		return new(YdaDecoder)
	})
}
