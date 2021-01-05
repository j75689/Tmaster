package config

import (
	"time"
)

type Config struct {
	LogLevel       string               `mapstructure:"log_level"`
	LogFormat      string               `mapstructure:"log_format"`
	HTTP           HttpConfig           `mapstructure:"http"`
	MQ             MQConfig             `mapstructure:"mq"`
	DB             DBConfig             `mapstructure:"db"`
	Redis          RedisConfig          `mapstructure:"redis"`
	JobEndpoint    JobEndpointConfig    `mapstructure:"job_endpoint"`
	JobInitializer JobInitializerConfig `mapstructure:"job_initializer"`
	TaskScheduler  TaskSchedulerConfig  `mapstructure:"task_scheduler"`
	DBHelper       DBHelperConfig       `mapstructure:"db_helper"`
	TaskWorker     TaskWorkerConfig     `mapstructure:"task_worker"`
	OpenTracing    OpenTracingConfig    `mapstructure:"open_tracing"`
}

type HttpConfig struct {
	Port    uint          `mapstructure:"port"`
	Graphql GraphqlConfig `mapstructure:"graphql"`
}

type GraphqlConfig struct {
	Endpoint   string                  `mapstructure:"endpoint"`
	Playground GraphqlPlaygroundConfig `mapstructure:"playground"`
}

type GraphqlPlaygroundConfig struct {
	Path    string `mapstructure:"path"`
	Title   string `mapstructure:"title"`
	Disable bool   `mapstructure:"disable"`
}

type JobEndpointConfig struct {
	InitJob PushQueue `mapstructure:"init_job"`
}

type JobInitializerConfig struct {
	MaxTaskExecution int       `mapstructure:"max_task_execution"`
	InitJob          PullQueue `mapstructure:"init_job"`
	TaskInput        PushQueue `mapstructure:"task_input"`
}

type TaskSchedulerConfig struct {
	TaskInput    PushQueue `mapstructure:"task_input"`
	TaskOutput   PullQueue `mapstructure:"task_output"`
	TaskDBHelper PushQueue `mapstructure:"task_db_helper"`
	JobDBHelper  PushQueue `mapstructure:"job_db_helper"`
}

type DBHelperConfig struct {
	Job  PullQueue `mapstructure:"job"`
	Task PullQueue `mapstructure:"task"`
}

type TaskWorkerConfig struct {
	InitJob    PushQueue `mapstructure:"init_job"`
	TaskInput  PullQueue `mapstructure:"task_input"`
	TaskOutput PushQueue `mapstructure:"task_output"`
}

type MQConfig struct {
	Driver       string          `mapstructure:"driver"`
	StopTimeout  time.Duration   `mapstructure:"stop_timeout"`
	Distribution int             `mapstructure:"distribution"`
	GooglePubSub GooglePubSubArg `mapstructure:"google_pub_sub"`
	Nats         NatsArg         `mapstructure:"nats"`
	RedisStream  RedisStreamArg  `mapstructure:"redis_stream"`
}

type GooglePubSubArg struct {
	CredentialPath         string `mapstructure:"credential_path"`
	Synchronous            bool   `mapstructure:"synchronous"`
	MaxOutstandingMessages int    `mapstructure:"max_outstanding_messages"`
	MaxOutstandingBytes    int    `mapstructure:"max_outstanding_bytes"`
	NumGoroutines          int    `mapstructure:"num_goroutines"`
}

type NatsArg struct {
	Servers              []string      `mapstructure:"servers"`
	MaxReconnect         int           `mapstructure:"max_reconnect"`
	ReconnectWait        time.Duration `mapstructure:"reconnect_wait"`
	RetryOnFailedConnect bool          `mapstructure:"retry_on_failed_connect"`
	ClusterName          string        `mapstructure:"cluster_name"`
	DurableName          string        `mapstructure:"durable_name"`
	QueueGroup           string        `mapstructure:"queue_group"`
	MaxInflight          int           `mapstructure:"max_inflight"`
	MaxPubAcksInflight   int           `mapstructure:"max_pub_acks_inflight"`
	WorkerSize           int           `mapstructure:"worker_size"`
	AckWait              time.Duration `mapstructure:"ack_wait"`
	PingInterval         int           `mapstructure:"ping_interval"`
	PingMaxOut           int           `mapstructure:"ping_max_out"`
	User                 string        `mapstructure:"user"`
	Password             string        `mapstructure:"password"`
	Token                string        `mapstructure:"token"`
}

type RedisStreamArg struct {
	RedisOption    RedisConfig   `mapstructure:"redis_option"`
	NumConsumers   int           `mapstructure:"num_consumers"`
	PrefetchLimit  int64         `mapstructure:"prefetch_limit"`
	PollDuration   time.Duration `mapstructure:"poll_duration"`
	ProcessTimeout time.Duration `mapstructure:"process_timeout"`
}

type PullQueue struct {
	ProjectID   string `mapstructure:"project_id"`
	Topic       string `mapstructure:"topic"`
	SubscribeID string `mapstructure:"subscribe_id"`
}

type PushQueue struct {
	ProjectID string `mapstructure:"project_id"`
	Topic     string `mapstructure:"topic"`
}

type DBConfig struct {
	Driver         string        `mapstructure:"driver"`
	Host           string        `mapstructure:"host"`
	Port           uint          `mapstructure:"port"`
	DBName         string        `mapstructure:"dbname"`
	InstanceName   string        `mapstructure:"instance_name"`
	User           string        `mapstructure:"user"`
	Password       string        `mapstructure:"password"`
	ConnectTimeout string        `mapstructure:"connect_timeout"`
	ReadTimeout    string        `mapstructure:"read_timeout"`
	WriteTimeout   string        `mapstructure:"write_timeout"`
	DialTimeout    time.Duration `mapstructure:"dial_timeout"`
	MaxLifetime    time.Duration `mapstructure:"max_lifetime"`
	MaxIdleTime    time.Duration `mapstructure:"max_idletime"`
	MaxIdleConn    int           `mapstructure:"max_idle_conn"`
	MaxOpenConn    int           `mapstructure:"max_open_conn"`
	SSLMode        bool          `mapstructure:"ssl_mode"`
	LogLevel       int           `mapstructure:"log_level"`
}

type RedisConfig struct {
	Host          string        `mapstructure:"host"`
	Port          uint          `mapstructure:"port"`
	DB            int           `mapstructure:"db"`
	Password      string        `mapstructure:"password"`
	MinIdleConns  int           `mapstructure:"min_idle_conns"`
	PoolSize      int           `mapstructure:"pool_size"`
	LockTimeout   time.Duration `mapstructure:"lock_timeout"`
	LockFlushTime time.Duration `mapstructure:"lock_flush_time"`
	DialTimeout   time.Duration `mapstructure:"dial_timeout"`
}

type OpenTracingConfig struct {
	Enable         bool                   `mapstructure:"enable"`
	Driver         string                 `mapstructure:"driver"`
	ServiceName    string                 `mapstructure:"service_name"`
	RemoteReporter string                 `mapstructure:"remote_reporter"`
	LocalReporter  string                 `mapstructure:"local_reporter"`
	CustomTag      map[string]interface{} `mapstructure:"custom_tag"`
}
