package plugin

import (
	"context"
	"encoding/json"
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

const timeSeriesType = "time series"
const tableType = "table"

// SlsDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
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
		return
	}
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *SlsDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	d.log.Info("QueryData called", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

func (d *SlsDatasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	d.log.Info("query params: ", query.JSON)
	response := backend.DataResponse{}

	// Unmarshal the JSON into our queryModel.
	var payload models.QueryPayload

	response.Error = json.Unmarshal(query.JSON, &payload)
	if response.Error != nil {
		return response
	}

	payload.From = query.TimeRange.From.UnixMilli() / 1000
	payload.To = query.TimeRange.To.UnixMilli() / 1000
	payload.MaxDataPoints = query.MaxDataPoints
	d.log.Info("query", "payload.Query", payload.Query, "format", payload.Format, "from", payload.From, "to", payload.To, "MaxDataPoints", payload.MaxDataPoints)

	logsResp, err := d.Client.GetLogs(d.Settings.Project, d.Settings.LogStore, "", payload.From, payload.To, payload.Query, payload.MaxDataPoints, 0, true)
	if err != nil {
		d.log.Error("GetLogs ", "query", payload.Query, "error ", err)
		return backend.DataResponse{
			Error: err,
		}
	}
	d.log.Info("query GetLogs ", "resp", query.QueryType)
	d.log.Info("query GetLogs ", "QueryType", query.QueryType)

	isTimeSeries := query.QueryType == timeSeriesType
	// create data frame response.
	frame := data.NewFrame(query.RefID)
	if isTimeSeries {
		for _, logRecord := range logsResp.Logs {
			// add fields.
			frame.Fields = append(frame.Fields,
				data.NewField("time", nil, logRecord["time"]),
				data.NewField("value", nil, []string{logRecord["value"]}),
			)
		}
	} else {
		// TODO 待补全
	}

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *SlsDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	d.log.Info("CheckHealth called", "request", req)

	_, err := d.Client.GetLogStore(d.Settings.Project, d.Settings.LogStore)
	if err != nil {
		d.log.Info("CheckHealth failed", "error", err)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "GetLogStore error",
		}, nil
	}

	d.log.Info("CheckHealth success")
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}
