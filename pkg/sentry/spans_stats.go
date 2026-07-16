package sentry

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type SentrySpansStats = map[string]interface{}

var allowedSpansStatsIntervals = []int64{15, 30, 60, 120, 300, 600, 900, 1800, 3600, 7200, 10800, 14400, 21600, 43200, 86400}

func (gei *GetSpansStatsInput) snapInterval() bool {
	min := time.Duration(allowedSpansStatsIntervals[0]) * time.Second
	max := time.Duration(allowedSpansStatsIntervals[len(allowedSpansStatsIntervals)-1]) * time.Second

	if gei.Interval < min || gei.Interval > max {
		return false
	}
	nearest := min
	for _, seconds := range allowedSpansStatsIntervals {
		valid := time.Duration(seconds) * time.Second
		if (gei.Interval - valid).Abs() < (gei.Interval - nearest).Abs() {
			nearest = valid
		}
	}
	if nearest > gei.To.Sub(gei.From) {
		return false
	}
	gei.Interval = nearest
	return true
}

type GetSpansStatsInput struct {
	OrganizationSlug string
	ProjectIds       []string
	Environments     []string
	Fields           []string
	YAxis            []string
	Query            string
	From             time.Time
	To               time.Time
	Sort             string
	Interval         time.Duration
	Limit            int64
}

func (gei *GetSpansStatsInput) ToQuery() string {
	urlPath := fmt.Sprintf("/api/0/organizations/%s/events-stats/?", gei.OrganizationSlug)
	if gei.Limit < 1 || gei.Limit > 10 {
		gei.Limit = 10
	}
	params := url.Values{}
	params.Set("dataset", "spans")
	params.Set("query", gei.Query)
	params.Set("start", gei.From.Format("2006-01-02T15:04:05Z07:00"))
	params.Set("end", gei.To.Format("2006-01-02T15:04:05Z07:00"))
	if gei.snapInterval() {
		params.Set("interval", FormatSentryInterval(gei.Interval))
	}
	params.Set("partial", "1")
	params.Set("excludeOther", "0")
	params.Set("sampling", "NORMAL")
	if gei.Sort != "" {
		params.Set("sort", gei.Sort)
	}
	params.Set("topEvents", strconv.FormatInt(gei.Limit, 10))
	for _, field := range gei.Fields {
		params.Add("field", field)
	}
	for _, field := range gei.YAxis {
		params.Add("yAxis", field)
	}
	for _, projectId := range gei.ProjectIds {
		params.Add("project", projectId)
	}
	for _, environment := range gei.Environments {
		params.Add("environment", environment)
	}
	return urlPath + params.Encode()
}

func (sc *SentryClient) GetSpansStats(gei GetSpansStatsInput) (SentrySpansStats, string, error) {
	var out SentrySpansStats
	executedQueryString := gei.ToQuery()
	err := sc.Fetch(executedQueryString, &out)
	return out, sc.BaseURL + executedQueryString, err
}
