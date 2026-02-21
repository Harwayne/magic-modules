package dataproc

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-provider-google/google/tpgresource"
	transport_tpg "github.com/hashicorp/terraform-provider-google/google/transport"

	"google.golang.org/api/dataproc/v1"
)

type dataprocJobOperationWaiter struct {
	Service   *dataproc.Service
	Region    string
	ProjectId string
	JobId     string
	Status    string
}

func (w *dataprocJobOperationWaiter) State() string {
	if w == nil {
		return "<nil>"
	}
	return w.Status
}

func (w *dataprocJobOperationWaiter) Error() error {
	// The "operation" is just the job, which has no special error field that we
	// want to expose.
	return nil
}

func (w *dataprocJobOperationWaiter) IsRetryable(error) bool {
	return false
}

func (w *dataprocJobOperationWaiter) SetOp(job interface{}) error {
	// The "operation" is just the job. Instead of holding onto the whole job
	// object, we only care about the state, which gets set in QueryOp, so this
	// doesn't have to do anything.
	return nil
}

func (w *dataprocJobOperationWaiter) QueryOp() (interface{}, error) {
	if w == nil {
		return nil, fmt.Errorf("Cannot query operation, it's unset or nil.")
	}
	job, err := w.Service.Projects.Regions.Jobs.Get(w.ProjectId, w.Region, w.JobId).Do()
	if job != nil {
		w.Status = job.Status.State
	}
	return job, err
}

func (w *dataprocJobOperationWaiter) OpName() string {
	if w == nil {
		return "<nil>"
	}
	return w.JobId
}

type DataprocSubmitJobOperationWaiter struct {
	dataprocJobOperationWaiter
}

func (w *DataprocSubmitJobOperationWaiter) PendingStates() []string {
	return []string{"PENDING", "CANCEL_PENDING", "CANCEL_STARTED", "SETUP_DONE"}
}

func (w *DataprocSubmitJobOperationWaiter) TargetStates() []string {
	return []string{"CANCELLED", "DONE", "ATTEMPT_FAILURE", "ERROR", "RUNNING"}
}

func DataprocSubmitJobOperationWait(config *transport_tpg.Config, region, projectId, jobId, activity, userAgent string, timeout time.Duration) error {
	w := &DataprocSubmitJobOperationWaiter{
		dataprocJobOperationWaiter{
			Service:   config.NewDataprocClient(userAgent),
			Region:    region,
			ProjectId: projectId,
			JobId:     jobId,
		},
	}
	return tpgresource.OperationWait(w, activity, timeout, config.PollInterval)
}

type DataprocCancelJobOperationWaiter struct {
	dataprocJobOperationWaiter
}

func (w *DataprocCancelJobOperationWaiter) PendingStates() []string {
	return []string{"PENDING", "CANCEL_PENDING", "CANCEL_STARTED", "SETUP_DONE", "RUNNING"}
}

func (w *DataprocCancelJobOperationWaiter) TargetStates() []string {
	return []string{"CANCELLED", "DONE", "ATTEMPT_FAILURE", "ERROR"}
}

func DataprocCancelJobOperationWait(config *transport_tpg.Config, region, projectId, jobId, activity, userAgent string, timeout time.Duration) error {
	w := &DataprocCancelJobOperationWaiter{
		dataprocJobOperationWaiter{
			Service:   config.NewDataprocClient(userAgent),
			Region:    region,
			ProjectId: projectId,
			JobId:     jobId,
		},
	}
	return tpgresource.OperationWait(w, activity, timeout, config.PollInterval)
}

type DataprocDeleteJobOperationWaiter struct {
	dataprocJobOperationWaiter
}

func (w *DataprocDeleteJobOperationWaiter) PendingStates() []string {
	return []string{"EXISTS", "ERROR"}
}

func (w *DataprocDeleteJobOperationWaiter) TargetStates() []string {
	return []string{"DELETED"}
}

func (w *DataprocDeleteJobOperationWaiter) QueryOp() (interface{}, error) {
	if w == nil {
		return nil, fmt.Errorf("Cannot query operation, it's unset or nil.")
	}
	job, err := w.Service.Projects.Regions.Jobs.Get(w.ProjectId, w.Region, w.JobId).Do()
	if err != nil {
		if transport_tpg.IsGoogleApiErrorWithCode(err, http.StatusNotFound) {
			w.Status = "DELETED"
			return job, nil
		}
		w.Status = "ERROR"
	}
	w.Status = "EXISTS"
	return job, err
}

func DataprocDeleteJobOperationWait(config *transport_tpg.Config, region, projectId, jobId, activity, userAgent string, timeout time.Duration) error {
	w := &DataprocDeleteJobOperationWaiter{
		dataprocJobOperationWaiter{
			Service:   config.NewDataprocClient(userAgent),
			Region:    region,
			ProjectId: projectId,
			JobId:     jobId,
		},
	}
	return tpgresource.OperationWait(w, activity, timeout, config.PollInterval)
}
