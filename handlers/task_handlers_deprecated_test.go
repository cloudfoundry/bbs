package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/bbs/handlers"
	"code.cloudfoundry.org/bbs/handlers/fake_controllers"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/bbs/models/test/model_helpers"
	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task Handlers", func() {
	var (
		logger     *lagertest.TestLogger
		controller *fake_controllers.FakeTaskController

		responseRecorder *httptest.ResponseRecorder

		handler *handlers.TaskHandler
		exitCh  chan struct{}

		requestBody interface{}
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		responseRecorder = httptest.NewRecorder()
		exitCh = make(chan struct{}, 1)
		controller = &fake_controllers.FakeTaskController{}
		handler = handlers.NewTaskHandler(controller, exitCh)
	})

	Describe("DesireTask", func() {
		var (
			taskGuid   = "task-guid"
			domain     = "domain"
			oldTaskDef *models.TaskDefinition
		)

		BeforeEach(func() {
			config, err := json.Marshal(map[string]string{"foo": "bar"})
			Expect(err).NotTo(HaveOccurred())

			oldTaskDef = model_helpers.NewValidTaskDefinition()
			oldTaskDef.VolumeMounts = []*models.VolumeMount{{
				Driver:             "my-driver",
				ContainerDir:       "/mnt/mypath",
				DeprecatedMode:     models.DeprecatedBindMountMode_RO,
				DeprecatedConfig:   config,
				DeprecatedVolumeId: "my-volume",
			}}

			requestBody = &models.DesireTaskRequest{
				TaskGuid:       taskGuid,
				Domain:         domain,
				TaskDefinition: oldTaskDef,
			}

		})

		JustBeforeEach(func() {
			request := newTestRequest(requestBody)
			handler.DesireTask_r1(logger, responseRecorder, request)
		})

		Context("when the desire is successful", func() {
			It("upconverts the deprecated volume mounts", func() {
				expectedTaskDef := model_helpers.NewValidTaskDefinition()

				Expect(controller.DesireTaskCallCount()).To(Equal(1))
				_, actualTaskDef, _, _ := controller.DesireTaskArgsForCall(0)
				Expect(actualTaskDef.VolumeMounts).To(Equal(expectedTaskDef.VolumeMounts))
				Expect(actualTaskDef).To(Equal(expectedTaskDef))

				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
				response := &models.TaskLifecycleResponse{}
				err := response.Unmarshal(responseRecorder.Body.Bytes())
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Error).To(BeNil())
			})
		})
	})
})

// var _ = Describe("Task Handlers", func() {
// 	var (
// 		logger               *lagertest.TestLogger
// 		fakeTaskDB           *dbfakes.FakeTaskDB
// 		fakeAuctioneerClient *auctioneerfakes.FakeClient
// 		responseRecorder     *httptest.ResponseRecorder
// 		exitCh               chan struct{}

// 		handler *handlers.TaskHandler

// 		task1 models.Task
// 		task2 models.Task

// 		requestBody interface{}
// 	)

// 	BeforeEach(func() {
// 		fakeTaskDB = new(dbfakes.FakeTaskDB)
// 		fakeAuctioneerClient = new(auctioneerfakes.FakeClient)
// 		logger = lagertest.NewTestLogger("test")
// 		responseRecorder = httptest.NewRecorder()
// 		exitCh = make(chan struct{}, 1)
// 		handler = handlers.NewTaskHandler(logger, fakeTaskDB, nil, fakeAuctioneerClient, fakeServiceClient, fakeRepClientFactory, exitCh)
// 	})

// 	Describe("Tasks_r0", func() {
// 		BeforeEach(func() {
// 			task1 = models.Task{Domain: "domain-1"}
// 			task2 = models.Task{CellId: "cell-id"}
// 			requestBody = &models.TasksRequest{}
// 		})

// 		JustBeforeEach(func() {
// 			request := newTestRequest(requestBody)
// 			handler.Tasks_r0(responseRecorder, request)
// 		})

// 		Context("when reading tasks from DB succeeds", func() {
// 			var tasks []*models.Task

// 			BeforeEach(func() {
// 				tasks = []*models.Task{&task1, &task2}
// 				fakeTaskDB.TasksReturns(tasks, nil)
// 			})

// 			It("returns a list of task", func() {
// 				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 				response := models.TasksResponse{}
// 				err := response.Unmarshal(responseRecorder.Body.Bytes())
// 				Expect(err).NotTo(HaveOccurred())

