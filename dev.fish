#!/usr/bin/env fish

# 设置环境变量以抑制警告
set -x JSC_SIGNAL_FOR_GC 10
set -x GTK_IM_MODULE ibus

echo "🚀 启动 Wails 开发服务器..."
echo ""
echo "提示："
echo "  - WebView 客户端"
echo "  - (在浏览器中开发: http://localhost:34115)"
echo ""

wails dev
