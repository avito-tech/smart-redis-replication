package status

const (
	// Connect означает что началось подключение к серверу
	Connect Status = "connect"

	// Reconnect означает что началось переподключение к серверу
	Reconnect Status = "reconnect"

	// Disconnect означает что произошло отключение от сервера
	Disconnect Status = "disconnect"

	// StartSync означает что инициировалась репликация
	StartSync Status = "start_sync"

	// RDB означает что началась секция с RDB
	RDB Status = "rdb"

	// StartCacheRDB означает что началось кеширование BulkString с RDB
	StartCacheRDB Status = "start_cache_rdb"

	// StopCacheRDB означает что закончилось кеширование BulkString с RDB
	StopCacheRDB Status = "stop_cache_rdb"

	// StartBacklog означает что началась запись в Backlog
	StartBacklog Status = "start_cache_real_time"

	// SkipReadRDB означает что файл rdb был скачан, но не разобран, перейдя сразу к чтению Backlog
	SkipReadRDB Status = "skip_read_rdb"

	// StartReadRDB означает что началось чтение из BulkString с RDB
	StartReadRDB Status = "start_read_rdb"

	// StopReadRDB означает что закончилось чтение из BulkString с RDB
	StopReadRDB Status = "stop_read_rdb"

	// StartReadBacklog означает что началось чтение из Backlog
	StartReadBacklog Status = "start_read_backlog"

	// StopDecoder означает что декодирование прекратилось
	StopDecoder Status = "stop_decoder"

	// StopReplication означает что репликация прекратилась
	StopReplication Status = "stop_replication"
)

// Status это статус репликации
type Status string
