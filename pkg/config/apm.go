package config

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/DataDog/datadog-agent/pkg/util/log"
)

func setupAPM(config Config) {
	config.SetKnown("apm_config.enabled")
	config.SetKnown("apm_config.env")
	config.SetKnown("apm_config.additional_endpoints.*")
	config.SetKnown("apm_config.apm_non_local_traffic")
	config.SetKnown("apm_config.max_traces_per_second")
	config.SetKnown("apm_config.max_memory")
	config.SetKnown("apm_config.log_file")
	config.SetKnown("apm_config.apm_dd_url")
	config.SetKnown("apm_config.profiling_dd_url")
	config.SetKnown("apm_config.profiling_additional_endpoints.*")
	config.SetKnown("apm_config.max_cpu_percent")
	config.SetKnown("apm_config.receiver_port")
	config.SetKnown("apm_config.receiver_socket")
	config.SetKnown("apm_config.connection_limit")
	config.SetKnown("apm_config.ignore_resources")
	config.SetKnown("apm_config.replace_tags")
	config.SetKnown("apm_config.obfuscation.elasticsearch.enabled")
	config.SetKnown("apm_config.obfuscation.elasticsearch.keep_values")
	config.SetKnown("apm_config.obfuscation.mongodb.enabled")
	config.SetKnown("apm_config.obfuscation.mongodb.keep_values")
	config.SetKnown("apm_config.obfuscation.http.remove_query_string")
	config.SetKnown("apm_config.obfuscation.http.remove_paths_with_digits")
	config.SetKnown("apm_config.obfuscation.remove_stack_traces")
	config.SetKnown("apm_config.obfuscation.redis.enabled")
	config.SetKnown("apm_config.obfuscation.memcached.enabled")
	config.SetKnown("apm_config.extra_sample_rate")
	config.SetKnown("apm_config.dd_agent_bin")
	config.SetKnown("apm_config.max_events_per_second")
	config.SetKnown("apm_config.trace_writer.connection_limit")
	config.SetKnown("apm_config.trace_writer.queue_size")
	config.SetKnown("apm_config.service_writer.connection_limit")
	config.SetKnown("apm_config.service_writer.queue_size")
	config.SetKnown("apm_config.stats_writer.connection_limit")
	config.SetKnown("apm_config.stats_writer.queue_size")
	config.SetKnown("apm_config.connection_reset_interval") // in seconds
	config.SetKnown("apm_config.analyzed_rate_by_service.*")
	config.SetKnown("apm_config.analyzed_spans.*")
	config.SetKnown("apm_config.log_throttling")
	config.SetKnown("apm_config.bucket_size_seconds")
	config.SetKnown("apm_config.receiver_timeout")
	config.SetKnown("apm_config.watchdog_check_delay")
	config.SetKnown("apm_config.max_payload_size")

	if runtime.GOARCH == "386" && runtime.GOOS == "windows" {
		// on Windows-32 bit, the trace agent isn't installed.  Set the default to disabled
		// so that there aren't messages in the log about failing to start.
		config.BindEnvAndSetDefault("apm_config.enabled", false, "DD_APM_ENABLED")
	} else {
		config.BindEnvAndSetDefault("apm_config.enabled", true, "DD_APM_ENABLED")
	}

	config.BindEnvAndSetDefault("apm_config.connection_limit", 0, "DD_APM_CONNECTION_LIMIT", "DD_CONNECTION_LIMIT")
	config.BindEnvAndSetDefault("apm_config.env", "none", "DD_APM_ENV")
	config.BindEnvAndSetDefault("apm_config.apm_non_local_traffic", false, "DD_APM_NON_LOCAL_TRAFFIC")
	config.BindEnvAndSetDefault("apm_config.apm_dd_url", "https://trace.agent.datadoghq.com", "DD_APM_DD_URL")
	config.BindEnvAndSetDefault("apm_config.connection_reset_interval", 0, "DD_APM_CONNECTION_RESET_INTERVAL")
	config.BindEnvAndSetDefault("apm_config.receiver_port", 8126, "DD_APM_RECEIVER_PORT", "DD_RECEIVER_PORT")
	config.BindEnvAndSetDefault("apm_config.max_events_per_second", 200, "DD_APM_MAX_EPS", "DD_MAX_EPS")
	config.BindEnvAndSetDefault("apm_config.max_traces_per_second", 10, "DD_APM_MAX_TPS", "DD_MAX_TPS")
	config.BindEnvAndSetDefault("apm_config.max_memory", 5e8, "DD_APM_MAX_MEMORY")
	config.BindEnvAndSetDefault("apm_config.max_cpu_percent", 50, "DD_APM_MAX_CPU_PERCENT")
	config.BindEnvAndSetDefault("apm_config.receiver_socket", "", "DD_APM_RECEIVER_SOCKET")
	config.BindEnvAndSetDefault("apm_config.profiling_dd_url", "https://intake.profile.datadoghq.com/v1/input", "DD_APM_PROFILING_DD_URL")

	config.BindEnv("apm_config.profiling_additional_endpoints", "DD_APM_PROFILING_ADDITIONAL_ENDPOINTS")
	config.BindEnv("apm_config.additional_endpoints", "DD_APM_ADDITIONAL_ENDPOINTS")
	config.BindEnv("apm_config.replace_tags", "DD_APM_REPLACE_TAGS")
	config.BindEnv("apm_config.analyzed_spans", "DD_APM_ANALYZED_SPANS")

	config.BindEnv("apm_config.ignore_resources", "DD_APM_IGNORE_RESOURCES", "DD_IGNORE_RESOURCE")

	config.SetEnvKeyTransformer("apm_config.ignore_resources", func(in string) interface{} {
		r, err := splitCSVString(in, ',')
		if err != nil {
			log.Warnf(`"apm_config.ignore_resources" can not be parsed: %v`, err)
			return []string{}
		}
		return r
	})

	config.SetEnvKeyTransformer("apm_config.replace_tags", func(in string) interface{} {
		var out []map[string]string
		if err := json.Unmarshal([]byte(in), &out); err != nil {
			log.Warnf(`"apm_config.replace_tags" can not be parsed: %v`, err)
		}
		return out
	})

	config.SetEnvKeyTransformer("apm_config.analyzed_spans", func(in string) interface{} {
		out, err := parseAnalyzedSpans(in)
		if err != nil {
			log.Errorf(`Bad format for "apm_config.analyzed_spans" it should be of the form \"service_name|operation_name=rate,other_service|other_operation=rate\", error: %v`, err)
		}
		return out
	})
}

func splitCSVString(s string, sep rune) ([]string, error) {
	r := csv.NewReader(strings.NewReader(s))
	r.TrimLeadingSpace = true
	r.LazyQuotes = true
	r.Comma = sep

	return r.Read()
}

func parseNameAndRate(token string) (string, float64, error) {
	parts := strings.Split(token, "=")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("Bad format")
	}
	rate, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return "", 0, fmt.Errorf("Unabled to parse rate")
	}
	return parts[0], rate, nil
}

// parseAnalyzedSpans parses the env string to extract a map of spans to be analyzed by service and operation.
// the format is: service_name|operation_name=rate,other_service|other_operation=rate
func parseAnalyzedSpans(env string) (analyzedSpans map[string]interface{}, err error) {
	analyzedSpans = make(map[string]interface{})
	if env == "" {
		return
	}
	tokens := strings.Split(env, ",")
	for _, token := range tokens {
		name, rate, err := parseNameAndRate(token)
		if err != nil {
			return nil, err
		}
		analyzedSpans[name] = rate
	}
	return
}
