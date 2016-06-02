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
	"time"
)

var (
	replacer *strings.Replacer
)

type XdaDecoder struct {
	format string         //x.da 日志正则字符串
	regexp *regexp.Regexp //x.da 日志正则
	debug  bool           //debug
}

func getConfString(config interface{}, key string) (string, error) {
	var (
		fieldConf interface{}
		ok        bool
	)
	conf := config.(pipeline.PluginConfig)
	if fieldConf, ok = conf[key]; !ok {
		return "", errors.New(fmt.Sprintf("No '%s' setting", key))
	}
	value, ok := fieldConf.(string)
	if ok {
		return value, nil
	}
	return "", nil
}

func timeParser(t string) (ms float64, err error) {
	var d time.Duration
	d, err = time.ParseDuration(t)
	if err != nil {
		fmt.Printf("time parser err:%s\n", err.Error())
		return
	}
	ms = float64(d) / float64(time.Millisecond)
	return
}

func (xd *XdaDecoder) Init(config interface{}) (err error) {
	format, _ := getConfString(config, "format")
	debug, _ := getConfString(config, "debug")
	fmt.Printf("xdadecoder init, format:%s, debug:%s\n", format, debug)
	if len(format) == 0 {
		err = errors.New("format config is empty")
	} else {
		xd.format = format
		xd.regexp = regexp.MustCompile(format)
	}
	xd.debug = (debug == "1")
	replacer = strings.NewReplacer("{", "", "}", "", ",", "&")
	return
}

func (xd *XdaDecoder) Decode(pack *pipeline.PipelinePack) (packs []*pipeline.PipelinePack, err error) {
	line := pack.Message.GetPayload()
	if xd.debug {
		fmt.Printf("decode line:%s\n", line)
	}
	if !xd.regexp.Match([]byte(line)) {
		fmt.Printf("regexp error:%s\n", line)
		return
	}
	parsedLine := strings.Replace(line, " ", "&", -1)
	values, err := url.ParseQuery(parsedLine)
	if err != nil {
		fmt.Printf("parse line error:%s\n", err.Error())
		return
	}
	for k, vs := range values {
		//parse non adinfo keys
		if "adinfo" == k {
			continue
		}
		field := message.NewFieldInit(k, message.Field_STRING, "")
		for _, v := range vs {
			if strings.Contains(k, "cost") {
				f, err := timeParser(v)
				if err != nil {
					field.AddValue(v)
				} else {
					field.AddValue(strconv.FormatFloat(f, 'f', 2, 64))
				}
			} else {
				field.AddValue(v)
			}

		}
		pack.Message.AddField(field)
	}
	if xd.debug {
		fmt.Printf("message:%+v\n", *(pack.Message))
	}
	//add non adinfo pack to packs
	packs = append(packs, pack)
	//parse adinfo keys
	for k, vs := range values {
		if "adinfo" != k {
			continue
		}
		for _, adinfo := range vs {
			//{c:394,aid:106752,mid:4642,cid:3848,adtype:1,order:2,time:15,trigger:0}
			apack := pipeline.NewPipelinePack(pack.RecycleChan)
			apack.Message = message.CopyMessage(pack.Message)
			apack.Message.SetType("adinfo")
			parseAdinfo(adinfo, apack.Message)
			packs = append(packs, apack)

			if xd.debug {
				fmt.Printf("adinfo message:%+v\n", *(apack.Message))
			}
		}
	}

	return packs, nil
}

func parseAdinfo(adinfo string, msg *message.Message) {
	parsedAdinfo = replacer.Replace(adinfo)
	values, err := url.ParseQuery(parsedAdinfo)
	if err != nil {
		fmt.Printf("parse line error:%s\n", err.Error())
		return
	}
	for k, vs := range values {
		field := message.NewFieldInit(k, message.Field_STRING, "")
		for _, v := range vs {
			field.AddValue(v)
		}
		msg.AddField(field)
	}

}

func init() {
	pipeline.RegisterPlugin("XdaDecoder", func() interface{} {
		return new(XdaDecoder)
	})
}
