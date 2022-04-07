package gcpstatus

import (
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func GetGcpStatus(gcpStatusEndpoint string) (GCPServiceStatuses GCPSserviceStatus, err error) {
	//url := "https://status.cloud.google.com/incidents.json"
	statusClient := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodGet, gcpStatusEndpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, getErr := statusClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	GCPServiceStatuses = GCPSserviceStatus{}
	jsonErr := json.Unmarshal(body, &GCPServiceStatuses)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	//fmt.Println("%+v", GcpServiceStatuses)
	//for _, val := range GcpServiceStatuses {
	//	fmt.Println(val.ID)
	//}

	return GCPServiceStatuses, nil

}
