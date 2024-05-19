package econet

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type Econet interface {
	GetParams() (*Params, error)
}

type econet struct {
	hostname string
	client   http.Client
	logger   *slog.Logger
}

func (e econet) GetParams() (*Params, error) {
	type Response struct {
		Curr struct {
			params Params
		} `json:"curr"`
	}
	r := Response{}
	resp, err := e.client.Get(e.hostname + "/econet/regParams")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	return &r.Curr.params, nil
}

type Params struct {
	PumpCOWorks   bool    `json:"pumpCOWorks"`
	BoilerPower   int     `json:"boilerPower"`
	BoilerPowerKW float32 `json:"boilerPowerKW"`
	TempCOSet     float32 `json:"tempCOSet"`
	TempCO        float32 `json:"tempCO"`
	TempCWUSet    float32 `json:"tempCWUSet"`
	TempCWU       float32 `json:"tempCWU"`
	TempFeeder    float32 `json:"tempFeeder"`
	FanWorks      bool    `json:"fanWorks"`
	FuelStream    float32 `json:"fuelStream"`
	FuelLevel     int     `json:"fuelLevel"`
	OperationMode int     `json:"mode"`
}

type SysParams struct {
	Uid          string `json:"uid"`
	ControllerID string `json:"controllerId"`
}

func NewEconet(hostname, username, password string, logger *slog.Logger) Econet {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	if !strings.HasPrefix(hostname, "http://") || !strings.HasPrefix(hostname, "https://") {
		hostname = "http://" + hostname
	}
	client := http.Client{}
	req, err := http.NewRequest("GET", hostname, nil)
	if err != nil {
		logger.Error("Error creating new request: ", err)
	}
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error executing request: ", err)
	}
	defer resp.Body.Close()
	return &econet{hostname: hostname, logger: logger, client: client}

}
