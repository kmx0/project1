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
	endpoint := fmt.Sprintf("%s/%s", AccSysSddr, number)
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
	// печатаем код ответа
	logrus.Info("Статус-код ", response.Status)
	defer response.Body.Close()
	// читаем поток из тела ответа
	// body, err := io.ReadAll(response.Body)
	// if err != nil {
	// 	logrus.Error("Error on Reading body")
	// 	logrus.Error(err)

	// }

	// bodyString := fmt.Sprintf(`  {
	// 	"order": "%s",
	// 	"status": "REGISTERED"
	// 	}`, number)
	// logrus.Info(bodyString)
	// body := ioutil.NopCloser(strings.NewReader(bodyString))
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
				time.Sleep(3 * time.Second)
				switch {
				// kill -SIGHUP XXXX [XXXX - идентификатор процесса для программы]
				case accrual.Status == "INVALID" || accrual.Status == "PROCESSED":
					logrus.Infof("Writing to table orders status %s for number %s", accrual.Status, number)
					err := store.WriteAccrual(accrual)
					logrus.Info(err)
					//write to db invalid or proc
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
			}
		}()
	case http.StatusTooManyRequests:
		logrus.Info("StatusTooManyRequests")
		return
	default:
		logrus.Info("Schet")
		return

	}
	return
}

func GetAccrualCicle(store storage.Storage, accrual *types.AccrualO, AccSysSddr string, number string) {
	client := &http.Client{}
	endpoint := fmt.Sprintf("%s/%s", AccSysSddr, number)
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
	// печатаем код ответа
	logrus.Info("Статус-код ", response.Status)
	defer response.Body.Close()
	// читаем поток из тела ответа
	// body, err := io.ReadAll(response.Body)
	// if err != nil {
	// 	logrus.Error("Error on Reading body")
	// 	logrus.Error(err)

	// }
	// if response.Status = "Accepted"
	// check is json/

	// test1 := fmt.Sprintf(`  {
	// 	"order": "%s",
	// 	"status": "REGISTERED"
	// 	}`, number)
	// test2 := fmt.Sprintf(`  {
	// 	"order": "%s",
	// 	"status": "PROCESSING"
	// }`, number)
	// test3 := fmt.Sprintf(`  {
	// 	"order": "%s",
	// 	"status": "PROCESSED",
	// 	"accrual": 400
	// 	}`, number)
	// test4 := fmt.Sprintf(`  {
	// 	"order": "%s",
	// 	"status": "INVALID"
	// 	}`, number)
	// testBody := []string{test1, test2, test3, test4}
	// rand.Seed(time.Now().UnixNano())
	// min := 0
	// max := 3
	// // fmt.Println()
	// rid := rand.Intn(max-min+1) + min
	// // logrus.Info(body)
	// body := ioutil.NopCloser(strings.NewReader(testBody[rid]))
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