// 				Expect(response.Error).To(BeNil())
// 				Expect(response.Tasks).To(Equal(tasks))
// 			})

// 			It("calls the DB with no filter", func() {
// 				Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
// 				_, filter := fakeTaskDB.TasksArgsForCall(0)
// 				Expect(filter).To(Equal(models.TaskFilter{}))
// 			})

// 			Context("and filtering by domain", func() {
// 				BeforeEach(func() {
// 					requestBody = &models.TasksRequest{
// 						Domain: "domain-1",
// 					}
// 				})

// 				It("calls the DB with a domain filter", func() {
// 					Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
// 					_, filter := fakeTaskDB.TasksArgsForCall(0)
// 					Expect(filter.Domain).To(Equal("domain-1"))
// 				})
// 			})

// 			Context("and filtering by cell id", func() {
// 				BeforeEach(func() {
// 					requestBody = &models.TasksRequest{
// 						CellId: "cell-id",
// 					}
// 				})

// 				It("calls the DB with a cell filter", func() {
// 					Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
// 					_, filter := fakeTaskDB.TasksArgsForCall(0)
// 					Expect(filter.CellID).To(Equal("cell-id"))
// 				})
// 			})

// 			Context("and the returned tasks have cache dependencies", func() {
// 				BeforeEach(func() {
// 					task1.TaskDefinition = &models.TaskDefinition{}
// 					task2.TaskDefinition = &models.TaskDefinition{}

// 					task1.Action = &models.Action{
// 						UploadAction: &models.UploadAction{
// 							From: "web_location",
// 						},
// 					}

// 					task1.CachedDependencies = []*models.CachedDependency{
// 						{Name: "name-1", From: "from-1", To: "to-1", CacheKey: "cache-key-1", LogSource: "log-source-1"},
// 					}

// 					task2.CachedDependencies = []*models.CachedDependency{
// 						{Name: "name-2", From: "from-2", To: "to-2", CacheKey: "cache-key-2", LogSource: "log-source-2"},
// 						{Name: "name-3", From: "from-3", To: "to-3", CacheKey: "cache-key-3", LogSource: "log-source-3"},
// 					}
// 				})

// 				It("translates the cache dependencies into download actions", func() {
// 					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 					response := models.TasksResponse{}
// 					err := response.Unmarshal(responseRecorder.Body.Bytes())
// 					Expect(err).NotTo(HaveOccurred())

// 					Expect(response.Error).To(BeNil())
// 					Expect(response.Tasks).To(HaveLen(2))
// 					Expect(response.Tasks[0]).To(Equal(task1.VersionDownTo(format.V0)))
// 					Expect(response.Tasks[1]).To(Equal(task2.VersionDownTo(format.V0)))
// 				})
// 			})
// 		})

// 		Context("when the DB returns an unrecoverable error", func() {
// 			BeforeEach(func() {
// 				fakeTaskDB.TasksReturns(nil, models.NewUnrecoverableError(nil))
// 			})

// 			It("logs and writes to the exit channel", func() {
// 				Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
// 				Eventually(exitCh).Should(Receive())
// 			})
// 		})

// 		Context("when the DB errors out", func() {
// 			BeforeEach(func() {
// 				fakeTaskDB.TasksReturns(nil, models.ErrUnknownError)
// 			})

// 			It("provides relevant error information", func() {
// 				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 				response := models.TasksResponse{}
// 				err := response.Unmarshal(responseRecorder.Body.Bytes())
// 				Expect(err).NotTo(HaveOccurred())

// 				Expect(response.Error).To(Equal(models.ErrUnknownError))
// 			})
// 		})
// 	})

// 	Describe("Tasks_r1", func() {
// 		BeforeEach(func() {
// 			task1 = models.Task{Domain: "domain-1"}
// 			task2 = models.Task{CellId: "cell-id"}
// 			requestBody = &models.TasksRequest{}
// 		})

// 		JustBeforeEach(func() {
// 			request := newTestRequest(requestBody)
// 			handler.Tasks_r1(responseRecorder, request)
// 		})

// 		Context("when reading tasks from DB succeeds", func() {
// 			var tasks []*models.Task

// 			BeforeEach(func() {
// 				tasks = []*models.Task{&task1, &task2}
// 				fakeTaskDB.TasksReturns(tasks, nil)
// 			})

