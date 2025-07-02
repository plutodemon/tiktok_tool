go-winres make

go build -o ./tools/tiktok_tool_dev.exe -ldflags "-X 'tiktok_tool/config.Debug=true'"

go build -o ./tools/tiktok_tool.exe -ldflags "-s -w -H=windowsgui"

pause