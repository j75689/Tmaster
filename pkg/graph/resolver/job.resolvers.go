package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"time"

	dbmodel "github.com/j75689/Tmaster/pkg/database/model"
	"github.com/j75689/Tmaster/pkg/graph/generated"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"github.com/j75689/Tmaster/pkg/message"
	"github.com/j75689/Tmaster/pkg/utils/gzip"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rs/xid"
)

func (r *mutationResolver) CreateJob(ctx context.Context, input *model.Job) (string, error) {
	id := xid.New().String()
	r.logger.Info().Msg("received new job: " + id)

	var carrier opentracing.TextMapCarrier
	if r.config.OpenTracing.Enable {
		traceRecord, err := r.tracer.TraceClient("CreateJob", map[string]interface{}{"job_id": id})
		if err != nil {
			r.logger.Err(err).Msg("create trace error")
		}
		defer traceRecord.Finish()
		carrier = traceRecord.Carrier()
	}

	r.logger.Info().Msg("marshal data new job: " + id)
	data, err := json.Marshal(&message.InitJob{
		Context: message.Context{
			TraceCarrier: carrier,
		},
		JobID: id,
		Job:   *input,
	})
	if err != nil {
		return "", err
	}
	r.logger.Info().Msg("pre create new job: " + id)
	data, err = gzip.Compress(data)
	if err != nil {
		return "", err
	}
	if err = r.Resolver.mq.Publish(
		time.Now().UnixNano(),
		r.config.JobEndpoint.InitJob.ProjectID,
		r.config.JobEndpoint.InitJob.Topic,
		data,
	); err != nil {
		return "", err
	}
	r.logger.Info().Msg("create new job: " + id)
	return id, nil
}

func (r *queryResolver) GetJob(ctx context.Context, id *int) (*model.JobStatus, error) {
	traceRecord, err := r.tracer.TraceClient("GetJob", map[string]interface{}{"job_id": *id})
	if err != nil {
		r.logger.Err(err).Msg("create trace error")
	}
	defer traceRecord.Finish()

	job := dbmodel.Job{}
	if err := r.Resolver.db.Where("id=?", *id).First(&job).Error; err != nil {
		return nil, err
	}

	task := []*dbmodel.Task{}
	if err := r.Resolver.db.Where("job_id=?", *id).Find(&task).Error; err != nil {
		return nil, err
	}

	for _, details := range task {
		job.JobStatus.TaskHistory = append(job.JobStatus.TaskHistory, details.TaskHistory)
	}
	return job.JobStatus, nil
}

func (r *queryResolver) GetJobs(ctx context.Context, id []*int) ([]*model.JobStatus, error) {
	traceRecord, err := r.tracer.TraceClient("GetJobs", map[string]interface{}{"job_ids": id})
	if err != nil {
		r.logger.Err(err).Msg("create trace error")
	}
	defer traceRecord.Finish()

	jobs := []*dbmodel.Job{}
	var jobStatus []*model.JobStatus

	if err := r.Resolver.db.Where("id in (?)", id).Find(&jobs).Error; err != nil {
		return nil, err
	}

	for idx, id := range id {
		task := []*dbmodel.Task{}

		if err := r.Resolver.db.Where("job_id=?", id).Find(&task).Error; err != nil {
			return nil, err
		}

		for _, details := range task {
			jobs[idx].JobStatus.TaskHistory = append(jobs[idx].JobStatus.TaskHistory, details.TaskHistory)
		}
	}

	for _, details := range jobs {
		jobStatus = append(jobStatus, details.JobStatus)
	}

	return jobStatus, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
