package accrual

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kmx0/project1/cmd/gophermart/storage"
	"github.com/kmx0/project1/internal/types"
	"github.com/sirupsen/logrus"
)

func GetAccrual(store storage.Storage, AccSysSddr string, number string) (err error) {
	client := &http.Client{}
	endpoint := fmt.Sprintf("%s/api/orders/%s", AccSysSddr, number)
	logrus.Info(endpoint)
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	errors.Is(nil, err)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	response, err := client.Do(request)
	if err != nil {
		logrus.Error("Error on requesting")
		logrus.Error(err)

	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK:
		decoder := json.NewDecoder(response.Body)
		var accrual types.AccrualO

		err = decoder.Decode(&accrual)
		if err != nil {
			logrus.Error(err)
			return err
		}
		logrus.Infof("%v", accrual)
		go func() {
			for {
				switch {
				case accrual.Status == "INVALID" || accrual.Status == "PROCESSED":
					logrus.Infof("Writing to table orders status %s for number %s", accrual.Status, number)
					err := store.WriteAccrual(accrual)
					logrus.Info(err)
					return
				case accrual.Status == "REGISTERED":
					logrus.Infof("Getted %s", accrual.Status)
					logrus.Infof("Getting new status for number  %s", number)
					GetAccrualCicle(store, &accrual, AccSysSddr, number)
				case accrual.Status == "PROCESSING":
					logrus.Infof("Writing to table balance status %s for number %s", accrual.Status, number)
					err := store.WriteAccrual(accrual)
					GetAccrualCicle(store, &accrual, AccSysSddr, number)
					logrus.Info(err)
				default:
					return
				}
				time.Sleep(3 * time.Second)
			}
		}()
	case http.StatusTooManyRequests:
		logrus.Info("StatusTooManyRequests")
		return
	default:
		logrus.Info("Default")
		return

	}
	return
}

func GetAccrualCicle(store storage.Storage, accrual *types.AccrualO, AccSysSddr string, number string) {
	client := &http.Client{}
	endpoint := fmt.Sprintf("%s/api/orders/%s", AccSysSddr, number)
	logrus.Info(AccSysSddr)
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	errors.Is(nil, err)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	response, err := client.Do(request)
	if err != nil {
		logrus.Error("Error on requesting")
		logrus.Error(err)

	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusOK:
		decoder := json.NewDecoder(response.Body)

		err = decoder.Decode(&accrual)
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.Info(accrual)
	case http.StatusTooManyRequests:
		time.Sleep(time.Second * 60)
	default:
		return
	}

}
