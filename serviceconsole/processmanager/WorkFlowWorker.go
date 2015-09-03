package processmanager

import (
	"duov6.com/serviceconsole/messaging"
	"fmt"
	"log"
)

type WorkFlowWorker struct {
}

func (worker WorkFlowWorker) GetWorkerName() string {
	return "WorkFlowWorker"
}

func (worker WorkFlowWorker) ExecuteWorker(request *messaging.ServiceRequest) messaging.ServiceResponse {
	fmt.Println("Not Implemented in WorkFlow Disconnect in RabbitMQ.")
	var temp = messaging.ServiceResponse{}
	if request.Body != nil {
		log.Printf("Received a message: %s", request.Body)
	}
	temp.IsSuccess = false
	return temp
}