// 			It("returns a list of task", func() {
// 				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 				response := models.TasksResponse{}
// 				err := response.Unmarshal(responseRecorder.Body.Bytes())
// 				Expect(err).NotTo(HaveOccurred())

// 				Expect(response.Error).To(BeNil())
// 				Expect(response.Tasks).To(Equal(tasks))
// 			})

// 			It("calls the DB with no filter", func() {
// 				Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
// 				_, filter := fakeTaskDB.TasksArgsForCall(0)
// 				Expect(filter).To(Equal(models.TaskFilter{}))
// 			})

// 			Context("and filtering by domain", func() {
// 				BeforeEach(func() {
// 					requestBody = &models.TasksRequest{
// 						Domain: "domain-1",
// 					}
// 				})

// 				It("calls the DB with a domain filter", func() {
// 					Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
// 					_, filter := fakeTaskDB.TasksArgsForCall(0)
// 					Expect(filter.Domain).To(Equal("domain-1"))
// 				})
// 			})

// 			Context("and filtering by cell id", func() {
// 				BeforeEach(func() {
// 					requestBody = &models.TasksRequest{
// 						CellId: "cell-id",
// 					}
// 				})

// 				It("calls the DB with a cell filter", func() {
// 					Expect(fakeTaskDB.TasksCallCount()).To(Equal(1))
// 					_, filter := fakeTaskDB.TasksArgsForCall(0)
// 					Expect(filter.CellID).To(Equal("cell-id"))
// 				})
// 			})

// 			Context("and the tasks have timeout not timeout_ms", func() {
// 				BeforeEach(func() {
// 					task1.TaskDefinition = &models.TaskDefinition{}
// 					task2.TaskDefinition = &models.TaskDefinition{}

// 					task1.Action = &models.Action{
// 						TimeoutAction: &models.TimeoutAction{
// 							Action: models.WrapAction(&models.UploadAction{
// 								From: "web_location",
// 							}),
// 							TimeoutMs: 10000,
// 						},
// 					}
// 				})

// 				It("translates the timeoutMs to timeout", func() {
// 					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 					response := models.TasksResponse{}
// 					err := response.Unmarshal(responseRecorder.Body.Bytes())
// 					Expect(err).NotTo(HaveOccurred())

// 					Expect(response.Error).To(BeNil())
// 					Expect(response.Tasks).To(HaveLen(2))
// 					Expect(response.Tasks[0]).To(Equal(task1.VersionDownTo(format.V1)))
// 					Expect(response.Tasks[1]).To(Equal(task2.VersionDownTo(format.V1)))
// 				})
// 			})
// 		})

// 		Context("when the DB returns an unrecoverable error", func() {
// 			BeforeEach(func() {
// 				fakeTaskDB.TasksReturns(nil, models.NewUnrecoverableError(nil))
// 			})

// 			It("logs and writes to the exit channel", func() {
// 				Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
// 				Eventually(exitCh).Should(Receive())
// 			})
// 		})

// 		Context("when the DB errors out", func() {
// 			BeforeEach(func() {
// 				fakeTaskDB.TasksReturns(nil, models.ErrUnknownError)
// 			})

// 			It("provides relevant error information", func() {
// 				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 				response := models.TasksResponse{}
// 				err := response.Unmarshal(responseRecorder.Body.Bytes())
// 				Expect(err).NotTo(HaveOccurred())

// 				Expect(response.Error).To(Equal(models.ErrUnknownError))
// 			})
// 		})
// 	})

// 	Describe("TaskByGuid_r0", func() {
// 		var taskGuid = "task-guid"

// 		BeforeEach(func() {
// 			requestBody = &models.TaskByGuidRequest{
// 				TaskGuid: taskGuid,
// 			}
// 		})

// 		JustBeforeEach(func() {
// 			request := newTestRequest(requestBody)
// 			handler.TaskByGuid_r0(responseRecorder, request)
// 		})

// 		Context("when reading a task from the DB succeeds", func() {
// 			var task *models.Task

// 			BeforeEach(func() {
// 				task = &models.Task{TaskGuid: taskGuid}
// 				fakeTaskDB.TaskByGuidReturns(task, nil)
// 			})

// 			It("fetches task by guid", func() {
// 				Expect(fakeTaskDB.TaskByGuidCallCount()).To(Equal(1))
// 				_, actualGuid := fakeTaskDB.TaskByGuidArgsForCall(0)
// 				Expect(actualGuid).To(Equal(taskGuid))
// 			})

