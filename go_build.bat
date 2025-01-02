@echo off
chcp 65001

REM 检查 syso 文件是否存在
if not exist "rsrc_windows_386.syso" if not exist "rsrc_windows_amd64.syso" (
    echo 未找到 syso 文件，正在生成...
    go-winres make
)

echo 正在构建开发版本...
go build -o ./tools/直播伴侣tool_dev.exe -ldflags "-X 'main.Debug=true'"

echo 正在构建发布版本...
go build -o ./tools/直播伴侣tool.exe -ldflags "-s -w -H=windowsgui"

echo 构建完成！
echo tiktok_tool_dev
echo tiktok_tool
pause