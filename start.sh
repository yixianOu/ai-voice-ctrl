#!/bin/bash

# AI 语音控制助手 - 启动脚本

echo "🚀 启动 AI 语音控制助手..."
echo ""

# 检查依赖
echo "📦 检查系统依赖..."

check_command() {
    if ! command -v $1 &> /dev/null; then
        echo "❌ $1 未安装"
        echo "   安装命令: $2"
        return 1
    else
        echo "✅ $1 已安装"
        return 0
    fi
}

# 检查必需工具
check_command "pactl" "sudo apt install pulseaudio-utils"
check_command "playerctl" "sudo apt install playerctl"
check_command "xdg-open" "sudo apt install xdg-utils"

echo ""
echo "🎤 准备启动应用..."
echo ""
echo "📌 使用提示:"
echo "  1. 点击麦克风按钮开始录音"
echo "  2. 允许浏览器访问麦克风"
echo "  3. 说出命令，例如："
echo "     - 打开 Firefox"
echo "     - 播放音乐"
echo "     - 音量设为 50%"
echo "     - 创建文件 test.txt"
echo ""

# 启动 Wails 开发服务器
wails dev