// 			It("returns the task", func() {
// 				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 				response := models.TaskResponse{}
// 				err := response.Unmarshal(responseRecorder.Body.Bytes())
// 				Expect(err).NotTo(HaveOccurred())

// 				Expect(response.Error).To(BeNil())
// 				Expect(response.Task).To(Equal(task))
// 			})

// 			Context("when the task has cache dependencies", func() {
// 				BeforeEach(func() {
// 					task.TaskDefinition = &models.TaskDefinition{}
// 					task.CachedDependencies = []*models.CachedDependency{
// 						{Name: "name-2", From: "from-2", To: "to-2", CacheKey: "cache-key-2", LogSource: "log-source-2"},
// 						{Name: "name-3", From: "from-3", To: "to-3", CacheKey: "cache-key-3", LogSource: "log-source-3"},
// 					}
// 				})

// 				It("moves them to the actions", func() {
// 					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 					response := models.TaskResponse{}
// 					err := response.Unmarshal(responseRecorder.Body.Bytes())
// 					Expect(err).NotTo(HaveOccurred())

// 					Expect(response.Error).To(BeNil())
// 					Expect(response.Task).To(Equal(task.VersionDownTo(format.V0)))
// 				})
// 			})

// 			Context("and the tasks have timeout not timeout_ms", func() {
// 				BeforeEach(func() {
// 					task.TaskDefinition = &models.TaskDefinition{}

// 					task.Action = &models.Action{
// 						TimeoutAction: &models.TimeoutAction{
// 							Action: models.WrapAction(&models.UploadAction{
// 								From: "web_location",
// 							}),
// 							TimeoutMs: 10000,
// 						},
// 					}
// 				})

// 				It("translates the timeoutMs to timeout", func() {
// 					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 					response := models.TaskResponse{}
// 					err := response.Unmarshal(responseRecorder.Body.Bytes())
// 					Expect(err).NotTo(HaveOccurred())

// 					Expect(response.Error).To(BeNil())
// 					Expect(response.Task).To(Equal(task.VersionDownTo(format.V1)))
// 				})
// 			})
// 		})

// 		Context("when the DB returns an unrecoverable error", func() {
// 			BeforeEach(func() {
// 				fakeTaskDB.TaskByGuidReturns(nil, models.NewUnrecoverableError(nil))
// 			})

// 			It("logs and writes to the exit channel", func() {
// 				Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
// 				Eventually(exitCh).Should(Receive())
// 			})
// 		})

// 		Context("when the DB returns no task", func() {
// 			BeforeEach(func() {
// 				fakeTaskDB.TaskByGuidReturns(nil, models.ErrResourceNotFound)
// 			})

// 			It("returns a resource not found error", func() {
// 				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 				response := models.TaskResponse{}
// 				err := response.Unmarshal(responseRecorder.Body.Bytes())
// 				Expect(err).NotTo(HaveOccurred())

// 				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
// 			})
// 		})

// 		Context("when the DB errors out", func() {
// 			BeforeEach(func() {
// 				fakeTaskDB.TaskByGuidReturns(nil, models.ErrUnknownError)
// 			})

// 			It("provides relevant error information", func() {
// 				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 				response := models.TaskResponse{}
// 				err := response.Unmarshal(responseRecorder.Body.Bytes())
// 				Expect(err).NotTo(HaveOccurred())

// 				Expect(response.Error).To(Equal(models.ErrUnknownError))
// 			})
// 		})
// 	})

// 	Describe("TaskByGuid_r1", func() {
// 		var taskGuid = "task-guid"

// 		BeforeEach(func() {
// 			requestBody = &models.TaskByGuidRequest{
// 				TaskGuid: taskGuid,
// 			}
// 		})

// 		JustBeforeEach(func() {
// 			request := newTestRequest(requestBody)
// 			handler.TaskByGuid_r1(responseRecorder, request)
// 		})

// 		Context("when reading a task from the DB succeeds", func() {
// 			var task *models.Task

// 			BeforeEach(func() {
// 				task = &models.Task{TaskGuid: taskGuid}
// 				fakeTaskDB.TaskByGuidReturns(task, nil)
// 			})

// 			It("fetches task by guid", func() {
// 				Expect(fakeTaskDB.TaskByGuidCallCount()).To(Equal(1))
// 				_, actualGuid := fakeTaskDB.TaskByGuidArgsForCall(0)
// 				Expect(actualGuid).To(Equal(taskGuid))
// 			})

