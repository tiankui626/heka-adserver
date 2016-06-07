# heka-adserver

heka日志decoder插件，用于解析nginx日志和kv格式的日志

## YdaDecoder
nginx日志decoder插件

#### 参数

    * format: nginx日志格式
    * query: nginx日志请求参数字段
    * float_keys: 解析日志的value为浮点类型的keys，用空格分隔
    * logger: 自定义日志名
    * field_filters: kv过滤，形式是“k1=v1&&k1=v11&k2=v2”,表示如果k1不为v1或者v11，过滤掉该字段，不进入下一个筛选逻辑，k2同理，可以多个key相同
    * debug: 0关闭debug日志，1打开debug日志

#### 示例

配置示例

```
[ydaLogDecoder]
#decode 类型设置为YdaDecoder
type = "YdaDecoder"
#format nginx日志格式已|分隔
format = "$remote_addr|$time_iso|$request_method|$request_uri|$status|$body_bytes_sent|$http_referer|$http_x_forwarded_for|$http_user_agent|$request_time|$upstream_response_time|$request_body|$http_host|$hostname"
#query nginx字段中表示请求参数字段，会调用url parser接口解析请求参数，解析成kv对
#request_uri='urlpath?a=1&b=2',会解析成一个request_path=urlpath,a=1,b=2
query= "request_uri"
#debug 是否开启debug
debug = "0"
# field_filters kv过滤，降低维度
field_filters = "appver=4.7.0&appver=4.6.9"
```
日志示例
```
10.100.2.253|2016-06-02T20:15:02+08:00|GET|/app/impression?d=AC18F74E-D77F-D1D4-D93D-E65CC4CB789B&v=3194432&ct=2734&cd=3353&s=94315&b=4388&t=1800&o=317&sp=94315,3353,2734&id=D7CA265C-6E1A-BD77-B13A-05C0923218AC&ip=123.180.3.16&tn=2&hid=1&video_len=5160&os=3&channel=1&adtype=1&isintact=1&adtotal=5&tag=repeat_0&rnd=1464869716478|200|43|http://player.hunantv.com/mango-tv3-main/MangoTV_3.swf?js_function_name=vjjFlash&video_id=3194432&skin_swf_url=http://player.hunantv.com/mango-tv3-skin/MangoTV_Skin_3.swf&player_swf_url=http://player.hunantv.com/mango-tv3-player/MangoTV_Player_3.swf&statistics_swf_url=http://player.hunantv.com/mango-tv3-statistics/MangoTV_Statistics_3.swf&mapd_swf_url=http://player.hunantv.com/mango-tv3-mapd/MangoTV_Mapd_3.swf&statistics_bigdata_bid=1&randkey=20160303|123.180.3.16|Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.101 Safari/537.36|0.000|-|-|y.da.hunantv.com|yg-ad.viki.10-100-2-51
```

## XdaDecoder

kv格式日志decoder插件

#### 参数

    * format: kv格式关键词，只有包含该关键词，才会解析
    * logger: 自定义日志名
    * field_filters: kv过滤，形式是“k1=v1&&k1=v11&k2=v2”,表示如果k1不为v1或者v11，过滤掉该字段，不进入下一个筛选逻辑，k2同理，可以多个key相同
    * debug: 0关闭debug日志，1打开debug日志
    * spliter: kv日志分隔符，默认为一个空格

#### 示例

配置示例

```
[xdaLogDecoder]
type = "XdaDecoder"
#包含INFO字段
format = "INFO"
debug = "0"
# version字段包含4.7.0或者4.6.9
field_filters = "version=4.7.0&version=4.6.9"
spliter = " "
```
日志示例
```
2016/06/07 11:00:01.747 context.go:46 >INFO - Uuid=V1Y4sUID81ArKP3x cost=16.449519ms func=VdPlayer method=POST version=4.6.8 osversion=android_4.4.4 isvip=0 ispay=0 ispreview=0 pid=4580 vid=1725145 hid=150115 ip=223.102.121.42 type=32 os=0 city=1156210000 vic_cost=2.005594ms isintact=1 video_len=3421 channel=2 uid=868436025590206 freq_cost=427.437µs adlen=3 repeated_switch=1 pmp_cost=10.481406ms pmplen=0 adinfo={c:213,aid:106752,mid:4560,cid:2717,adtype:1,order:1,time:15,trigger:0} adinfo={c:327,aid:106750,mid:3687,cid:2984,adtype:1,order:2,time:15,trigger:0} adinfo={c:361,aid:106754,mid:4684,cid:3357,adtype:4,order:1,time:0,trigger:0}
```