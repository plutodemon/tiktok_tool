package llog

// LogEntry 表示一条要发送到服务器的日志
type LogEntry struct {
	Time    int64  `json:"time"` // 使用 Unix 时间戳代替 time.Time
	Level   string `json:"level"`
	Message string `json:"message"`
	Host    string `json:"host"`
}

// LogSetting 日志配置结构
type LogSetting struct {
	Console      bool   `json:"console" toml:"console"`             // 是否输出到控制台
	File         bool   `json:"file" toml:"file"`                   // 是否输出到文件
	FilePath     string `json:"file_path" toml:"file_path"`         // 日志文件路径
	MaxSize      int    `json:"max_size" toml:"max_size"`           // 每个日志文件的最大大小（MB）
	MaxAge       int    `json:"max_age" toml:"max_age"`             // 保留旧日志文件的最大天数
	MaxBackups   int    `json:"max_backups" toml:"max_backups"`     // 保留的旧日志文件的最大数量
	Compress     bool   `json:"compress" toml:"compress"`           // 是否压缩旧日志文件
	LocalTime    bool   `json:"local_time" toml:"local_time"`       // 是否使用本地时间
	Format       string `json:"format" toml:"format"`               // 日志文件名格式
	Level        string `json:"level" toml:"level"`                 // 日志级别
	OutputFormat string `json:"output_format" toml:"output_format"` // 日志输出格式 ("json" 或 "text")
}