// 			It("returns the task", func() {
// 				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 				response := models.TaskResponse{}
// 				err := response.Unmarshal(responseRecorder.Body.Bytes())
// 				Expect(err).NotTo(HaveOccurred())

// 				Expect(response.Error).To(BeNil())
// 				Expect(response.Task).To(Equal(task))
// 			})

// 			Context("when the task has cache dependencies", func() {
// 				BeforeEach(func() {
// 					task.TaskDefinition = &models.TaskDefinition{}
// 					task.CachedDependencies = []*models.CachedDependency{
// 						{Name: "name-2", From: "from-2", To: "to-2", CacheKey: "cache-key-2", LogSource: "log-source-2"},
// 						{Name: "name-3", From: "from-3", To: "to-3", CacheKey: "cache-key-3", LogSource: "log-source-3"},
// 					}
// 				})

// 				It("moves them to the actions", func() {
// 					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 					response := models.TaskResponse{}
// 					err := response.Unmarshal(responseRecorder.Body.Bytes())
// 					Expect(err).NotTo(HaveOccurred())

// 					Expect(response.Error).To(BeNil())
// 					Expect(response.Task).To(Equal(task.VersionDownTo(format.V1)))
// 				})
// 			})

// 			Context("and the tasks have timeout not timeout_ms", func() {
// 				BeforeEach(func() {
// 					task.TaskDefinition = &models.TaskDefinition{}

// 					task.Action = &models.Action{
// 						TimeoutAction: &models.TimeoutAction{
// 							Action: models.WrapAction(&models.UploadAction{
// 								From: "web_location",
// 							}),
// 							TimeoutMs: 10000,
// 						},
// 					}
// 				})

// 				It("translates the timeoutMs to timeout", func() {
// 					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 					response := models.TaskResponse{}
// 					err := response.Unmarshal(responseRecorder.Body.Bytes())
// 					Expect(err).NotTo(HaveOccurred())

// 					Expect(response.Error).To(BeNil())
// 					Expect(response.Task).To(Equal(task.VersionDownTo(format.V1)))
// 				})
// 			})
// 		})

// 		Context("when the DB returns an unrecoverable error", func() {
// 			BeforeEach(func() {
// 				fakeTaskDB.TaskByGuidReturns(nil, models.NewUnrecoverableError(nil))
// 			})

// 			It("logs and writes to the exit channel", func() {
// 				Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
// 				Eventually(exitCh).Should(Receive())
// 			})
// 		})

// 		Context("when the DB returns no task", func() {
// 			BeforeEach(func() {
// 				fakeTaskDB.TaskByGuidReturns(nil, models.ErrResourceNotFound)
// 			})

// 			It("returns a resource not found error", func() {
// 				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 				response := models.TaskResponse{}
// 				err := response.Unmarshal(responseRecorder.Body.Bytes())
// 				Expect(err).NotTo(HaveOccurred())

// 				Expect(response.Error).To(Equal(models.ErrResourceNotFound))
// 			})
// 		})

// 		Context("when the DB errors out", func() {
// 			BeforeEach(func() {
// 				fakeTaskDB.TaskByGuidReturns(nil, models.ErrUnknownError)
// 			})

// 			It("provides relevant error information", func() {
// 				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 				response := models.TaskResponse{}
// 				err := response.Unmarshal(responseRecorder.Body.Bytes())
// 				Expect(err).NotTo(HaveOccurred())

// 				Expect(response.Error).To(Equal(models.ErrUnknownError))
// 			})
// 		})
// 	})

// 	Describe("DesireTask_r0", func() {
// 		var (
// 			taskGuid = "task-guid"
// 			domain   = "domain"
// 			taskDef  *models.TaskDefinition
// 		)

// 		BeforeEach(func() {
// 			taskDef = model_helpers.NewValidTaskDefinition()
// 			taskDef.Action = &models.Action{
// 				TimeoutAction: &models.TimeoutAction{
// 					Action: models.WrapAction(&models.UploadAction{
// 						From: "web_location",
// 						To:   "potato",
// 						User: "face",
// 					}),
// 					DeprecatedTimeoutNs: int64(time.Second),
// 				},
// 			}
// 			requestBody = &models.DesireTaskRequest{
// 				TaskGuid:       taskGuid,
// 				Domain:         domain,
// 				TaskDefinition: taskDef,
// 			}
// 		})

// 		JustBeforeEach(func() {
// 			request := newTestRequest(requestBody)
// 			handler.DesireTask_r0(responseRecorder, request)
// 		})

