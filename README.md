<p align="center">
  <img src="assets/social-preview.png" alt="WorkBuddy2API Panel" width="100%">
</p>

# WorkBuddy2API Panel

一个自托管的账号管理面板和 OpenAI 兼容网关。用 Web 面板管理自己的 WorkBuddy、CodeBuddy、Trae、Qoder、MiniMax、豆包和华为云 CodeArts 账号，也可以接入 Google Gemini 官方 API Key。

[下载发行版](https://github.com/xu20060420fan-cell/workbuddy2api-trae-qoder-minimax-panel/releases/latest) · [使用前必读](#使用前必读) · [快速开始](#快速开始) · [调用 API](#调用-api) · [更新日志](CHANGELOG.md) · [支持开发](#支持开发)

## 使用前必读

> **警惕中转站：天下没有便宜的午餐。**
>
> 市面上大量「低价 API 中转 / 镜像站 / 共享号池」并不值得信任。你看到的模型名可能只是个壳：背后真正跑的可能是 DeepSeek 之类的廉价模型，也可能是从 WorkBuddy 这类客户端、甚至从网页版对话里反向扒出来的流量，再包装成「官方 API」卖给你。
>
> 这类服务随时可能跑路、记录并转卖你的提示词、篡改返回内容，或者把你的账号凭据二次利用。**不要把账号、密码、Cookie、API Key 交给任何「代挂」「代充」「共享池」。**
>
> 本项目是**自托管**工具：凭据、会话和请求都在你自己的机器上，直连官方接口。想要稳定和隐私，就自己部署；模型能力请以官方为准，不要相信任何「一块钱一亿 token」的说法。

> **关于华为云 CodeArts（码道）**
>
> 华为这条线的接口改动频繁，登录态和风控都比其他平台麻烦；官方资料更偏「方案 / 流程说明」，落地细节很少，维护成本明显高于其他平台。作者的精力有限，后续对它只做「能修就修」：修得动就修，修不动就保持现状，不再为它单独堆兼容层。
>
> 需要的用户可以自己 clone 回本地改。提 issue 时请附上**真实请求与响应**（脱敏后），不要只贴一段方案或者「应该这样做」的推测——这类内容对定位问题没有帮助。

> **关于这个项目**
>
> - 个人自用工具开源，不是商业产品：没有 SLA、没有客服，也不承诺「永久可用」。平台接口一变，能跟就跟，跟不上就是跟不上。
> - 面板只替你点「能点的按钮」。需要真人完成的（扫码登录、客户端内对话、关注公众号满 24 小时等）不会被伪造，也不会为了好看假装完成。
> - 仅支持接入**本人有权使用**的账号。共享账号池、代挂、对外提供付费服务都不在支持范围内，风险自负。
> - 遇到问题先看 [CHANGELOG.md](CHANGELOG.md) 和 issue 历史；提问请带上版本号、平台、完整报错和复现步骤。
>
> 如果这个项目确实帮到了你，欢迎点个 Star，或者到文末扫码打赏——金额随意，心意最重要。

## 平台与功能

| 平台 | 面板功能 | API 通道 |
| --- | --- | --- |
| WorkBuddy / CodeBuddy 国服 | OAuth 登录、账号池、签到、余额和成长任务 | `cn:` |
| WorkBuddy 国际版 | OAuth 登录、账号池、余额、成长路径与领奖 | `global:` |
| Trae | IDE 账号管理、切号、签到和积分 | `trae:` |
| Qoder | IDE 账号管理、切号、签到和积分 | |
| MiniMax Code | 浏览器授权、账号管理、签到和积分 | `minimax:` |
| 豆包 Doubao | 会话管理和签到 | |
| 华为云 CodeArts 码道 | 会话管理、签到和积分总览 | |
| Google Gemini | 官方 API Key 分组管理 | 默认 `gemini:` |

- 网关提供聊天补全、Responses 和模型列表，支持 SSE 流式响应。账号池支持轮转、会话粘性、冷却与熔断。
- 在任务中心可以扫描待办、执行当前可自动处理的项目和领取符合条件的奖励。国际版成长路径的未达标项目会显示客户端操作指引。
- 签到资格和奖励以平台返回为准。Qoder 的设备自动模式使用本机设备。

## 快速开始

### 运行发行版

1. 从 [GitHub Releases](https://github.com/xu20060420fan-cell/workbuddy2api-trae-qoder-minimax-panel/releases/latest) 下载与系统匹配的压缩包并解压。
2. 在解压目录启动程序。首次运行可以生成 `config.json` 和网关密钥，密钥见启动日志。使用示例配置时，请先替换 `api_key` 的示例值。
3. 打开 [本机管理面板](http://127.0.0.1:7863/panel/)，进入对应平台的账号页，按提示登录或添加个人账号。

**Windows PowerShell**

```powershell
.\wb2api.exe -config config.json
```

**Linux / macOS**

```bash
./wb2api -config config.json
```

### Docker Compose

```bash
git clone https://github.com/xu20060420fan-cell/workbuddy2api-trae-qoder-minimax-panel.git
cd workbuddy2api-trae-qoder-minimax-panel
mkdir -p auths data
cp config.example.json config.json
```

编辑 `config.json`，替换示例 `api_key`。Compose 默认使用 UID/GID `10001`。Linux 用户可用下面的命令以自己的 UID/GID 启动，确保挂载的配置和数据目录可写。

```bash
PUID=$(id -u) PGID=$(id -g) docker compose up -d --build
```

启动后访问 [管理面板](http://127.0.0.1:7863/panel/)。用 `docker compose logs -f` 查看启动日志。

### 从源码运行

构建需要 Go 1.26 或更高版本，具体要求见 [go.mod](go.mod)。

```bash
go run ./cmd/server -config config.json
go test ./...
```

## 调用 API

客户端的 Base URL 填 `http://127.0.0.1:7863/v1`，API Key 填 `config.json` 中的 `api_key`。先请求模型列表，再从返回的 `id` 中选择模型。模型名前缀用来区分平台。

```bash
curl http://127.0.0.1:7863/v1/models \
  -H "Authorization: Bearer YOUR_GATEWAY_API_KEY"
```

| 接口 | 用途 |
| --- | --- |
| `GET /v1/models` | 列出可用模型 |
| `POST /v1/chat/completions` | Chat Completions 请求 |
| `POST /v1/responses` | Responses 请求 |
| `GET /status` | 查看账号池状态 |
| `GET /healthz` | 健康检查 |

### 接入 Gemini

把官方密钥配置保存在与 `state_file` 同一目录，默认文件名为 `data/gemini_api_keys.json`。下面的占位值替换为你自己的 Gemini API Key，再重启网关。

```json
{
  "enabled": true,
  "groups": [
    {
      "id": "gemini",
      "name": "Gemini",
      "keys": ["YOUR_GEMINI_API_KEY"]
    }
  ]
}
```

模型会出现在 `/v1/models` 中。分组的 `id` 是模型名前缀。需要为该通道指定出站代理时，可在文件中配置 `proxy_url`。

## 凭据与使用范围

- 账号凭据、会话、密钥和运行数据保存在本机。提交代码或反馈问题前，检查 `config.json`、`auths/`、`data/`、日志和截图，避免公开凭据。
- 服务仅供自己使用时，直接运行可把 `listen` 设为 `127.0.0.1:7863`。使用 Docker 时，可将宿主机端口映射限制为 `127.0.0.1:7863:7863`。
- 请仅接入本人有权使用的账号，并遵守各平台服务条款。本项目供个人自托管使用，作者与第三方付费分发或共享账号池服务无关。

## 支持开发

这个项目从最早的一个脚本长到现在，绝大部分维护都是在深夜和周末挤时间做的。如果它帮你省下了时间、少踩了坑，欢迎：

- 点一个 **Star**，让更多人看到这个项目；
- 提 issue / PR，把遇到的问题和抓到的真实接口反馈回来；
- **扫码打赏**——**如果你觉得这个项目对你有帮助，请多多打赏**，你的支持是继续维护下去最直接的动力。

| 微信赞赏 | 支付宝赞赏 |
| :---: | :---: |
| <img src="assets/donate/wechat.png" alt="微信赞赏码" width="220"> | <img src="assets/donate/alipay.png" alt="支付宝赞赏码" width="220"> |

打赏纯属自愿，不构成任何服务承诺；但每一笔支持都会被记住。谢谢！

## 致谢与许可

本项目源自 [Sliverkiss/workbuddy2api](https://github.com/Sliverkiss/workbuddy2api)，由本仓库继续维护。各平台早期独立实现见 [归档目录](archive/README.md)。项目采用 [MIT License](LICENSE)。
