package matchers

import (
	"code.cloudfoundry.org/bbs/models"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func MatchTask(task *models.Task) types.GomegaMatcher {
	return gomega.SatisfyAll(
		gomega.WithTransform(func(t *models.Task) string {
			return t.TaskGuid
		}, gomega.Equal(task.TaskGuid)),
		gomega.WithTransform(func(t *models.Task) string {
			return t.Domain
		}, gomega.Equal(task.Domain)),
		gomega.WithTransform(func(t *models.Task) *models.TaskDefinition {
			return t.TaskDefinition
		}, gomega.Equal(task.TaskDefinition)),
	)
}

func MatchTasks(tasks []*models.Task) types.GomegaMatcher {
	matchers := []types.GomegaMatcher{}
	matchers = append(matchers, gomega.HaveLen(len(tasks)))

	for _, task := range tasks {
		matchers = append(matchers, gomega.ContainElement(MatchTask(task)))
	}

	return gomega.SatisfyAll(matchers...)
}
