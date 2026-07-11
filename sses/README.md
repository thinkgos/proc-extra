# sses

基于频道订阅的通用`sse`服务.

## 参数说明

- `channel` 频道名称, 订阅该频道, 服务端以进行定向推送至该频道
- `userId` 用户id
- `sessionId` 会话id
- `sse.event` 事件结构
  - `id`: 事件id, 递增id, 标识该事件
  - `event`: 事件类型, 默认`message`, 可以区分不同的事件(可以在`channel`下继续细分)
  - `data`: 事件数据

## 功能

- 用户可订阅了感兴趣的`channel`
- 向订阅了指定的`channel`进行广播推送
  - 比如公告, 向订阅了该`channel`的所有用户的会话都能收到
- 向订阅了指定的`channel`的用户进行消息推送
  - 比如私信, 向订阅了该`channel`的指定用户的会话都能收到
- 向订阅了指定的`channel`的用户的会话进行定向消息推送
  - 比如文件上传, 服务端异步处理, 客户端使用sse等待处理结果
  - 处理思路:
    - 客户端上传前申请一个会话id(`sessionId`)
    - 开始订阅指定的`channel`
    - 在文件上传url传输该`sessionId`, 服务端异步处理完成后, 使用`userId`, `channel`, `sessionId`进行定向消息推送
