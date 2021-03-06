package adserver

import (
	"errors"
	"fmt"
	"github.com/mozilla-services/heka/message"
	"github.com/mozilla-services/heka/pipeline"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type YdaDecoder struct {
	format       string         //y.da 日志正则字符串
	regexp       *regexp.Regexp //y.da 日志正则
	queryKey     string         //y.da query字段
	debug        bool           //debug
	floatKeys    []string       //float keys
	logger       string
	fieldFilters url.Values
}

func (xd *YdaDecoder) Init(config interface{}) (err error) {
	format, _ := getConfString(config, "format")
	queryKey, _ := getConfString(config, "query")
	debug, _ := getConfString(config, "debug")
	floatkeys, _ := getConfString(config, "float_keys")
	xd.logger, _ = getConfString(config, "logger")
	ffilters, _ := getConfString(config, "field_filters")
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
	xd.floatKeys = strings.Split(floatkeys, " ")
	xd.fieldFilters, _ = url.ParseQuery(ffilters)
	fmt.Printf("config, format:%s, queryKey:%s, debug:%s, floatKeys:%+v, logger:%s, fieldFilters:%+v\n",
		format, queryKey, debug, xd.floatKeys, xd.logger, xd.fieldFilters)
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
			fmt.Printf("i:%d, name:%s, value:%s\n", i, name, fields[i])
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
					ffvalue := xd.fieldFilters.Get(k)
					if len(ffvalue) != 0 {
						//k is in field filters, check it
						isInFiledFilters := false
						for fk, ffvs := range xd.fieldFilters {
							if fk != k {
								continue
							}
							for _, ffv := range ffvs {
								if strings.Contains(vs[0], ffv) {
									isInFiledFilters = true
									break
								}
							}
						}
						if !isInFiledFilters {
							//k is in filed filters, but values is not in ffvalues, do not add value to message
							//continue
							vs[0] = "OTHERS"
						}
					}
					//只取相同key的第一个
					isFloatKey := false
					for _, fkey := range xd.floatKeys {
						if k == fkey {
							isFloatKey = true
							break
						}
					}
					if isFloatKey {
						v_float, e := strconv.ParseFloat(vs[0], 64)
						if e == nil {
							field := message.NewFieldInit(k, message.Field_DOUBLE, "")
							field.AddValue(v_float)
							pack.Message.AddField(field)
						}

					} else {
						field := message.NewFieldInit(k, message.Field_STRING, "")
						e := field.AddValue(vs[0])
						if e != nil {
							fmt.Printf("key:%s,value:%s, add value failed:%s\n", k, vs[0], e.Error())
						}
						pack.Message.AddField(field)
					}
				}
			}
		} else {
			field := message.NewFieldInit(name, message.Field_STRING, "")
			field.AddValue(fields[i])
			pack.Message.AddField(field)
		}

	}
	//set logger
	pack.Message.SetLogger(xd.logger)
	pack.Message.SetType("YDaDecoder")
	if xd.debug {
		fmt.Printf("message:%+v\n", *pack.Message)
	}

	return []*pipeline.PipelinePack{pack}, nil
}

func init() {
	pipeline.RegisterPlugin("YdaDecoder", func() interface{} {
		return new(YdaDecoder)
	})
}
