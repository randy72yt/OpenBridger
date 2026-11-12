package model

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PerfMetricAux stores aggregated relay samples that must not pollute the
// public perf_metrics numbers: probe traffic, internal tokens (including the
// playground) and excluded client errors. Row shape mirrors PerfMetric plus a
// source dimension; counters accumulate through the same upsert pattern.
type PerfMetricAux struct {
	Id             int    `json:"id" gorm:"primaryKey"`
	ModelName      string `json:"model_name" gorm:"size:128;uniqueIndex:idx_perf_aux_key,priority:1"`
	Group          string `json:"group" gorm:"column:group;size:64;uniqueIndex:idx_perf_aux_key,priority:2"`
	Source         string `json:"source" gorm:"size:16;uniqueIndex:idx_perf_aux_key,priority:3"` // probe|internal|client_error
	BucketTs       int64  `json:"bucket_ts" gorm:"uniqueIndex:idx_perf_aux_key,priority:4;index:idx_perf_aux_bucket"`
	RequestCount   int64  `json:"-" gorm:"default:0"`
	SuccessCount   int64  `json:"-" gorm:"default:0"`
	TotalLatencyMs int64  `json:"-" gorm:"default:0"`
	TtftSumMs      int64  `json:"-" gorm:"default:0"`
	TtftCount      int64  `json:"-" gorm:"default:0"`
	OutputTokens   int64  `json:"-" gorm:"default:0"`
	GenerationMs   int64  `json:"-" gorm:"default:0"`
}

func (PerfMetricAux) TableName() string {
	return "perf_metrics_aux"
}

func UpsertPerfMetricAux(metric *PerfMetricAux) error {
	if metric == nil || metric.RequestCount == 0 {
		return nil
	}
	return DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "model_name"},
			{Name: "group"},
			{Name: "source"},
			{Name: "bucket_ts"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"request_count":    gorm.Expr("perf_metrics_aux.request_count + ?", metric.RequestCount),
			"success_count":    gorm.Expr("perf_metrics_aux.success_count + ?", metric.SuccessCount),
			"total_latency_ms": gorm.Expr("perf_metrics_aux.total_latency_ms + ?", metric.TotalLatencyMs),
			"ttft_sum_ms":      gorm.Expr("perf_metrics_aux.ttft_sum_ms + ?", metric.TtftSumMs),
			"ttft_count":       gorm.Expr("perf_metrics_aux.ttft_count + ?", metric.TtftCount),
			"output_tokens":    gorm.Expr("perf_metrics_aux.output_tokens + ?", metric.OutputTokens),
			"generation_ms":    gorm.Expr("perf_metrics_aux.generation_ms + ?", metric.GenerationMs),
		}),
	}).Create(metric).Error
}

func GetPerfMetricsAux(modelName string, group string, source string, startTs int64, endTs int64) ([]PerfMetricAux, error) {
	var metrics []PerfMetricAux
	query := DB.Model(&PerfMetricAux{}).
		Where("model_name = ? AND bucket_ts >= ? AND bucket_ts <= ?", modelName, startTs, endTs)
	if group != "" {
		query = query.Where(commonGroupCol+" = ?", group)
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}
	err := query.Order("bucket_ts ASC").Find(&metrics).Error
	return metrics, err
}

// GetPerfMetricsAuxBucketsAll returns raw per-bucket aux rows for every source;
// callers merge the source dimension themselves.
func GetPerfMetricsAuxBucketsAll(startTs int64, endTs int64, groups []string) ([]PerfMetricAux, error) {
	var metrics []PerfMetricAux
	query := DB.Model(&PerfMetricAux{}).
		Where("bucket_ts >= ? AND bucket_ts <= ?", startTs, endTs)
	if groups != nil {
		if len(groups) == 0 {
			return metrics, nil
		}
		query = query.Where(commonGroupCol+" IN ?", groups)
	}
	err := query.Order("bucket_ts ASC").Find(&metrics).Error
	return metrics, err
}

func DeletePerfMetricsAuxBefore(cutoffTs int64) error {
	if cutoffTs <= 0 {
		return nil
	}
	return DB.Where("bucket_ts < ?", cutoffTs).Delete(&PerfMetricAux{}).Error
}
