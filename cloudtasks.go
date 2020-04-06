package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sinmetal/gcpmetadata"
	"github.com/vvakame/sdlog/aelog"
	"google.golang.org/api/cloudtasks/v2"
)

type CloudTaskService struct {
	queue   string
	service string
	tasks   *cloudtasks.Service
}

type SampleTask struct {
	Message string
	Count   int
}

func NewSampleTask(ctx context.Context) (*CloudTaskService, error) {
	pID, err := gcpmetadata.GetProjectID()
	if err != nil {
		return nil, err
	}
	region, err := gcpmetadata.GetRegion()
	if err != nil {
		return nil, err
	}
	service, err := gcpmetadata.GetAppEngineService()
	if err != nil {
		return nil, err
	}

	s, err := cloudtasks.NewService(ctx)
	if err != nil {
		return nil, err
	}
	return &CloudTaskService{
		queue:   fmt.Sprintf("projects/%s/locations/%s/queues/%s", pID, region, "sample"),
		service: service,
		tasks:   s,
	}, nil
}

func (s *CloudTaskService) AddTask(t *SampleTask) (*cloudtasks.Task, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	task, err := s.tasks.Projects.Locations.Queues.Tasks.Create(
		s.queue,
		&cloudtasks.CreateTaskRequest{
			Task: &cloudtasks.Task{
				AppEngineHttpRequest: &cloudtasks.AppEngineHttpRequest{
					AppEngineRouting: &cloudtasks.AppEngineRouting{
						Service: s.service,
					},
					RelativeUri: "/task/process",
					HttpMethod:  http.MethodPost,
					Headers:     map[string]string{"Content-Type": "application/json"},
					Body:        base64.StdEncoding.EncodeToString(b),
				},
			},
		}).Do()
	if err != nil {
		return nil, err
	}
	return task, nil
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx := aelog.WithHTTPRequest(r.Context(), r)

	t := &SampleTask{
		Message: "Hello Cloud Tasks",
		Count:   100,
	}

	s, err := NewSampleTask(ctx)
	if err != nil {
		aelog.Errorf(ctx, "Error NewSample Task %+v\n", err)
		http.Error(w, "Error NewSampleTask", http.StatusInternalServerError)
		return
	}

	_, err = s.AddTask(t)
	if err != nil {
		// aelog.Errorf(ctx, "Error AddTask Task %+v\n", err)
		fmt.Printf("Error AddTask() %+v\n", err)
		http.Error(w, "Error AddTask", http.StatusInternalServerError)
		return
	}
}

func processTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx := aelog.WithHTTPRequest(r.Context(), r)

	var task *SampleTask
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Error incrementing visitor counter", http.StatusInternalServerError)
		return
	}

	aelog.Infof(ctx, "%+v\n", task)
}
