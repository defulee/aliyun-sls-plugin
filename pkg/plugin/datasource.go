package plugin

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/models"
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
		d.formatData(payload, logsResp, frame, len(payload.TimeFormat) > 0)
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

	// 时间(格式如："2018-07-11 15:07:51") to 时间戳
	dataRecords, fieldArr := d.parseDateRecord(logsResp, formatTime, payload, loc)

	if formatTime {
		// sort record by time field
		sort.Slice(dataRecords, func(i, j int) bool {
			return dataRecords[i].Time.Before(dataRecords[j].Time)
		})

		var timeArr []time.Time
		for _, record := range dataRecords {
			timeArr = append(timeArr, record.Time)
		}

		if len(timeArr) > 0 {
			d.log.Info("SlsDatasource#formatData add time field to frame")
			// add fields.
			frame.Fields = append(frame.Fields, data.NewField("time", nil, timeArr))
		}
	}

	if payload.Format == timeSeriesType {
		// 解析数据字段
		numberFieldMap := d.parseNumberFields(fieldArr, dataRecords)

		// 处理数据字段
		metricFieldMap, numberRecords := d.processNumberField(dataRecords, numberFieldMap)

		d.log.Info("SlsDatasource#formatData convert row to col")
		for metricField, _ := range metricFieldMap {
			d.log.Info("SlsDatasource#formatData", "numberField", metricField)
			var numberArr = make([]*float64, len(numberRecords))
			for _, record := range numberRecords {
				d.log.Info("SlsDatasource#formatData numberField", "record", record)
				var number *float64 = nil
				for field, val := range record {
					if field == metricField {
						number = val
					}
				}
				numberArr = append(numberArr, number)
			}
			frame.Fields = append(frame.Fields, data.NewField(metricField, nil, numberArr))
		}
	} else {
		for _, strfield := range fieldArr {
			var valArr []string
			for _, record := range dataRecords {
				var value = ""
				for field, val := range record.FieldValDict {
					if field == strfield {
						value = val
					}
				}
				valArr = append(valArr, value)
			}

			frame.Fields = append(frame.Fields, data.NewField(strfield, nil, valArr))
		}
	}
}

// 从日志记录解析数据记录
func (d *SlsDatasource) parseDateRecord(logsResp *sls.GetLogsResponse, formatTime bool, payload *models.QueryPayload, loc *time.Location) ([]models.DataRecord, []string) {
	var dataRecords []models.DataRecord
	fieldDict := make(map[string]string)
	for _, logRecord := range logsResp.Logs {
		d.log.Info("SlsDatasource#formatData logRecord:", logRecord)
		var parseErr error
		var timeVal time.Time
		fieldStringValDict := make(map[string]string)
		for k, v := range logRecord {
			d.log.Info("SlsDatasource#formatData", "k", k, "v", v)
			if len(v) > 0 {
				if formatTime && k == payload.TimeField {
					// 时间(格式如："2018-07-11 15:07:51") to 时间戳
					timeVal, parseErr = time.ParseInLocation(payload.TimeFormat, v, loc)
					if parseErr != nil {
						d.log.Error("SlsDatasource#formatData time is illegal", k, v)
					}
				} else if strings.Index(k, "__") != 0 {
					fieldStringValDict[k] = v
				}
			}
		}

		if parseErr == nil {
			for field, _ := range fieldStringValDict {
				fieldDict[field] = "1"
			}
			dataRecords = append(dataRecords, models.DataRecord{Time: timeVal, FieldValDict: fieldStringValDict})
		}
	}

	var fieldArr []string
	for field, _ := range fieldDict {
		fieldArr = append(fieldArr, field)
	}

	return dataRecords, fieldArr
}

// 添加时间字段
func (d *SlsDatasource) AddTimeFieldToFrame(dataRecords *[]models.DataRecord, frame *data.Frame) {
	d.log.Info("SlsDatasource#formatData sort records")

}

func (d *SlsDatasource) processNumberField(dataRecords []models.DataRecord, numberFieldMap map[string]string) (map[string]string, []map[string]*float64) {
	d.log.Info("SlsDatasource#processNumberField")
	fieldDict := make(map[string]string)
	numberRecords := make([]map[string]*float64, len(dataRecords))
	for _, record := range dataRecords {
		var metricPrefix string = ""
		for field, value := range record.FieldValDict {
			if _, ok := numberFieldMap[field]; !ok {
				metricPrefix = metricPrefix + value + "#"
			}
		}
		d.log.Info("SlsDatasource#processNumberField", "metricPrefix", metricPrefix)

		fieldNumberValDict := make(map[string]*float64)
		for field, val := range record.FieldValDict {
			if _, ok := numberFieldMap[field]; ok {
				metric := metricPrefix + field
				d.log.Info("SlsDatasource#processNumberField", "metric", metric)
				number, parseErr := strconv.ParseFloat(val, 64)
				d.log.Info("SlsDatasource#processNumberField ParseFloat", "val", val, "number", number)
				if parseErr != nil {
					d.log.Info("SlsDatasource#processNumberField number is nil")
					fieldNumberValDict[metric] = nil
				} else {
					d.log.Info("SlsDatasource#processNumberField", "number", number)
					fieldNumberValDict[metric] = &number
				}
				fieldDict[metric] = "1"
				numberRecords = append(numberRecords, fieldNumberValDict)
			}
		}
	}
	return fieldDict, numberRecords
}

func (d *SlsDatasource) parseNumberFields(fieldArr []string, dataRecords []models.DataRecord) map[string]string {
	d.log.Info("SlsDatasource#parseNumberFields")
	var numberFieldMap = make(map[string]string)
	for _, field := range fieldArr {
		var isNumber = true
		for _, record := range dataRecords {
			if val, ok := record.FieldValDict[field]; ok {
				_, parseErr := strconv.ParseFloat(val, 64)
				if parseErr != nil {
					isNumber = false
				}
			}
		}
		if isNumber {
			numberFieldMap[field] = "1"
		}
	}
	return numberFieldMap
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
