package replica

// Config это минимальный набор данных для запуска репликации
type Config struct {
	// ReadRDB означает будут ли обрабатываться данные из RDB или их нужно пропустить
	ReadRDB bool

	// CacheRDB означает требуется ли кешировать RDB
	CacheRDB bool

	// Debug включает запись отладки
	Debug bool

	// СacheRDBFile это адрес файла для хранения кеша RDB
	CacheRDBFile string

	// BacklogSize это размер Backlog
	BacklogSize int

	// DebugDumpDir это адрес каталога в который сохраняются дампы
	DebugDumpDir string
}
