package models

import (
	"net/url"
	"regexp"

	"github.com/cloudfoundry-incubator/bbs/format"
	"github.com/pivotal-golang/lager"
)

var taskGuidPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type TaskFilter struct {
	Domain string
	CellID string
}

func (t *Task) Version() format.Version {
	return format.V1
}

func (t *Task) MigrateFromVersion(v format.Version) error {
	return nil
}

func (t *Task) LagerData() lager.Data {
	return lager.Data{
		"task-guid": t.TaskGuid,
		"domain":    t.Domain,
		"state":     t.State,
		"cell-id":   t.CellId,
	}
}

func (task *Task) Validate() error {
	var validationError ValidationError

	if task.Domain == "" {
		validationError = validationError.Append(ErrInvalidField{"domain"})
	}

	if !taskGuidPattern.MatchString(task.TaskGuid) {
		validationError = validationError.Append(ErrInvalidField{"task_guid"})
	}

	if task.TaskDefinition == nil {
		validationError = validationError.Append(ErrInvalidField{"task_definition"})
	} else if defErr := task.TaskDefinition.Validate(); defErr != nil {
		validationError = validationError.Append(defErr)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}

func (t *Task) Copy() *Task {
	newTask := *t
	return &newTask
}

func (t *Task) VersionDownTo(v format.Version) *Task {
	t = t.Copy()
	switch v {
	case format.V0:
		t.TaskDefinition = newTaskDefWithCachedDependenciesAsActions(t.TaskDefinition)
		return t
	default:
		return t
	}
}

func newTaskDefWithCachedDependenciesAsActions(t *TaskDefinition) *TaskDefinition {
	t = t.Copy()
	if len(t.CachedDependencies) > 0 {
		cachedDownloads := Parallel(t.actionsFromCachedDependencies()...)
		if t.Action != nil {
			t.Action = WrapAction(Serial(cachedDownloads, UnwrapAction(t.Action)))
		} else {
			t.Action = WrapAction(Serial(cachedDownloads))
		}
		t.CachedDependencies = nil
	}
	return t
}

func (t *TaskDefinition) actionsFromCachedDependencies() []ActionInterface {
	actions := make([]ActionInterface, len(t.CachedDependencies))
	for i := range t.CachedDependencies {
		cacheDependency := t.CachedDependencies[i]
		actions[i] = &DownloadAction{
			Artifact:  cacheDependency.Name,
			From:      cacheDependency.From,
			To:        cacheDependency.To,
			CacheKey:  cacheDependency.CacheKey,
			LogSource: cacheDependency.LogSource,
			User:      t.LegacyDownloadUser,
		}
	}
	return actions
}

func (t *TaskDefinition) Copy() *TaskDefinition {
	newTaskDef := *t
	return &newTaskDef
}
func (def *TaskDefinition) Validate() error {
	var validationError ValidationError

	if def.RootFs == "" {
		validationError = validationError.Append(ErrInvalidField{"rootfs"})
	} else {
		rootFsURL, err := url.Parse(def.RootFs)
		if err != nil || rootFsURL.Scheme == "" {
			validationError = validationError.Append(ErrInvalidField{"rootfs"})
		}
	}

	if def.Action == nil {
		validationError = validationError.Append(ErrInvalidActionType)
	} else if err := def.Action.Validate(); err != nil {
		validationError = validationError.Append(ErrInvalidField{"action"})
		validationError = validationError.Append(err)
	}

	if def.CpuWeight > 100 {
		validationError = validationError.Append(ErrInvalidField{"cpu_weight"})
	}

	if len(def.Annotation) > maximumAnnotationLength {
		validationError = validationError.Append(ErrInvalidField{"annotation"})
	}

	for _, rule := range def.EgressRules {
		err := rule.Validate()
		if err != nil {
			validationError = validationError.Append(ErrInvalidField{"egress_rules"})
		}
	}

	err := validateCachedDependencies(def.CachedDependencies, def.LegacyDownloadUser)
	if err != nil {
		validationError = validationError.Append(err)
	}

	if !validationError.Empty() {
		return validationError
	}

	return nil
}