// 		Context("when the desire is successful", func() {
// 			It("desires the task with the requested definitions", func() {
// 				Expect(fakeTaskDB.DesireTaskCallCount()).To(Equal(1))
// 				_, actualTaskDef, actualTaskGuid, actualDomain := fakeTaskDB.DesireTaskArgsForCall(0)
// 				taskDef.Action = &models.Action{
// 					TimeoutAction: &models.TimeoutAction{
// 						Action: models.WrapAction(&models.UploadAction{
// 							From: "web_location",
// 							To:   "potato",
// 							User: "face",
// 						}),
// 						DeprecatedTimeoutNs: int64(time.Second),
// 						TimeoutMs:           1000,
// 					},
// 				}
// 				Expect(actualTaskDef).To(Equal(taskDef))
// 				Expect(actualTaskGuid).To(Equal(taskGuid))
// 				Expect(actualDomain).To(Equal(domain))

// 				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 				response := &models.TaskLifecycleResponse{}
// 				err := response.Unmarshal(responseRecorder.Body.Bytes())
// 				Expect(err).NotTo(HaveOccurred())

// 				Expect(response.Error).To(BeNil())
// 			})

// 			It("requests an auction", func() {
// 				Expect(fakeAuctioneerClient.RequestTaskAuctionsCallCount()).To(Equal(1))

// 				var volumeMounts []string
// 				for _, volMount := range taskDef.VolumeMounts {
// 					volumeMounts = append(volumeMounts, volMount.Driver)
// 				}

// 				expectedStartRequest := auctioneer.TaskStartRequest{
// 					Task: rep.Task{
// 						TaskGuid: taskGuid,
// 						Domain:   domain,
// 						Resource: rep.Resource{
// 							MemoryMB:      256,
// 							DiskMB:        1024,
// 							RootFs:        "docker:///docker.com/docker",
// 							VolumeDrivers: volumeMounts,
// 						},
// 					},
// 				}

// 				requestedTasks := fakeAuctioneerClient.RequestTaskAuctionsArgsForCall(0)
// 				Expect(requestedTasks).To(HaveLen(1))
// 				Expect(*requestedTasks[0]).To(Equal(expectedStartRequest))
// 			})

// 			Context("when requesting a task auction succeeds", func() {
// 				BeforeEach(func() {
// 					fakeAuctioneerClient.RequestTaskAuctionsReturns(nil)
// 				})

// 				It("does not return an error", func() {
// 					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 					response := &models.TaskLifecycleResponse{}
// 					err := response.Unmarshal(responseRecorder.Body.Bytes())
// 					Expect(err).NotTo(HaveOccurred())

// 					Expect(response.Error).To(BeNil())
// 				})
// 			})

// 			Context("when requesting a task auction fails", func() {
// 				BeforeEach(func() {
// 					fakeAuctioneerClient.RequestTaskAuctionsReturns(errors.New("oops"))
// 				})

// 				It("does not return an error", func() {
// 					Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 					response := &models.TaskLifecycleResponse{}
// 					err := response.Unmarshal(responseRecorder.Body.Bytes())
// 					Expect(err).NotTo(HaveOccurred())

// 					Expect(response.Error).To(BeNil())
// 				})

// 				It("does not request a second auction", func() {
// 					Consistently(fakeAuctioneerClient.RequestTaskAuctionsCallCount).Should(Equal(1))
// 				})
// 			})
// 		})

// 		Context("when the DB returns an unrecoverable error", func() {
// 			BeforeEach(func() {
// 				fakeTaskDB.DesireTaskReturns(models.NewUnrecoverableError(nil))
// 			})

// 			It("logs and writes to the exit channel", func() {
// 				Eventually(logger).Should(gbytes.Say("unrecoverable-error"))
// 				Eventually(exitCh).Should(Receive())
// 			})
// 		})

// 		Context("when desiring the task fails", func() {
// 			BeforeEach(func() {
// 				fakeTaskDB.DesireTaskReturns(models.ErrUnknownError)
// 			})

// 			It("responds with an error", func() {
// 				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
// 				response := &models.TaskLifecycleResponse{}
// 				err := response.Unmarshal(responseRecorder.Body.Bytes())
// 				Expect(err).NotTo(HaveOccurred())

// 				Expect(response.Error).To(Equal(models.ErrUnknownError))
// 			})
// 		})
// 	})
// })
