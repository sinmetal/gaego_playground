package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/cloudtasks/apiv2"
	"github.com/sinmetal/gcpmetadata"
	"github.com/vvakame/sdlog/aelog"
	"google.golang.org/genproto/googleapis/cloud/tasks/v2"
)

type CloudTaskService struct {
	queue               string
	service             string
	serviceAccountEmail string
	client              *cloudtasks.Client
}

type SampleTask struct {
	Message string
	Count   int
}

func NewSampleTask(ctx context.Context, client *cloudtasks.Client) (*CloudTaskService, error) {
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

	sae, err := gcpmetadata.GetServiceAccountEmail()
	if err != nil {
		return nil, err
	}

	return &CloudTaskService{
		queue:               fmt.Sprintf("projects/%s/locations/%s/queues/%s", pID, region, "sample"),
		service:             service,
		serviceAccountEmail: sae,
		client:              client,
	}, nil
}

func (s *CloudTaskService) CreateAppEngineTask(ctx context.Context, t *SampleTask) (*tasks.Task, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	task, err := s.client.CreateTask(ctx, &tasks.CreateTaskRequest{
		Parent: s.queue,
		Task: &tasks.Task{
			MessageType: &tasks.Task_AppEngineHttpRequest{
				AppEngineHttpRequest: &tasks.AppEngineHttpRequest{
					HttpMethod: tasks.HttpMethod_POST,
					AppEngineRouting: &tasks.AppEngineRouting{
						Service: s.service,
					},
					RelativeUri: "/task/process",
					Headers:     map[string]string{"Content-Type": "application/json"},
					Body:        b,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return task, nil
}

// CreateHttpTask
// OIDC_Token指定でHttp TaskをCloudTaskに追加するサンプル
// 動かしてはいない
func (s *CloudTaskService) CreateHttpTask(ctx context.Context, t *SampleTask) (*tasks.Task, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	task, err := s.client.CreateTask(ctx, &tasks.CreateTaskRequest{
		Parent: s.queue,
		Task: &tasks.Task{
			MessageType: &tasks.Task_HttpRequest{
				HttpRequest: &tasks.HttpRequest{
					Url:        "https://example.com/tasks/process",
					HttpMethod: tasks.HttpMethod_POST,
					Headers:    map[string]string{"Content-Type": "application/json"},
					Body:       b,
					AuthorizationHeader: &tasks.HttpRequest_OidcToken{OidcToken: &tasks.OidcToken{
						ServiceAccountEmail: s.serviceAccountEmail,
					}},
				},
			},
		},
	})
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

	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		aelog.Errorf(ctx, "Error cloudtasks.NewClient  %+v\n", err)
		http.Error(w, "Error NewSampleTask", http.StatusInternalServerError)
		return
	}

	s, err := NewSampleTask(ctx, client)
	if err != nil {
		aelog.Errorf(ctx, "Error NewSample Task %+v\n", err)
		http.Error(w, "Error NewSampleTask", http.StatusInternalServerError)
		return
	}

	_, err = s.CreateAppEngineTask(ctx, t)
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
