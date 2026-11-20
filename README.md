# OpenClaw KMS 凭据插件 (kmscli)

[![Build Release](https://github.com/aliyun/alibabacloud-kms-cli/actions/workflows/build.yml/badge.svg)](https://github.com/aliyun/alibabacloud-kms-cli/actions/workflows/build.yml)
[![Release](https://img.shields.io/github/v/release/aliyun/alibabacloud-kms-cli)](https://github.com/aliyun/alibabacloud-kms-cli/releases/latest)

将 OpenClaw 配置中的明文 API Key 替换为阿里云 KMS 托管，实现配置中无明文 API Key，ECS 上不存在明文的 API Key 和 AKSK。

## 下载

- **[最新 Release 版本](https://github.com/aliyun/alibabacloud-kms-cli/releases/latest)** - 推荐，包含各平台预编译二进制文件
- **[Actions 构建产物](https://github.com/aliyun/alibabacloud-kms-cli/actions)** - 每次提交自动构建

---

## 示例环境

本文档所有示例均基于以下部署环境：

- 运行用户：`admin`
- OpenClaw 安装目录：`/home/admin/.openclaw`
- 实际部署时请根据你的用户名和安装路径替换对应路径

---

## 快速开始

### 1. 前置准备

- 已安装 OpenClaw 并初始化配置

### 2. 下载并上传 kmscli

下载对应平台的 kmscli 并上传到 ECS 服务器的 `~/.openclaw/` 目录：

```bash
# 本地执行上传
scp kmscli-linux-amd64 user@ecs-ip:~/.openclaw/kmscli
```

### 3. 创建 KMS 凭据

在 [KMS 控制台](https://yundun.console.aliyun.com/?p=kms#/overview/cn-hangzhou/) 创建凭据：
- 凭据名称：`bailian-api-key`
- 凭据值：你的 API Key 值（如百炼的 SK）
- 记录凭据 ARN 和密钥 ARN

**示例：**
- 凭据名称：`bailian-api-key`
- 凭据值：`sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`

### 4. 创建 RAM 角色并绑定到 ECS

创建自定义策略（最小权限）：

> **参数说明**：
> - `<region>`：KMS 所在地域（如 `cn-hangzhou`）
> - `<uid>`：阿里云账号 UID
> - `<secret-name>`：KMS 凭据名称（如 `bailian-api-key`）
> - `<key-id>`：加密该凭据的 KMS 密钥 ID，可在 KMS 控制台查看

```json
{
  "Version": "1",
  "Statement": [
    {
      "Action": "kms:GetSecretValue",
      "Resource": "acs:kms:<region>:<uid>:secret/<secret-name>",
      "Effect": "Allow"
    },
    {
      "Action": "kms:Decrypt",
      "Resource": "acs:kms:<region>:<uid>:key/<key-id>",
      "Effect": "Allow"
    }
  ]
}
```

创建 RAM 角色（信任服务选择 ECS），绑定到 ECS 实例。

### 5. 部署到 OpenClaw

按以下步骤手动配置：

#### 步骤 1：复制并设置权限

```bash
# 复制文件
cp kmscli-linux-amd64 ~/.openclaw/kmscli

# 设置权限（Linux/macOS）
sudo chown $(id -u):$(id -g) ~/.openclaw/kmscli
chmod +x ~/.openclaw/kmscli
```

#### 步骤 2：备份配置文件

```bash
cp ~/.openclaw/openclaw.json ~/.openclaw/openclaw.json.bak
```

#### 步骤 3：编辑 openclaw.json

**重要提示**：编辑时保持 JSON 节点原有顺序，不要删除其他内容！

**添加 secrets 配置**（在根节点下）：

```json
{
  "secrets": {
    "providers": {
      "kms_bailian": {
        "source": "exec",
        "command": "/home/admin/.openclaw/kmscli",
        "args": ["openclaw", "getsecret"],
        "jsonOnly": true
      }
    }
  },
  "models": { ... }
}
```

**修改 apiKey 配置**（在 models.providers 下）：

```json
{
  "models": {
    "providers": {
      "dashscope": {
        "apiKey": {
          "source": "exec",
          "provider": "kms_bailian",
          "id": "<secret-name>"
        }
      }
    }
  }
}
```

#### 步骤 4：验证配置

```bash
# 审计配置
openclaw secrets audit --check

# 重新加载
openclaw secrets reload
```

---

## 工作原理

1. **配置阶段**：将 `openclaw.json` 中的明文 `apiKey` 替换为 KMS 凭据引用
2. **运行阶段**：OpenClaw 调用 kmscli 从 KMS 动态获取凭据值
3. **安全优势**：配置文件中无敏感信息，ECS 上无明文 API Key

---

## 配置示例

### 方式一：默认凭证链（ECS RAM Role）

ECS 实例已绑定 RAM 角色，通过阿里云默认凭证链自动获取 AK/SK：

```json
{
  "secrets": {
    "providers": {
      "kms_bailian": {
        "source": "exec",
        "command": "/home/admin/.openclaw/kmscli",
        "args": ["openclaw", "getsecret"],
        "jsonOnly": true
      }
    }
  },
  "models": {
    "providers": {
      "dashscope": {
        "apiKey": {
          "source": "exec",
          "provider": "kms_bailian",
          "id": "bailian-api-key"
        }
      }
    }
  }
}
```

---

### 方式二：云原生认证（mauth）

云原生环境（ECS 实例身份 / ACK OIDC / ClientKey）推荐使用配置文件。

在 `~/.config/kmscli/config.json` 创建配置文件，kmscli 启动时自动读取：

**配置文件路径优先级**：
1. `KMSCLI_CONFIG_PATH` 环境变量指定的路径
2. 默认路径 `~/.config/kmscli/config.json`
3. 兼容旧路径 `~/.openclaw/kmscli.json`
4. 当前工作目录下的 `kmscli.json` 或 `.openclaw/kmscli.json`（兜底）

**配置优先级**（高到低）：命令行 flag > 环境变量 > 配置文件 > 默认值

**mauth ECS 实例身份**：

```json
{
  "aapArn": "acs:kms:cn-hangzhou:123456789:applicationaccesspoint/your-aap-id"
}
```

**mauth ACK OIDC**（需挂载 ServiceAccount Token）：

```json
{
  "aapArn": "acs:kms:cn-hangzhou:123456789:applicationaccesspoint/your-aap-id"
}
```

**mauth ClientKey**：

```json
{
  "clientKeyConfigPath": "/path/to/client_key.json",
  "clientKeyPasswordPath": "/path/to/password.txt",
  "regionId": "cn-hangzhou"
}
```

**配置字段说明**：

| 字段 | 说明 | 必填 | 示例值 |
|---|---|---|---|
| `authMethod` | 认证方式：`ECSInstanceIdentity` / `ACKOidcJwt` / `ClientKey` | 否，程序根据其他字段自动推断 | `ECSInstanceIdentity` |
| `aapArn` | AAP ARN | ECS/ACK 场景必填 | `acs:kms:cn-hangzhou:123456789:applicationaccesspoint/your-aap-id` |
| `tokenPath` | ACK OIDC Token 路径（默认 `/var/run/secrets/kubernetes.io/serviceaccount/token`） | 否 | `/var/run/secrets/tokens/token` |
| `clientKeyConfigPath` | ClientKey 配置文件路径 | ClientKey 场景必填 | `/path/to/client_key.json` |
| `clientKeyPasswordPath` | ClientKey 密码文件路径 | ClientKey 场景必填 | `/path/to/password.txt` |
| `kmsEndpoint` | KMS 端点 | 否 | `kms-vpc.cn-hangzhou.aliyuncs.com` |
| `regionId` | 地域 ID，用于自动拼接 endpoint | ACK 可选；ClientKey 必填 | `cn-hangzhou` |
| `endpointType` | `Public` 或 `Vpc`（默认 `Vpc`） | 否 | `Vpc` |

若不方便使用配置文件，也可通过环境变量配置：

**ECS 实例身份**：

```bash
export KMSCLI_AAP_ARN=acs:kms:cn-hangzhou:123456789:applicationaccesspoint/your-aap-id
```

**ACK OIDC**：

```bash
export KMSCLI_AAP_ARN=acs:kms:cn-hangzhou:123456789:applicationaccesspoint/your-aap-id
```

**ClientKey**：

```bash
export KMSCLI_CLIENT_KEY_CONFIG_PATH=/path/to/client_key.json
export KMSCLI_CLIENT_KEY_PASSWORD_PATH=/path/to/password.txt
export REGION_ID=cn-hangzhou
```
---

**进阶：Flag 方式**（不推荐，除非环境变量无法使用）

如果环境变量不便配置，也可通过 `args` 直接传参：

| Flag | 短别名 | 说明 |
|---|---|---|
| `--aap-arn` | `-a` | AAP ARN |
| `--token-path` | `-t` | ACK OIDC Token 路径 |
| `--client-key-config-path` | `-k` | ClientKey 配置文件路径 |
| `--client-key-password-path` | `-p` | ClientKey 密码文件路径 |
| `--kms-endpoint` | `-e` | KMS 端点 |
| `--auth-method` | `-A` | 认证方式 |

ECS 实例身份：
```json
"args": ["openclaw", "getsecret", "-a", "acs:kms:cn-hangzhou:123456789:applicationaccesspoint/your-aap-id"]
```

ACK OIDC：
```json
"args": ["openclaw", "getsecret", "-a", "acs:kms:cn-hangzhou:123456789:applicationaccesspoint/your-aap-id"]
```

ClientKey：
```json
"args": ["openclaw", "getsecret", "-k", "/path/to/client_key.json", "-p", "/path/to/password.txt", "-e", "kms-vpc.cn-hangzhou.aliyuncs.com"]
```

---

**环境变量说明**

| 环境变量 | 说明 | 必填 |
|---|---|---|
| `KMSCLI_CONFIG_PATH` | 配置文件路径，覆盖默认路径 | 否 |
| `KMSCLI_AAP_ARN` | AAP ARN | ECS/ACK 场景必填 |
| `KMSCLI_TOKEN_PATH` | ACK OIDC 的 JWT token 路径（默认 `/var/run/secrets/kubernetes.io/serviceaccount/token`） | 否 |
| `KMSCLI_CLIENT_KEY_CONFIG_PATH` | ClientKey 配置文件路径 | ClientKey 场景必填 |
| `KMSCLI_CLIENT_KEY_PASSWORD_PATH` | ClientKey 密码文件路径 | ClientKey 场景必填 |
| `KMSCLI_KMS_ENDPOINT` | KMS 端点 | 否 |
| `REGION_ID` | 地域 ID，用于自动拼接 endpoint | ACK 可选；ClientKey 必填 |
| `ALIBABA_CLOUD_REGION_ID` | 地域 ID（与 `REGION_ID` 等价） | ACK 可选；ClientKey 必填 |
| `ENDPOINT_TYPE` | Endpoint 类型：`Public` 或 `Vpc`（默认 `Vpc`） | 否 |
| `KMSCLI_AUTH_METHOD` | 认证方式：`ECSInstanceIdentity` / `ACKOidcJwt` / `ClientKey` | 否，通常自动推断 |

**自动推断规则**

未指定 `--auth-method` 时，程序按以下优先级自动推断：
- 存在 `KMSCLI_CLIENT_KEY_CONFIG_PATH` → `ClientKey`
- 存在 `KMSCLI_TOKEN_PATH` → `ACKOidcJwt`
- 存在默认 ACK ServiceAccount Token 文件且配置了 `KMSCLI_AAP_ARN` → `ACKOidcJwt`
- 存在 `KMSCLI_AAP_ARN` → `ECSInstanceIdentity`

---

## 命令行工具

kmscli 提供以下命令：

```bash
# 普通模式（明文输出，可直接执行）
kmscli secret getsecret <secretName>

# 普通模式 + 自动解密（Secret 值为 KMS 加密密文时适用）
# 注意：getsecret-decrypt 仅支持默认凭证链（RAM Role / AK / STS），mauth 模式不支持
kmscli secret getsecret-decrypt <secretName>

# OpenClaw 模式（JSON 格式输出，供 OpenClaw 调用，依赖 OpenClaw 配置）
# 此命令需要 OpenClaw 通过 stdin 传入 JSON 请求，不可直接手动执行
kmscli openclaw getsecret

# OpenClaw 模式 + 自动解密
# 注意：getsecret-decrypt 仅支持默认凭证链（RAM Role / AK / STS），mauth 模式不支持
kmscli openclaw getsecret-decrypt
```

---

## 注意事项

- ECS 与 KMS 建议同地域，使用 VPC 内网访问
- 默认使用VPC网络访问 KMS，可通过环境变量指定公网访问：`export ENDPOINT_TYPE=Public`。当ECS 与 KMS 不在同一个地域时只能使用公网。
- 修改配置前请备份 `openclaw.json` 为 `.bak`
- 如需恢复：`cp ~/.openclaw/openclaw.json.bak ~/.openclaw/openclaw.json`

---

## 常见问题

### 1. 权限不足

**现象**：`chmod: changing permissions of 'kmscli': Operation not permitted`

**解决**：
```bash
sudo chown $(id -u):$(id -g) ~/.openclaw/kmscli
sudo chmod +x ~/.openclaw/kmscli
```

### 2. 配置未生效

**现象**：端口未监听或进程未启动

**解决**：
- 等待几秒后再次检查
- 查看日志排查错误：
  ```bash
  # 查看 openclaw 日志
  ls -la /tmp/openclaw/
  cat /tmp/openclaw/openclaw-*.log
  ```
- 确认 JSON 配置格式正确

### 3. ECS 未绑定 RAM 角色

**现象**：日志中出现以下错误之一：
- `NoCredentialProviders` / `ECS metadata access failed`
- `get role name failed: GET http://100.100.100.200/latest/meta-data/ram/security-credentials/ 404`
- `unable to get credentials from any of the providers in the chain`

**解决**：
- 在 ECS 控制台为实例绑定已创建 RAM 角色
- 确认 RAM 角色已附加 KMS 访问策略
- 绑定后等待 1-2 分钟使权限生效

### 4. JSON 节点顺序问题

**现象**：配置更新后 openclaw 无法启动

**解决**：
- 使用备份文件恢复：`cp openclaw.json.bak openclaw.json`
- 重新配置时注意保持节点原有顺序
- 不要删除或重建整个 provider 节点

---

## 验证命令

```bash
CONFIG_FILE="${OPENCLAW_DIR:-$HOME/.openclaw}/openclaw.json"

# 验证 secrets 配置
jq '.secrets.providers.kms_bailian' "$CONFIG_FILE"

# 验证 apiKey 配置
jq '.models.providers.dashscope.apiKey' "$CONFIG_FILE"

# 检查服务进程
ps aux | grep openclaw-gateway
```
