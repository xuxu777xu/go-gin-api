#!/bin/bash

# 业务模块生成器脚本
# 用法: ./newbiz.sh YourBizName
# 例如: ./newbiz.sh Order
# 注意: 首次使用可能需要添加执行权限: chmod +x newbiz.sh

# --- 配置 ---
MODULE_PATH="tongcheng" # 从 go.mod 获取的模块路径
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
TEMPLATE_DIR="$SCRIPT_DIR/templates"
TARGET_BASE_DIR="$SCRIPT_DIR/../internal" # 指向 internal 目录
HANDLER_TPL="handler.go.tpl"
SERVICE_TPL="service.go.tpl"
REPO_TPL="repo.go.tpl"

# --- 参数检查 ---
if [ -z "$1" ]; then
  echo "错误: 请提供业务域名作为参数。"
  echo "用法: $0 YourBizName"
  exit 1
fi

# --- 变量处理 ---
BizName="$1"
# 将首字母转换为小写
bizName="$(echo "${BizName:0:1}" | tr '[:upper:]' '[:lower:]')${BizName:1}"

echo "正在为业务 '$BizName' (小写: '$bizName') 生成代码..."

# --- 目标路径定义 ---
HANDLER_TARGET_DIR="$TARGET_BASE_DIR/handler"
SERVICE_TARGET_DIR="$TARGET_BASE_DIR/service"
REPO_TARGET_DIR="$TARGET_BASE_DIR/repo"

HANDLER_TARGET_FILE="$HANDLER_TARGET_DIR/${bizName}.go"
SERVICE_TARGET_FILE="$SERVICE_TARGET_DIR/${bizName}.go"
REPO_TARGET_FILE="$REPO_TARGET_DIR/${bizName}.go"

# --- 文件存在性检查 ---
if [ -f "$HANDLER_TARGET_FILE" ]; then
  echo "错误: Handler 文件已存在: $HANDLER_TARGET_FILE"
  exit 1
fi
if [ -f "$SERVICE_TARGET_FILE" ]; then
  echo "错误: Service 文件已存在: $SERVICE_TARGET_FILE"
  exit 1
fi
if [ -f "$REPO_TARGET_FILE" ]; then
  echo "错误: Repo 文件已存在: $REPO_TARGET_FILE"
  exit 1
fi

# --- 创建目录 ---
mkdir -p "$HANDLER_TARGET_DIR"
mkdir -p "$SERVICE_TARGET_DIR"
mkdir -p "$REPO_TARGET_DIR"

# --- 复制和替换模板 ---
echo "创建 Handler 文件: $HANDLER_TARGET_FILE"
sed "s/{{BizName}}/$BizName/g; s/{{bizName}}/$bizName/g; s|{{modulePath}}|$MODULE_PATH|g" "$TEMPLATE_DIR/$HANDLER_TPL" > "$HANDLER_TARGET_FILE"

echo "创建 Service 文件: $SERVICE_TARGET_FILE"
sed "s/{{BizName}}/$BizName/g; s/{{bizName}}/$bizName/g; s|{{modulePath}}|$MODULE_PATH|g" "$TEMPLATE_DIR/$SERVICE_TPL" > "$SERVICE_TARGET_FILE"

echo "创建 Repo 文件: $REPO_TARGET_FILE"
sed "s/{{BizName}}/$BizName/g; s/{{bizName}}/$bizName/g; s|{{modulePath}}|$MODULE_PATH|g" "$TEMPLATE_DIR/$REPO_TPL" > "$REPO_TARGET_FILE"

echo "成功为业务 '$BizName' 生成以下文件:"
echo "- $HANDLER_TARGET_FILE"
echo "- $SERVICE_TARGET_FILE"
echo "- $REPO_TARGET_FILE"
echo "请记得检查并填充这些文件中的具体逻辑。"
echo "您可能需要运行 'go mod tidy' 来更新依赖。"

exit 0