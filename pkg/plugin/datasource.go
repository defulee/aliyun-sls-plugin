package plugin

import (
	"context"
	"encoding/json"
	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
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

type DatasourceInstance struct {
	Client   *sls.Client
	Settings *models.PluginSettings
}

func newDataSourceInstance(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	log.DefaultLogger.Info("newDataSourceInstance called", "Settings", settings)
	pluginSettings, _ := models.LoadPluginSettings(settings)
	return DatasourceInstance{
		Client: &sls.Client{
			AccessKeyID:     pluginSettings.AccessKeyId,
			AccessKeySecret: pluginSettings.Secrets.AccessKeySecret,
			Endpoint:        pluginSettings.Endpoint,
		},
		Settings: pluginSettings,
	}, nil
}

// SlsDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type SlsDatasource struct {
	InstanceManager instancemgmt.InstanceManager
}

// NewSlsDatasource creates a new datasource instance.
func NewSlsDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	instanceManager := datasource.NewInstanceManager(newDataSourceInstance)
	return &SlsDatasource{
		InstanceManager: instanceManager,
	}, nil
}

func (d *SlsDatasource) getInstance(ctx backend.PluginContext) (*DatasourceInstance, error) {
	instance, err := d.InstanceManager.Get(ctx)
	if err != nil {
		return nil, err
	}
	return instance.(*DatasourceInstance), nil
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource Settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSlsDatasource factory function.
func (d *SlsDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *SlsDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

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

type queryModel struct {
	From       int64  `json:"from"`
	To         int64  `json:"to"`
	Query      string `json:"queryExp"`
	MaxLineNum int64  `json:"maxLineNum"`
	Offset     int64  `json:"offset"`
	Reverse    bool   `json:"reverse"`
}

func (d *SlsDatasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}

	// Unmarshal the JSON into our queryModel.
	var qm queryModel

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	instance, _ := d.getInstance(pCtx)

	from := query.TimeRange.From.UnixMilli() / 1000
	to := query.TimeRange.To.UnixMilli() / 1000

	logsResp, err := instance.Client.GetLogs(instance.Settings.Project, instance.Settings.LogStore, "", from, to, qm.Query, 500, 0, false)
	if err != nil {
		return backend.DataResponse{}
	}
	if err != nil {
		log.DefaultLogger.Error("GetLogs ", "query : ", qm.Query, "error ", err)
		return backend.DataResponse{
			Error: err,
		}
	}
	log.DefaultLogger.Info("GetLogs ", "QueryType : ", query.QueryType)

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
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	instance, _ := d.getInstance(req.PluginContext)

	_, err := instance.Client.GetLogStore(instance.Settings.Project, instance.Settings.LogStore)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "GetLogStore error",
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}
