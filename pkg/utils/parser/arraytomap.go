package parser

import "github.com/j75689/Tmaster/pkg/graph/model"

func TaskArrayToMap(tasks []*model.Task) map[string]*model.Task {
	taskMap := make(map[string]*model.Task, len(tasks))
	for idx, task := range tasks {
		taskMap[task.Name] = tasks[idx]
	}
	return taskMap
}
