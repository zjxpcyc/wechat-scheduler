# WechatScheduler 微信相关定时器
因为业务需求, 微信 token, ticket 等获取需要抽离出来, 单独处理。以实现业务端的平滑重启。

## 目标
- [x] 支持 Job 动态注册
- [ ] 支持基本的访问校验
- [ ] 本系统可平滑重启

## 支持任务列表
1. 公众号 access_token [官方说明](https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421140183)
2. 公众号 jsapi_ticket [官方说明](https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421141115)
3. 公众号 Oauth2 access_token [官方说明](https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421140842)
4. 第三方平台 component_access_token [官方说明](https://open.weixin.qq.com/cgi-bin/showdocument?action=dir_list&t=resource/res_list&verify=1&id=open1453779503&token=&lang=zh_CN)
5. 第三方平台 authorizer_access_token [官方说明](https://open.weixin.qq.com/cgi-bin/showdocument?action=dir_list&t=resource/res_list&verify=1&id=open1453779503&token=&lang=zh_CN)

## 使用说明
本系统独立运行, 与业务系统通过 http 的方式进行交互。 目前支持的接口如下:


### `/registe` 注册任务

参数需要通过 http body 传入 json 数据。 `Content-type: application/json`

参数格式如下:
```json
{
	"appid": "",
	"appsecret": "",
	"tasks": [
		{
			"type": 0,
			"notify": "",
			"params": "",
		},
		...
	]
}
```
其中:

`appid`, `appsecret` 不做解释了，必填字段

`type`: 任务类型, 目前支持的有
| 值   |      任务      |
|----------|:-------------:|
| 0 | 公众号 access_token |
| 1 | 公众号 jsapi_ticket |
| 2 | 公众号 Oauth2 access_token |
| 3 | 第三方平台 component_access_token |
| 4 | 第三方平台 authorizer_access_token |

`notify`: 为业务系统提供的一个回调地址。本系统每完成一次任务更新, 就会对地址进行回调访问，并传输回调结果。比如此值为 `http://somedomain.com/foo/bar`, 那么本系统会对此接口进行 `POST` 访问, 并带上 `appid`、`type` query 参数。同时 body 的 `Content-type: application/json` 内容为相关任务的结果。比如 access_token 就会返回 
```json
{
  "access_token":"ACCESS_TOKEN",
  "expires_in":7200
}
```
如果, 不需要实时通知业务系统, 此参数可以为空。


`params`: 在部分的任务循环进行的时候, 有时候需要一些额外的值, 比如 `jsapi_ticket`, 这个任务需要 `access_token`、第三方平台的 `component_access_token`, 需要十分钟一次的 `ticket`。 这种参数需要业务系统传入。请求的方式为 `GET`, 同时会加入 `appid`、`type` query 参数。业务系统通过 http body 返回 `Content-type: application/json` 的内容。 比如返回 `ticket` 那么反馈结果需要为
```json
{
  "component_verify_ticket":"xxxxxxxxxx"
}
```
如果，有任务不需要额外参数的，则此参数可以不传。部分需要传输的可以视情况而定，是否传入。比如如果注册了 `access_token` 任务, 那么注册 `jsapi_ticket` 任务时就不需要传入了，因为系统会先从 `access_token` 任务中获取需要的值。

`params` 在以下几种情况下是必须要设置的:

1. 注册了 `jsapi_ticket` 任务, 但是没有注册 `access_token` 任务, 需要传入
```json
{
  "access_token": "xxxx"
}
```

2. `Oauth2 access_token` 任务需要传入
```json
{
  "refresh_token": "xxxx"
}
```
如果, 之前已经注册过该任务, 并且 `refresh_token` 未过期， 则可以不用传入。

3. 第三方平台 `component_verify_ticket` 任务 需要传入
```json
{
  "component_verify_ticket":"xxxxxxxxxx"
}
```
4. 第三方平台 `authorizer_access_token` 任务需要转入
```json
{
  "authorizer_refresh_token":"xxxxxxxxxx"
}
```


### /task/:appid/:type 获取 type 任务结果

一般如果在注册任务的时候，注册了 `notify` 地址, 那么这个接口是不需要的。如果没有注册，可以通过这个接口进行获取。

返回内容同样为 `Content-type: application/json`

格式如下:
```json
{
    "code": 200,
    "message": "",
    "result": ...
}
```

`code` 为 200 时, 代表结果正常。
非 200 时代表有错误。 `message` 为错误提示。`result` 为结果正常时的期望返回结果，目前所有的都是 `string` 类型。

此接口与 `notify` 注册的结果会不同。 此接口 `result` 只会返回最终期望结果，比如 `access_token` 任务只会返回字符串结果，并不会将微信返回的 json 整个返回。

## 系统启动
go build 结束之后会生成可执行文件。比如默认生成一个 `wechat-scheduler` 文件。

执行方式如下:
```bash
./wechat-scheduler -p=8080
```

`-p` 是设置启动端口, 默认是 9001

`-v` 是查询当前系统版本号
