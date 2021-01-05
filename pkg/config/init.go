package config

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gorm.io/gorm/logger"
)

func NewConfig(configPath string) (config Config, err error) {
	var file *os.File
	file, _ = os.Open(configPath)

	v := viper.New()
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	/* default */
	v.SetDefault("log_level", "INFO")
	v.SetDefault("log_format", "console")
	v.SetDefault("http.port", "8080")
	v.SetDefault("http.graphql.endpoint", "/api/v1/graphql")
	v.SetDefault("http.graphql.playground.path", "/graphql/playground")
	v.SetDefault("http.graphql.playground.title", "GraphQL playground")
	v.SetDefault("http.graphql.playground.disable", false)
	v.SetDefault("mq.driver", "google_pub_sub")
	v.SetDefault("mq.stop_timeout", "30s")
	v.SetDefault("mq.distribution", 10)

	// pubsub args
	v.SetDefault("mq.google_pub_sub.synchronous", true)
	v.SetDefault("mq.google_pub_sub.max_outstanding_messages", 10)
	v.SetDefault("mq.google_pub_sub.max_outstanding_bytes", 1e10)
	v.SetDefault("mq.google_pub_sub.num_goroutines", 100)

	// nats args
	v.SetDefault("mq.nats.servers", []string{})
	v.SetDefault("mq.nats.max_reconnect", 5)
	v.SetDefault("mq.nats.reconnect_wait", 5*time.Second)
	v.SetDefault("mq.nats.retry_on_failed_connect", true)
	v.SetDefault("mq.nats.cluster_name", "test-cluster")
	v.SetDefault("mq.nats.durable_name", "durable")
	v.SetDefault("mq.nats.queue_group", "Tmaster")
	v.SetDefault("mq.nats.max_inflight", 1024)
	v.SetDefault("mq.nats.max_pub_acks_inflight", 512)
	v.SetDefault("mq.nats.worker_size", 100)
	v.SetDefault("mq.nats.ack_wait", 10*time.Second)
	v.SetDefault("mq.nats.ping_interval", 10)
	v.SetDefault("mq.nats.ping_max_out", 3)
	v.SetDefault("mq.nats.user", "")
	v.SetDefault("mq.nats.password", "")
	v.SetDefault("mq.nats.token", "")

	// redis_stream args
	v.SetDefault("mq.redis_stream.redis_option.host", "localhost")
	v.SetDefault("mq.redis_stream.redis_option.port", 6379)
	v.SetDefault("mq.redis_stream.redis_option.db", 0)
	v.SetDefault("mq.redis_stream.redis_option.pool_size", 10)
	v.SetDefault("mq.redis_stream.redis_option.min_idle_conns", 5)
	v.SetDefault("mq.redis_stream.redis_option.dial_timeout", "1s")
	v.SetDefault("mq.redis_stream.num_consumers", 100)
	v.SetDefault("mq.redis_stream.prefetch_limit", 1000)
	v.SetDefault("mq.redis_stream.process_timeout", 10*time.Second)

	// db
	v.SetDefault("db.driver", "mysql")
	v.SetDefault("db.host", "localhost")
	v.SetDefault("db.port", 3306)
	v.SetDefault("db.dbname", "Tmaster")
	v.SetDefault("db.instance_name", "")
	v.SetDefault("db.user", "root")
	v.SetDefault("db.password", "")
	v.SetDefault("db.connect_timeout", "10s")
	v.SetDefault("db.read_timeout", "30s")
	v.SetDefault("db.write_timeout", "30s")
	v.SetDefault("db.dial_timeout", "10s")
	v.SetDefault("db.max_idletime", "1h")
	v.SetDefault("db.max_lifetime", "1h")
	v.SetDefault("db.max_idle_conn", 2)
	v.SetDefault("db.max_open_conn", 5)
	v.SetDefault("db.log_level", logger.Info)
	v.SetDefault("db.ssl_mode", false)

	// redis
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 5)
	v.SetDefault("redis.lock_timeout", "30s")
	v.SetDefault("redis.lock_flush_time", "1s")
	v.SetDefault("redis.dial_timeout", "1s")

	// opentracing
	v.SetDefault("open_tracing.enable", false)
	v.SetDefault("open_tracing.driver", "")
	v.SetDefault("open_tracing.service_name", "")
	v.SetDefault("open_tracing.remote_reporter", "")
	v.SetDefault("open_tracing.local_reporter", "")
	v.SetDefault("open_tracing.custom_tag", make(map[string]interface{}, 0))

	// endpoint
	v.SetDefault("job_endpoint.init_job.project_id", "tmaster")
	v.SetDefault("job_endpoint.init_job.topic", "tmaster_init_job")

	// initializer worker
	v.SetDefault("job_initializer.init_job.project_id", "tmaster")
	v.SetDefault("job_initializer.init_job.topic", "tmaster_init_job")
	v.SetDefault("job_initializer.init_job.subscribe_id", "tmaster_init_job")
	v.SetDefault("job_initializer.task_input.project_id", "tmaster")
	v.SetDefault("job_initializer.task_input.topic", "tmaster_task_input")
	v.SetDefault("job_initializer.max_task_execution", 500)

	// scheduler worker
	v.SetDefault("task_scheduler.task_output.project_id", "tmaster")
	v.SetDefault("task_scheduler.task_output.topic", "tmaster_task_output")
	v.SetDefault("task_scheduler.task_output.subscribe_id", "tmaster_task_output")
	v.SetDefault("task_scheduler.task_input.project_id", "tmaster")
	v.SetDefault("task_scheduler.task_input.topic", "tmaster_task_input")
	v.SetDefault("task_scheduler.job_db_helper.project_id", "tmaster")
	v.SetDefault("task_scheduler.job_db_helper.topic", "tmaster_job_db_helper")
	v.SetDefault("task_scheduler.task_db_helper.project_id", "tmaster")
	v.SetDefault("task_scheduler.task_db_helper.topic", "tmaster_task_db_helper")

	// db helper worker
	v.SetDefault("db_helper.job.project_id", "tmaster")
	v.SetDefault("db_helper.job.topic", "tmaster_job_db_helper")
	v.SetDefault("db_helper.job.subscribe_id", "tmaster_job_db_helper")
	v.SetDefault("db_helper.task.project_id", "tmaster")
	v.SetDefault("db_helper.task.topic", "tmaster_task_db_helper")
	v.SetDefault("db_helper.task.subscribe_id", "tmaster_task_db_helper")

	// task worker
	v.SetDefault("task_worker.init_job.project_id", "tmaster")
	v.SetDefault("task_worker.init_job.topic", "tmaster_init_job")
	v.SetDefault("task_worker.task_input.project_id", "tmaster")
	v.SetDefault("task_worker.task_input.topic", "tmaster_task_input")
	v.SetDefault("task_worker.task_input.subscribe_id", "tmaster_task_input")
	v.SetDefault("task_worker.task_output.project_id", "tmaster")
	v.SetDefault("task_worker.task_output.topic", "tmaster_task_output")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.ReadConfig(file)

	if err = v.Unmarshal(&config); err != nil {
		return
	}

	return
}
