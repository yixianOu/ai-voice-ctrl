# 快速配置脚本 (PowerShell)
# 用于设置 OpenAI API 和代理

Write-Host "=== AI Voice Control 配置向导 ===" -ForegroundColor Cyan
Write-Host ""

# 1. 配置 API Key
Write-Host "步骤 1: 配置 OpenAI API Key" -ForegroundColor Yellow
$apiKey = Read-Host "请输入您的 OpenAI API Key (sk-proj-...)"
if ($apiKey -ne "") {
    $env:OPENAI_API_KEY = $apiKey
    Write-Host "✅ API Key 已设置" -ForegroundColor Green
} else {
    Write-Host "❌ API Key 未设置" -ForegroundColor Red
    exit 1
}

Write-Host ""

# 2. 配置代理（可选）
Write-Host "步骤 2: 配置代理 (可选)" -ForegroundColor Yellow
Write-Host "如果您在国内或需要通过代理访问 OpenAI，请配置代理"
Write-Host "常见代理端口: Clash=7890, V2Ray=10809, SSR=1080"
Write-Host ""

$useProxy = Read-Host "是否需要配置代理? (y/n)"
if ($useProxy -eq "y" -or $useProxy -eq "Y") {
    $proxyHost = Read-Host "代理地址 (默认: 127.0.0.1)"
    if ($proxyHost -eq "") {
        $proxyHost = "127.0.0.1"
    }
    
    $proxyPort = Read-Host "代理端口 (默认: 7890)"
    if ($proxyPort -eq "") {
        $proxyPort = "7890"
    }
    
    $proxyType = Read-Host "代理类型 (http/socks5, 默认: http)"
    if ($proxyType -eq "") {
        $proxyType = "http"
    }
    
    $proxyURL = "${proxyType}://${proxyHost}:${proxyPort}"
    $env:HTTP_PROXY = $proxyURL
    $env:HTTPS_PROXY = $proxyURL
    
    Write-Host "✅ 代理已设置: $proxyURL" -ForegroundColor Green
    
    # 测试代理
    Write-Host ""
    Write-Host "测试代理连接..." -ForegroundColor Yellow
    try {
        $response = Invoke-WebRequest -Uri "https://api.openai.com/v1/models" -Proxy $proxyURL -TimeoutSec 5 -ErrorAction Stop
        Write-Host "✅ 代理连接成功！" -ForegroundColor Green
    } catch {
        Write-Host "⚠️  代理连接测试失败，但您仍可以继续尝试" -ForegroundColor Yellow
        Write-Host "   请确保代理软件正在运行" -ForegroundColor Yellow
    }
} else {
    Write-Host "⏭️  跳过代理配置" -ForegroundColor Gray
}

Write-Host ""
Write-Host "=== 配置完成 ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "当前配置:" -ForegroundColor White
Write-Host "  OPENAI_API_KEY: " -NoNewline
Write-Host "已设置" -ForegroundColor Green
if ($env:HTTP_PROXY) {
    Write-Host "  HTTP_PROXY: " -NoNewline
    Write-Host $env:HTTP_PROXY -ForegroundColor Green
}
Write-Host ""

Write-Host "现在可以运行示例了:" -ForegroundColor Yellow
Write-Host "  cd examples" -ForegroundColor Cyan
Write-Host "  go run ." -ForegroundColor Cyan
Write-Host ""

# 询问是否立即运行
$runNow = Read-Host "是否立即运行示例? (y/n)"
if ($runNow -eq "y" -or $runNow -eq "Y") {
    Write-Host ""
    Write-Host "正在运行示例..." -ForegroundColor Cyan
    Set-Location -Path "examples"
    go run .
}
