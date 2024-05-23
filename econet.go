package econet

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type Econet interface {
	getRequest(cmd string) (*http.Request, error)
	setParams(paramName int, paramValue int) error
	GetParams() (Params, error)
	SetHUWTemp(temp int) error
	SetBoilerStatus(status BoilerStatus) error
	SetCOTemp(temp int) error
	DisableHUW() error
}

type econet struct {
	username string
	password string
	hostname string
	client   http.Client
	logger   *slog.Logger
}

func (e econet) SetBoilerStatus(status BoilerStatus) error {
	type response struct {
		ParamKey   string `json:"paramKey"`
		ParamValue int    `json:"paramValue"`
		Result     string `json:"result"`
	}
	r := &response{}
	cmd := fmt.Sprintf("newParam?newParamName=BOILER_CONTROL&newParamValue=%d", status)
	req, err := e.getRequest(cmd)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return err
	}
	if r.Result != "OK" {
		return errors.New("unable to set new param: " + r.Result)
	}
	return nil
}

func (e econet) SetHUWTemp(temp int) error {
	if err := e.setParams(HuwTemp, temp); err != nil {
		return err
	}
	return nil
}

func (e econet) ChangeHUWMode(mode int) error {
	if err := e.setParams(HUWHeater, mode); err != nil {
		return err
	}
	return nil
}
func (e econet) SetCOTemp(temp int) error {
	if err := e.setParams(COTemp, temp); err != nil {
		return err
	}
	return nil
}

func (e econet) setParams(paramName int, paramValue int) error {
	type response struct {
		ParamKey   string `json:"paramKey"`
		ParamValue int    `json:"paramValue"`
		Result     string `json:"result"`
	}
	r := response{}
	cmd := fmt.Sprintf("rmCurrNewParam?newParamKey=%d&newParamValue=%d", paramName, paramValue)
	req, err := e.getRequest(cmd)
	if err != nil {
		return err
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return err
	}
	if r.Result != "OK" {
		return errors.New("unable to set new param: " + r.Result)
	}
	return nil
}
func (e econet) getRequest(cmd string) (*http.Request, error) {
	req, err := http.NewRequest("GET", e.hostname+"/econet/"+cmd, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(e.username, e.password)
	return req, nil
}
func (e econet) GetParams() (Params, error) {
	type Response struct {
		Param Params `json:"curr"`
	}
	r := Response{}
	req, err := e.getRequest("regParams")
	resp, err := e.client.Do(req)
	if err != nil {
		return Params{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Params{}, err
	}
	if err := json.Unmarshal(body, &r); err != nil {
		return Params{}, err
	}
	return r.Param, nil
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
	return &econet{username: username, password: password, hostname: hostname, logger: logger, client: client}

}
