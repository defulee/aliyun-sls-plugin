package plugin

import (
	"context"
	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/models"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Make sure SlsDatasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler, backend.StreamHandler interfaces. Plugin should not
// implement all these interfaces - only those which are required for a particular task.
// For example if plugin does not need streaming functionality then you are free to remove
// methods that implement backend.StreamHandler. Implementing instancemgmt.InstanceDisposer
// is useful to clean up resources used by previous datasource instance when a new datasource
// instance created upon datasource Settings changed.
var (
	_ backend.QueryDataHandler      = (*SlsDatasource)(nil)
	_ backend.CheckHealthHandler    = (*SlsDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*SlsDatasource)(nil)
)

const timeSeriesType = "TimeSeries"
const tableType = "Table"

// SlsDatasource is a datasource which can respond to data queries, reports its health.
type SlsDatasource struct {
	Client   *sls.Client
	Settings *models.PluginSettings
	log      log.Logger
}

// NewSlsDatasource creates a new datasource instance.
func NewSlsDatasource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	log.DefaultLogger.Info("NewSlsDatasource called")
	pluginSettings, _ := models.LoadPluginSettings(settings)
	log.DefaultLogger.Info("NewSlsDatasource pluginSettings",
		"AccessKeyId", pluginSettings.AccessKeyId,
		"AccessKeySecret", pluginSettings.Secrets.AccessKeySecret,
		"Endpoint", pluginSettings.Endpoint,
		"Project", pluginSettings.Project,
		"LogStore", pluginSettings.LogStore)

	return &SlsDatasource{
		Client: &sls.Client{
			AccessKeyID:     pluginSettings.AccessKeyId,
			AccessKeySecret: pluginSettings.Secrets.AccessKeySecret,
			Endpoint:        pluginSettings.Endpoint,
		},
		Settings: pluginSettings,
		log:      log.DefaultLogger,
	}, nil
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource Settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSlsDatasource factory function.
func (d *SlsDatasource) Dispose() {
	// Clean up datasource instance resources.
	err := d.Client.Close()
	if err != nil {
		d.log.Warn("SlsDatasource#Dispose close client error", err)
		return
	}
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *SlsDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	d.log.Info("SlsDatasource#QueryData called", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

func (d *SlsDatasource) query(_ context.Context, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}

	payload, err := models.ParsePayload(query)
	if err != nil || payload.Hide {
		return response
	}

	logsResp, err := d.Client.GetLogs(d.Settings.Project, d.Settings.LogStore, "", payload.From, payload.To, payload.Query, payload.MaxDataPoints, 0, true)
	if err != nil {
		d.log.Error("SlsDatasource#GetLogs ", "query", payload.Query, "error ", err)
		return backend.DataResponse{
			Error: err,
		}
	}
	d.log.Info("SlsDatasource#query GetLogs ", "logsCount", len(logsResp.Logs))

	// create data frame response.
	frame := data.NewFrame(query.RefID)
	switch payload.Format {
	case timeSeriesType, tableType:
		d.formatData(payload, logsResp, frame, payload.Format == timeSeriesType)
	default:
		d.log.Error("SlsDatasource#query not support format")
	}

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

func (d *SlsDatasource) formatData(payload *models.QueryPayload, logsResp *sls.GetLogsResponse, frame *data.Frame, formatTime bool) {
	// 设置时区
	loc, _ := time.LoadLocation(payload.Timezone)

	var dataRecords []models.DataRecord
	for _, logRecord := range logsResp.Logs {
		var parseErr error
		var val float64
		var timeVal time.Time
		otherFieldVal := make(map[string]float64)
		for k, v := range logRecord {
			if len(v) > 0 {
				if formatTime && k == payload.TimeField {
					// 时间(格式如："2018-07-11 15:07:51") to 时间戳
					timeVal, parseErr = time.ParseInLocation(payload.TimeFormat, v, loc)
					// timeVal, parseErr = dateparse.ParseIn(v, loc)
					if parseErr != nil {
						d.log.Error("SlsDatasource#formatData time is illegal", k, v)
					}
				} else if strings.Index(k, "__") != 0 {
					val, parseErr = strconv.ParseFloat(v, 64)
					if parseErr != nil {
						d.log.Error("SlsDatasource#formatData val is not float64", "key", k, "value", v)
					} else {
						otherFieldVal[k] = val
					}
				}
			}
		}

		if parseErr == nil {
			dataRecords = append(dataRecords, models.DataRecord{Time: timeVal, Values: otherFieldVal})
		}
	}

	// sort record by time field
	sort.Slice(dataRecords, func(i, j int) bool {
		return dataRecords[i].Time.Before(dataRecords[j].Time)
	})
	var timeArr []time.Time
	fieldValArrMap := make(map[string][]float64)
	for _, record := range dataRecords {
		timeArr = append(timeArr, record.Time)
		for field, val := range record.Values {
			if valArr, ok := fieldValArrMap[field]; ok {
				valArr = append(valArr, val)
				fieldValArrMap[field] = valArr
			} else {
				fieldValArrMap[field] = []float64{val}
			}
		}
	}

	// add fields.
	frame.Fields = append(frame.Fields, data.NewField("time", nil, timeArr))
	for field, valArr := range fieldValArrMap {
		frame.Fields = append(frame.Fields, data.NewField(field, nil, valArr))
	}

	frame.SetMeta(&data.FrameMeta{})
	frame.Meta.Type = data.FrameTypeTimeSeriesWide
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *SlsDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	d.log.Info("SlsDatasource#CheckHealth called", "request", req)

	_, err := d.Client.GetLogStore(d.Settings.Project, d.Settings.LogStore)
	if err != nil {
		d.log.Info("SlsDatasource#CheckHealth failed", "error", err)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "GetLogStore error",
		}, nil
	}

	d.log.Info("SlsDatasource#CheckHealth success")
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}
