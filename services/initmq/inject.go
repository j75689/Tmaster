package initmq

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/mq"
)

type Application struct {
	config config.Config
	logger zerolog.Logger
	mq     mq.MQ
}

func (application Application) InitGraphqlMQ() error {
	ctx := context.Background()
	application.logger.Info().Msg("prepare topic: " + application.config.JobEndpoint.InitJob.Topic)
	if err := application.mq.InitTopic(
		ctx,
		application.config.JobEndpoint.InitJob.ProjectID,
		application.config.JobEndpoint.InitJob.Topic,
	); err != nil {
		return err
	}
	return nil
}

func (application Application) InitInitializerMQ() error {
	ctx := context.Background()
	application.logger.Info().Msg("prepare topic: " + application.config.JobInitializer.TaskInput.Topic)
	if err := application.mq.InitTopic(
		ctx,
		application.config.JobInitializer.TaskInput.ProjectID,
		application.config.JobInitializer.TaskInput.Topic,
	); err != nil {
		return err
	}
	application.logger.Info().Msg("prepare subscribe: " + application.config.JobInitializer.InitJob.SubscribeID)
	if err := application.mq.InitSubscriber(
		ctx,
		application.config.JobInitializer.InitJob.ProjectID,
		application.config.JobInitializer.InitJob.Topic,
		application.config.JobInitializer.InitJob.SubscribeID,
	); err != nil {
		return err
	}
	return nil
}

func (application Application) InitSchedulerMQ() error {
	ctx := context.Background()
	application.logger.Info().Msg("prepare topic: " + application.config.TaskScheduler.TaskInput.Topic)
	if err := application.mq.InitTopic(
		ctx,
		application.config.TaskScheduler.TaskInput.ProjectID,
		application.config.TaskScheduler.TaskInput.Topic,
	); err != nil {
		return err
	}
	application.logger.Info().Msg("prepare subscribe: " + application.config.TaskScheduler.TaskOutput.SubscribeID)
	if err := application.mq.InitSubscriber(
		ctx,
		application.config.TaskScheduler.TaskOutput.ProjectID,
		application.config.TaskScheduler.TaskOutput.Topic,
		application.config.TaskScheduler.TaskOutput.SubscribeID,
	); err != nil {
		return err
	}
	return nil
}

func (application Application) InitWorkerMQ() error {
	ctx := context.Background()
	application.logger.Info().Msg("prepare topic: " + application.config.TaskWorker.TaskOutput.Topic)
	if err := application.mq.InitTopic(
		ctx,
		application.config.TaskWorker.TaskOutput.ProjectID,
		application.config.TaskWorker.TaskOutput.Topic,
	); err != nil {
		return err
	}
	application.logger.Info().Msg("prepare subscribe: " + application.config.TaskWorker.TaskInput.SubscribeID)
	if err := application.mq.InitSubscriber(
		ctx,
		application.config.TaskWorker.TaskInput.ProjectID,
		application.config.TaskWorker.TaskInput.Topic,
		application.config.TaskWorker.TaskInput.SubscribeID,
	); err != nil {
		return err
	}
	return nil
}

func (application Application) InitDBHelperMQ() error {
	ctx := context.Background()
	application.logger.Info().Msg("prepare topic: " + application.config.TaskScheduler.JobDBHelper.Topic)
	if err := application.mq.InitTopic(
		ctx,
		application.config.TaskScheduler.JobDBHelper.ProjectID,
		application.config.TaskScheduler.JobDBHelper.Topic,
	); err != nil {
		return err
	}
	application.logger.Info().Msg("prepare topic: " + application.config.TaskScheduler.TaskDBHelper.Topic)
	if err := application.mq.InitTopic(
		ctx,
		application.config.TaskScheduler.TaskDBHelper.ProjectID,
		application.config.TaskScheduler.TaskDBHelper.Topic,
	); err != nil {
		return err
	}

	application.logger.Info().Msg("prepare subscribe: " + application.config.DBHelper.Job.SubscribeID)
	if err := application.mq.InitSubscriber(
		ctx,
		application.config.DBHelper.Job.ProjectID,
		application.config.DBHelper.Job.Topic,
		application.config.DBHelper.Job.SubscribeID,
	); err != nil {
		return err
	}
	application.logger.Info().Msg("prepare subscribe: " + application.config.DBHelper.Task.SubscribeID)
	if err := application.mq.InitSubscriber(
		ctx,
		application.config.DBHelper.Task.ProjectID,
		application.config.DBHelper.Task.Topic,
		application.config.DBHelper.Task.SubscribeID,
	); err != nil {
		return err
	}
	return nil
}

func newApplication(config config.Config, logger zerolog.Logger, mq mq.MQ) Application {
	return Application{
		config: config,
		logger: logger,
		mq:     mq,
	}
}
