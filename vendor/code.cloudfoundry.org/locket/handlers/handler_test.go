package handlers_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/diego-db-helpers/sqldb/helpers"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"code.cloudfoundry.org/locket/db"
	"code.cloudfoundry.org/locket/db/dbfakes"
	"code.cloudfoundry.org/locket/expiration/expirationfakes"
	"code.cloudfoundry.org/locket/handlers"
	"code.cloudfoundry.org/locket/metrics/helpers/helpersfakes"
	"code.cloudfoundry.org/locket/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"google.golang.org/grpc/metadata"
)

var _ = Describe("LocketHandler", func() {
	var (
		fakeLockDB         *dbfakes.FakeLockDB
		fakeLockPick       *expirationfakes.FakeLockPick
		logger             *lagertest.TestLogger
		locketHandler      models.LocketServer
		resource           *models.Resource
		exitCh             chan struct{}
		fakeRequestMetrics *helpersfakes.FakeRequestMetrics
	)

	BeforeEach(func() {
		fakeLockDB = &dbfakes.FakeLockDB{}
		fakeLockPick = &expirationfakes.FakeLockPick{}
		fakeRequestMetrics = &helpersfakes.FakeRequestMetrics{}

		logger = lagertest.NewTestLogger("locket-handler")
		exitCh = make(chan struct{}, 1)

		resource = &models.Resource{
			Key:      "test",
			Value:    "test-value",
			Owner:    "myself",
			TypeCode: models.LOCK,
		}

		locketHandler = handlers.NewLocketHandler(
			logger,
			fakeLockDB,
			fakeLockPick,
			fakeRequestMetrics,
			exitCh,
			handlers.DefaultDBOperationTimeout,
		)
	})

	Context("Lock", func() {
		var (
			request      *models.LockRequest
			expectedLock *db.Lock
		)

		BeforeEach(func() {
			request = &models.LockRequest{
				Resource:     resource,
				TtlInSeconds: 10,
			}

			expectedLock = &db.Lock{
				Resource:      resource,
				TtlInSeconds:  10,
				ModifiedIndex: 2,
			}

			fakeLockDB.LockReturns(expectedLock, nil)
		})

		It("reserves the lock in the database", func() {
			_, err := locketHandler.Lock(context.Background(), request)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeLockDB.LockCallCount()).To(Equal(1))

			_, _, actualResource, ttl := fakeLockDB.LockArgsForCall(0)
			Expect(actualResource).To(Equal(resource))
			Expect(ttl).To(BeEquivalentTo(10))

			metricsRecordSuccess(fakeRequestMetrics)
			metricsUseCorrectCallTags(fakeRequestMetrics, "Lock")
		})

		It("increments the in-flight counter and then decrements it when done", func() {
			_, err := locketHandler.Lock(context.Background(), request)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeLockDB.LockCallCount()).To(Equal(1))

			_, delta := fakeRequestMetrics.IncrementRequestsInFlightCounterArgsForCall(0)
			Expect(delta).To(BeEquivalentTo(1))
			_, delta = fakeRequestMetrics.DecrementRequestsInFlightCounterArgsForCall(0)
			Expect(delta).To(BeEquivalentTo(1))
		})

		It("registers the lock and ttl with the lock pick", func() {
			_, err := locketHandler.Lock(context.Background(), request)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLockPick.RegisterTTLCallCount()).To(Equal(1))
			_, lock := fakeLockPick.RegisterTTLArgsForCall(0)
			Expect(lock).To(Equal(expectedLock))

			metricsRecordSuccess(fakeRequestMetrics)
			metricsUseCorrectCallTags(fakeRequestMetrics, "Lock")
		})

		Context("validate lock type", func() {
			Context("when type_code is set", func() {
				It("should be valid on a valid type code and empty type", func() {
					request.Resource.Type = ""
					request.Resource.TypeCode = models.LOCK
					_, err := locketHandler.Lock(context.Background(), request)
					Expect(err).NotTo(HaveOccurred())

					metricsRecordSuccess(fakeRequestMetrics)
					metricsUseCorrectCallTags(fakeRequestMetrics, "Lock")
				})

				It("should be invalid on an UNKNOWN type code and empty type", func() {
					request.Resource.Type = ""
					request.Resource.TypeCode = models.UNKNOWN
					_, err := locketHandler.Lock(context.Background(), request)
					Expect(err).To(HaveOccurred())

					metricsRecordFailure(fakeRequestMetrics)
					metricsUseCorrectCallTags(fakeRequestMetrics, "Lock")
				})

				It("should be invalid on an non-existent type code", func() {
					request.Resource.Type = ""
					request.Resource.TypeCode = 4
					_, err := locketHandler.Lock(context.Background(), request)
					Expect(err).To(HaveOccurred())

					metricsRecordFailure(fakeRequestMetrics)
					metricsUseCorrectCallTags(fakeRequestMetrics, "Lock")
				})
			})
		})

		Context("when request does not have TTL", func() {
			BeforeEach(func() {
				request = &models.LockRequest{
					Resource: resource,
				}
			})

			It("returns a validation error", func() {
				_, err := locketHandler.Lock(context.Background(), request)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrInvalidTTL))
				Expect(logger).To(gbytes.Say(models.ErrInvalidTTL.Error()))
				Expect(logger).To(gbytes.Say("\"key\":"))
				Expect(logger).To(gbytes.Say("\"owner\":"))

				metricsRecordFailure(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "Lock")
			})
		})

		Context("when the request does not have an owner", func() {
			BeforeEach(func() {
				resource.Owner = ""
			})

			It("returns a validation error", func() {
				_, err := locketHandler.Lock(context.Background(), request)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(models.ErrInvalidOwner))
				Expect(logger).To(gbytes.Say(models.ErrInvalidOwner.Error()))
				Expect(logger).To(gbytes.Say("\"key\":"))
				Expect(logger).To(gbytes.Say("\"owner\":"))

				metricsRecordFailure(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "Lock")
			})
		})

		Context("when locking errors", func() {
			var (
				err error
			)

			BeforeEach(func() {
				fakeLockDB.LockReturns(nil, errors.New("Boom."))
			})

			JustBeforeEach(func() {
				_, err = locketHandler.Lock(context.Background(), request)
			})

			It("returns the error", func() {
				Expect(err).To(HaveOccurred())

				metricsRecordFailure(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "Lock")
			})

			It("logs the error with identifying information", func() {
				Expect(logger).To(gbytes.Say("Boom."))
				Expect(logger).To(gbytes.Say("\"key\":"))
				Expect(logger).To(gbytes.Say("\"owner\":"))
			})

			Context("when lock collision error occurs", func() {
				BeforeEach(func() {
					fakeLockDB.LockReturns(nil, models.ErrLockCollision)
				})

				It("does not log the error", func() {
					Expect(logger).NotTo(gbytes.Say("lock-collision"))
				})

				It("counts the request in the metrics as a success", func() {
					metricsRecordSuccess(fakeRequestMetrics)
				})
			})
		})

		Context("when an unrecoverable error is returned", func() {
			BeforeEach(func() {
				fakeLockDB.LockReturns(nil, helpers.ErrUnrecoverableError)
			})

			It("logs and writes to the exit channel", func() {
				locketHandler.Lock(context.Background(), request)
				Expect(logger).To(gbytes.Say("unrecoverable-error"))

				metricsRecordFailure(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "Lock")

				Expect(exitCh).To(Receive())
			})
		})

		Context("when the context errors", func() {
			var ctx context.Context

			BeforeEach(func() {
				ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("uuid", "some-request-id"))
			})

			JustBeforeEach(func() {
				fakeLockDB.LockReturns(nil, errors.New("Boom."))
				locketHandler.Lock(ctx, request)
			})

			Context("when the context was closed due to a client cancellation", func() {
				BeforeEach(func() {
					var cancel func()
					ctx, cancel = context.WithCancel(ctx)
					cancel()
				})

				It("logs the context cancelled error with request data", func() {
					Expect(logger).To(gbytes.Say("context-cancelled"))
					Expect(logger).To(gbytes.Say("some-request-id"))
					Expect(logger).To(gbytes.Say("Lock"))
					Expect(logger).To(gbytes.Say("test"))
					Expect(logger).To(gbytes.Say("myself"))
				})

				It("records an increase in the request cancelled metric", func() {
					Expect(fakeRequestMetrics.IncrementRequestsCancelledCounterCallCount()).To(Equal(1))
					_, requestsCancelled := fakeRequestMetrics.IncrementRequestsCancelledCounterArgsForCall(0)
					Expect(requestsCancelled).To(BeEquivalentTo(1))
				})
			})

			Context("when the context was closed due to an exceeded deadline", func() {
				BeforeEach(func() {
					var cancel context.CancelFunc
					ctx, cancel = context.WithDeadline(ctx, time.Unix(0, 0)) // nolint
					defer cancel()
				})

				It("logs the context deadline exceeded error", func() {
					Expect(logger).To(gbytes.Say("context-deadline-exceeded"))
					Expect(logger).To(gbytes.Say("some-request-id"))
					Expect(logger).To(gbytes.Say("Lock"))
					Expect(logger).To(gbytes.Say("test"))
					Expect(logger).To(gbytes.Say("myself"))
				})

				It("does not emit a requests cancelled metric", func() {
					Expect(fakeRequestMetrics.IncrementRequestsCancelledCounterCallCount()).To(Equal(0))
				})
			})

			Context("when the context has no uuid metadata", func() {
				BeforeEach(func() {
					ctx = context.Background()
					var cancel func()
					ctx, cancel = context.WithCancel(ctx)
					cancel()
				})

				It("does not panic and logs empty request-id", func() {
					Expect(logger).To(gbytes.Say("context-cancelled"))
				})
			})
		})
	})

	Context("Release", func() {
		Context("when the db request take too long", func() {
			var (
				blockDB chan struct{}
			)

			BeforeEach(func() {
				blockDB = make(chan struct{})
				fakeLockDB.ReleaseStub = func(ctx context.Context, logger lager.Logger, resource *models.Resource) error {
					<-blockDB
					return nil
				}
			})

			AfterEach(func() {
				select {
				case <-blockDB:
				default:
					close(blockDB)
				}
			})

			JustBeforeEach(func() {
				go func() {
					_, err := locketHandler.Release(context.Background(), &models.ReleaseRequest{Resource: resource})
					Expect(err).NotTo(HaveOccurred())
				}()
			})

			It("updates the in-flight counter", func() {
				Eventually(fakeRequestMetrics.IncrementRequestsInFlightCounterCallCount).Should(Equal(1))
				requestType, value := fakeRequestMetrics.IncrementRequestsInFlightCounterArgsForCall(0)
				Expect(value).To(BeEquivalentTo(1))
				Expect(requestType).To(Equal("Release"))
			})

			It("does not decrement the counter", func() {
				Consistently(fakeRequestMetrics.DecrementRequestsInFlightCounterCallCount).Should(BeZero())
			})

			Context("when the db request is finished", func() {
				BeforeEach(func() {
					close(blockDB)
				})

				It("decrements the counter", func() {
					Eventually(fakeRequestMetrics.DecrementRequestsInFlightCounterCallCount).Should(Equal(1))
					requestType, value := fakeRequestMetrics.DecrementRequestsInFlightCounterArgsForCall(0)
					Expect(value).To(BeEquivalentTo(1))
					Expect(requestType).To(Equal("Release"))

				})
			})
		})

		It("releases the lock in the database", func() {
			_, err := locketHandler.Release(context.Background(), &models.ReleaseRequest{Resource: resource})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLockDB.ReleaseCallCount()).Should(Equal(1))
			_, _, actualResource := fakeLockDB.ReleaseArgsForCall(0)
			Expect(actualResource).To(Equal(resource))

			metricsRecordSuccess(fakeRequestMetrics)
			metricsUseCorrectCallTags(fakeRequestMetrics, "Release")
		})

		Context("when releasing errors", func() {
			BeforeEach(func() {
				fakeLockDB.ReleaseReturns(errors.New("Boom."))
			})

			It("returns the error", func() {
				_, err := locketHandler.Release(context.Background(), &models.ReleaseRequest{Resource: resource})
				Expect(err).To(HaveOccurred())

				metricsRecordFailure(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "Release")
			})
		})

		Context("when an unrecoverable error is returned", func() {
			BeforeEach(func() {
				fakeLockDB.ReleaseReturns(helpers.ErrUnrecoverableError)
			})

			It("logs and writes to the exit channel", func() {
				locketHandler.Release(context.Background(), &models.ReleaseRequest{Resource: resource})
				Expect(logger).To(gbytes.Say("unrecoverable-error"))

				metricsRecordFailure(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "Release")

				Expect(exitCh).To(Receive())
			})
		})

		Context("when the context errors", func() {
			var ctx context.Context
			BeforeEach(func() {
				ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("uuid", "some-request-id"))
			})

			JustBeforeEach(func() {
				fakeLockDB.ReleaseReturns(errors.New("Boom."))
				locketHandler.Release(ctx, &models.ReleaseRequest{Resource: resource})
			})

			Context("when the context was closed due to a client cancellation", func() {
				BeforeEach(func() {
					var cancel func()
					ctx, cancel = context.WithCancel(ctx)
					cancel()
				})

				It("logs the context cancelled error", func() {
					Expect(logger).To(gbytes.Say("context-cancelled"))
					Expect(logger).To(gbytes.Say("some-request-id"))
					Expect(logger).To(gbytes.Say("Release"))
					Expect(logger).To(gbytes.Say("test"))
					Expect(logger).To(gbytes.Say("myself"))
				})

				It("records an increase in the request cancelled metric", func() {
					Expect(fakeRequestMetrics.IncrementRequestsCancelledCounterCallCount()).To(Equal(1))
					_, requestsCancelled := fakeRequestMetrics.IncrementRequestsCancelledCounterArgsForCall(0)
					Expect(requestsCancelled).To(BeEquivalentTo(1))
				})
			})

			Context("when the context was closed due to an exceeded deadline", func() {
				BeforeEach(func() {
					var cancel context.CancelFunc
					ctx, cancel = context.WithDeadline(ctx, time.Unix(0, 0)) // nolint
					defer cancel()
				})

				It("logs the context deadline exceeded error", func() {
					Expect(logger).To(gbytes.Say("context-deadline-exceeded"))
					Expect(logger).To(gbytes.Say("some-request-id"))
					Expect(logger).To(gbytes.Say("Release"))
					Expect(logger).To(gbytes.Say("test"))
					Expect(logger).To(gbytes.Say("myself"))
				})

				It("does not emit a requests cancelled metric", func() {
					Expect(fakeRequestMetrics.IncrementRequestsCancelledCounterCallCount()).To(Equal(0))
				})
			})
		})
	})

	Context("Fetch", func() {
		BeforeEach(func() {
			fakeLockDB.FetchReturns(&db.Lock{Resource: resource}, nil)
		})

		It("fetches the lock in the database", func() {
			fetchResp, err := locketHandler.Fetch(context.Background(), &models.FetchRequest{Key: "test-fetch"})
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchResp.Resource).To(Equal(resource))

			Expect(fakeLockDB.FetchCallCount()).Should(Equal(1))
			_, _, key := fakeLockDB.FetchArgsForCall(0)
			Expect(key).To(Equal("test-fetch"))

			metricsRecordSuccess(fakeRequestMetrics)
			metricsUseCorrectCallTags(fakeRequestMetrics, "Fetch")
		})

		Context("when fetching errors", func() {
			BeforeEach(func() {
				fakeLockDB.FetchReturns(nil, errors.New("boom"))
			})

			It("returns the error", func() {
				_, err := locketHandler.Fetch(context.Background(), &models.FetchRequest{Key: "test-fetch"})
				Expect(err).To(HaveOccurred())

				metricsRecordFailure(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "Fetch")
			})
		})

		Context("when an unrecoverable error is returned", func() {
			BeforeEach(func() {
				fakeLockDB.FetchReturns(nil, helpers.ErrUnrecoverableError)
			})

			It("logs and writes to the exit channel", func() {
				locketHandler.Fetch(context.Background(), &models.FetchRequest{Key: "test-fetch"})
				Expect(logger).To(gbytes.Say("unrecoverable-error"))

				metricsRecordFailure(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "Fetch")

				Expect(exitCh).To(Receive())
			})
		})

		Context("when the context errors", func() {
			var ctx context.Context
			BeforeEach(func() {
				ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("uuid", "some-request-id"))
			})

			JustBeforeEach(func() {
				fakeLockDB.FetchReturns(nil, errors.New("boom"))
				locketHandler.Fetch(ctx, &models.FetchRequest{Key: "test-fetch"})
			})

			Context("when the context was closed due to a client cancellation", func() {
				BeforeEach(func() {
					var cancel func()
					ctx, cancel = context.WithCancel(ctx)
					cancel()
				})

				It("logs the context cancelled error", func() {
					Expect(logger).To(gbytes.Say("context-cancelled"))
					Expect(logger).To(gbytes.Say("some-request-id"))
					Expect(logger).To(gbytes.Say("Fetch"))
					Expect(logger).To(gbytes.Say("test"))
				})

				It("records an increase in the request cancelled metric", func() {
					Expect(fakeRequestMetrics.IncrementRequestsCancelledCounterCallCount()).To(Equal(1))
					_, requestsCancelled := fakeRequestMetrics.IncrementRequestsCancelledCounterArgsForCall(0)
					Expect(requestsCancelled).To(BeEquivalentTo(1))
				})
			})

			Context("when the context was closed due to an exceeded deadline", func() {
				BeforeEach(func() {
					var cancel context.CancelFunc
					ctx, cancel = context.WithDeadline(ctx, time.Unix(0, 0)) // nolint
					defer cancel()
				})

				It("logs the context deadline exceeded error", func() {
					Expect(logger).To(gbytes.Say("context-deadline-exceeded"))
					Expect(logger).To(gbytes.Say("some-request-id"))
					Expect(logger).To(gbytes.Say("Fetch"))
					Expect(logger).To(gbytes.Say("test"))
				})

				It("does not emit a requests cancelled metric", func() {
					Expect(fakeRequestMetrics.IncrementRequestsCancelledCounterCallCount()).To(Equal(0))
				})
			})
		})
	})

	Context("FetchAll", func() {
		var expectedResources []*models.Resource
		BeforeEach(func() {
			expectedResources = []*models.Resource{
				resource,
				&models.Resource{Key: "cell", Owner: "cell-1", Value: "{}"},
			}

			var locks []*db.Lock
			for _, r := range expectedResources {
				locks = append(locks, &db.Lock{Resource: r})
			}
			fakeLockDB.FetchAllReturns(locks, nil)
		})

		Context("validate lock type", func() {
			Context("when type_code is set", func() {
				It("should be valid on a valid type code and empty type", func() {
					_, err := locketHandler.FetchAll(context.Background(), &models.FetchAllRequest{TypeCode: models.LOCK})
					Expect(err).NotTo(HaveOccurred())

					metricsRecordSuccess(fakeRequestMetrics)
					metricsUseCorrectCallTags(fakeRequestMetrics, "FetchAll")

					_, err = locketHandler.FetchAll(context.Background(), &models.FetchAllRequest{TypeCode: models.PRESENCE})
					Expect(err).NotTo(HaveOccurred())
				})

				It("should be invalid on an UNKNOWN type code and empty type", func() {
					_, err := locketHandler.FetchAll(context.Background(), &models.FetchAllRequest{})
					Expect(err).To(HaveOccurred())

					metricsRecordFailure(fakeRequestMetrics)
					metricsUseCorrectCallTags(fakeRequestMetrics, "FetchAll")
				})
			})
		})

		Context("when the type is valid", func() {
			It("fetches all the presence locks in the database by type", func() {
				fetchResp, err := locketHandler.FetchAll(context.Background(), &models.FetchAllRequest{TypeCode: models.PRESENCE})
				Expect(err).NotTo(HaveOccurred())

				metricsRecordSuccess(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "FetchAll")

				Expect(fetchResp.Resources).To(Equal(expectedResources))
				Expect(fakeLockDB.FetchAllCallCount()).Should(Equal(1))
				_, _, lockType := fakeLockDB.FetchAllArgsForCall(0)
				Expect(lockType).To(Equal("presence"))
			})

			It("fetches all the lock locks in the database by type", func() {
				fetchResp, err := locketHandler.FetchAll(context.Background(), &models.FetchAllRequest{TypeCode: models.LOCK})
				Expect(err).NotTo(HaveOccurred())

				metricsRecordSuccess(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "FetchAll")

				Expect(fetchResp.Resources).To(Equal(expectedResources))
				Expect(fakeLockDB.FetchAllCallCount()).Should(Equal(1))
				_, _, lockType := fakeLockDB.FetchAllArgsForCall(0)
				Expect(lockType).To(Equal("lock"))
			})

			It("fetches all the presence locks in the database by type code", func() {
				fetchResp, err := locketHandler.FetchAll(context.Background(), &models.FetchAllRequest{TypeCode: models.PRESENCE})
				Expect(err).NotTo(HaveOccurred())

				metricsRecordSuccess(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "FetchAll")

				Expect(fetchResp.Resources).To(Equal(expectedResources))
				Expect(fakeLockDB.FetchAllCallCount()).Should(Equal(1))
				_, _, lockType := fakeLockDB.FetchAllArgsForCall(0)
				Expect(lockType).To(Equal("presence"))
			})

			It("fetches all the lock locks in the database by type code", func() {
				fetchResp, err := locketHandler.FetchAll(context.Background(), &models.FetchAllRequest{TypeCode: models.LOCK})
				Expect(err).NotTo(HaveOccurred())

				metricsRecordSuccess(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "FetchAll")

				Expect(fetchResp.Resources).To(Equal(expectedResources))
				Expect(fakeLockDB.FetchAllCallCount()).Should(Equal(1))
				_, _, lockType := fakeLockDB.FetchAllArgsForCall(0)
				Expect(lockType).To(Equal("lock"))
			})
		})

		Context("when the type is invalid", func() {
			It("returns an invalid type error", func() {
				_, err := locketHandler.FetchAll(context.Background(), &models.FetchAllRequest{Type: "dawg"})
				Expect(err).To(HaveOccurred())

				metricsRecordFailure(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "FetchAll")
			})
		})

		Context("when the type code is UNKNOWN", func() {
			It("returns an invalid type error", func() {
				_, err := locketHandler.FetchAll(context.Background(), &models.FetchAllRequest{TypeCode: models.UNKNOWN})
				Expect(err).To(HaveOccurred())

				metricsRecordFailure(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "FetchAll")
			})
		})

		Context("when fetching errors", func() {
			BeforeEach(func() {
				fakeLockDB.FetchAllReturns(nil, errors.New("boom"))
			})

			It("returns the error", func() {
				_, err := locketHandler.FetchAll(context.Background(), &models.FetchAllRequest{})
				Expect(err).To(HaveOccurred())

				metricsRecordFailure(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "FetchAll")
			})
		})

		Context("when an unrecoverable error is returned", func() {
			BeforeEach(func() {
				fakeLockDB.FetchAllReturns(nil, helpers.ErrUnrecoverableError)
			})

			It("logs and writes to the exit channel", func() {
				locketHandler.FetchAll(context.Background(), &models.FetchAllRequest{TypeCode: models.PRESENCE})
				Expect(logger).To(gbytes.Say("unrecoverable-error"))

				metricsRecordFailure(fakeRequestMetrics)
				metricsUseCorrectCallTags(fakeRequestMetrics, "FetchAll")

				Expect(exitCh).To(Receive())
			})
		})

		Context("when the context errors", func() {
			var ctx context.Context
			BeforeEach(func() {
				ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("uuid", "some-request-id"))
			})

			JustBeforeEach(func() {
				fakeLockDB.FetchAllReturns(nil, errors.New("boom"))
				locketHandler.FetchAll(ctx, &models.FetchAllRequest{})
			})

			Context("when the context was closed due to a client cancellation", func() {
				BeforeEach(func() {
					var cancel func()
					ctx, cancel = context.WithCancel(ctx)
					cancel()
				})

				It("logs the context cancelled error", func() {
					Expect(logger).To(gbytes.Say("context-cancelled"))
					Expect(logger).To(gbytes.Say("some-request-id"))
					Expect(logger).To(gbytes.Say("FetchAll"))
				})

				It("records an increase in the request cancelled metric", func() {
					Expect(fakeRequestMetrics.IncrementRequestsCancelledCounterCallCount()).To(Equal(1))
					_, requestsCancelled := fakeRequestMetrics.IncrementRequestsCancelledCounterArgsForCall(0)
					Expect(requestsCancelled).To(BeEquivalentTo(1))
				})
			})

			Context("when the context was closed due to an exceeded deadline", func() {
				BeforeEach(func() {
					var cancel context.CancelFunc
					ctx, cancel = context.WithDeadline(ctx, time.Unix(0, 0)) // nolint
					defer cancel()
				})

				It("logs the context deadline exceeded error", func() {
					Expect(logger).To(gbytes.Say("context-deadline-exceeded"))
					Expect(logger).To(gbytes.Say("some-request-id"))
					Expect(logger).To(gbytes.Say("FetchAll"))
				})

				It("does not emit a requests cancelled metric", func() {
					Expect(fakeRequestMetrics.IncrementRequestsCancelledCounterCallCount()).To(Equal(0))
				})
			})
		})
	})

	Context("DB context isolation", func() {
		var (
			blockDB chan struct{}
		)

		BeforeEach(func() {
			blockDB = make(chan struct{})
		})

		AfterEach(func() {
			select {
			case <-blockDB:
			default:
				close(blockDB)
			}
		})

		verifyDBContextIsolation := func(
			waitForDBCall func(),
			callHandler func(ctx context.Context),
			getDBContext func() context.Context,
		) {
			grpcCtx, grpcCancel := context.WithCancel(context.Background())

			done := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				callHandler(grpcCtx)
				close(done)
			}()

			waitForDBCall()
			grpcCancel()

			Consistently(done, 200*time.Millisecond).ShouldNot(BeClosed())
			Expect(getDBContext().Err()).To(BeNil())

			close(blockDB)
			Eventually(done).Should(BeClosed())
		}

		It("Lock: does not cancel the DB operation when the gRPC context is cancelled", func() {
			fakeLockDB.LockStub = func(ctx context.Context, logger lager.Logger, resource *models.Resource, ttl int64) (*db.Lock, error) {
				<-blockDB
				return &db.Lock{Resource: resource, TtlInSeconds: ttl, ModifiedIndex: 1}, nil
			}
			verifyDBContextIsolation(
				func() { Eventually(fakeLockDB.LockCallCount).Should(Equal(1)) },
				func(ctx context.Context) {
					_, err := locketHandler.Lock(ctx, &models.LockRequest{Resource: resource, TtlInSeconds: 10})
					Expect(err).NotTo(HaveOccurred())
				},
				func() context.Context { ctx, _, _, _ := fakeLockDB.LockArgsForCall(0); return ctx },
			)
		})

		It("Release: does not cancel the DB operation when the gRPC context is cancelled", func() {
			fakeLockDB.ReleaseStub = func(ctx context.Context, logger lager.Logger, resource *models.Resource) error {
				<-blockDB
				return nil
			}
			verifyDBContextIsolation(
				func() { Eventually(fakeLockDB.ReleaseCallCount).Should(Equal(1)) },
				func(ctx context.Context) {
					_, err := locketHandler.Release(ctx, &models.ReleaseRequest{Resource: resource})
					Expect(err).NotTo(HaveOccurred())
				},
				func() context.Context { ctx, _, _ := fakeLockDB.ReleaseArgsForCall(0); return ctx },
			)
		})

		It("Fetch: does not cancel the DB operation when the gRPC context is cancelled", func() {
			fakeLockDB.FetchStub = func(ctx context.Context, logger lager.Logger, key string) (*db.Lock, error) {
				<-blockDB
				return &db.Lock{Resource: resource}, nil
			}
			verifyDBContextIsolation(
				func() { Eventually(fakeLockDB.FetchCallCount).Should(Equal(1)) },
				func(ctx context.Context) {
					_, err := locketHandler.Fetch(ctx, &models.FetchRequest{Key: "test"})
					Expect(err).NotTo(HaveOccurred())
				},
				func() context.Context { ctx, _, _ := fakeLockDB.FetchArgsForCall(0); return ctx },
			)
		})

		It("FetchAll: does not cancel the DB operation when the gRPC context is cancelled", func() {
			fakeLockDB.FetchAllStub = func(ctx context.Context, logger lager.Logger, lockType string) ([]*db.Lock, error) {
				<-blockDB
				return []*db.Lock{{Resource: resource}}, nil
			}
			verifyDBContextIsolation(
				func() { Eventually(fakeLockDB.FetchAllCallCount).Should(Equal(1)) },
				func(ctx context.Context) {
					_, err := locketHandler.FetchAll(ctx, &models.FetchAllRequest{TypeCode: models.LOCK})
					Expect(err).NotTo(HaveOccurred())
				},
				func() context.Context { ctx, _, _ := fakeLockDB.FetchAllArgsForCall(0); return ctx },
			)
		})

		It("uses a context with the configured timeout for DB operations", func() {
			shortTimeout := 100 * time.Millisecond
			shortTimeoutHandler := handlers.NewLocketHandler(
				logger,
				fakeLockDB,
				fakeLockPick,
				fakeRequestMetrics,
				exitCh,
				shortTimeout,
			)

			fakeLockDB.LockStub = func(ctx context.Context, logger lager.Logger, resource *models.Resource, ttl int64) (*db.Lock, error) {
				deadline, ok := ctx.Deadline()
				Expect(ok).To(BeTrue())
				Expect(time.Until(deadline)).To(BeNumerically("<=", shortTimeout))
				return &db.Lock{Resource: resource, TtlInSeconds: ttl, ModifiedIndex: 1}, nil
			}

			_, err := shortTimeoutHandler.Lock(context.Background(), &models.LockRequest{
				Resource:     resource,
				TtlInSeconds: 10,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func metricsRecordSuccess(fakeRequestMetrics *helpersfakes.FakeRequestMetrics) {
	Expect(fakeRequestMetrics.IncrementRequestsStartedCounterCallCount()).To(Equal(1))
	_, started := fakeRequestMetrics.IncrementRequestsStartedCounterArgsForCall(0)
	Expect(started).To(BeEquivalentTo(1))

	Expect(fakeRequestMetrics.IncrementRequestsSucceededCounterCallCount()).To(Equal(1))
	_, succeeded := fakeRequestMetrics.IncrementRequestsSucceededCounterArgsForCall(0)
	Expect(succeeded).To(BeEquivalentTo(1))

	Expect(fakeRequestMetrics.IncrementRequestsFailedCounterCallCount()).To(Equal(0))

	Expect(fakeRequestMetrics.IncrementRequestsInFlightCounterCallCount()).To(Equal(1))
	_, incInFlight := fakeRequestMetrics.IncrementRequestsInFlightCounterArgsForCall(0)
	Expect(incInFlight).To(BeEquivalentTo(1))

	Expect(fakeRequestMetrics.DecrementRequestsInFlightCounterCallCount()).To(Equal(1))
	_, decInFlight := fakeRequestMetrics.IncrementRequestsInFlightCounterArgsForCall(0)
	Expect(decInFlight).To(BeEquivalentTo(1))

	Expect(fakeRequestMetrics.UpdateLatencyCallCount()).To(Equal(1))
	_, latency := fakeRequestMetrics.UpdateLatencyArgsForCall(0)
	Expect(latency).To(BeNumerically(">=", 0))
}

func metricsRecordFailure(fakeRequestMetrics *helpersfakes.FakeRequestMetrics) {
	Expect(fakeRequestMetrics.IncrementRequestsStartedCounterCallCount()).To(Equal(1))
	_, started := fakeRequestMetrics.IncrementRequestsStartedCounterArgsForCall(0)
	Expect(started).To(BeEquivalentTo(1))

	Expect(fakeRequestMetrics.IncrementRequestsSucceededCounterCallCount()).To(Equal(0))

	Expect(fakeRequestMetrics.IncrementRequestsFailedCounterCallCount()).To(Equal(1))
	_, failed := fakeRequestMetrics.IncrementRequestsFailedCounterArgsForCall(0)
	Expect(failed).To(BeEquivalentTo(1))

	Expect(fakeRequestMetrics.IncrementRequestsInFlightCounterCallCount()).To(Equal(1))
	_, incInFlight := fakeRequestMetrics.IncrementRequestsInFlightCounterArgsForCall(0)
	Expect(incInFlight).To(BeEquivalentTo(1))

	Expect(fakeRequestMetrics.DecrementRequestsInFlightCounterCallCount()).To(Equal(1))
	_, decInFlight := fakeRequestMetrics.IncrementRequestsInFlightCounterArgsForCall(0)
	Expect(decInFlight).To(BeEquivalentTo(1))

	Expect(fakeRequestMetrics.UpdateLatencyCallCount()).To(Equal(1))
	_, latency := fakeRequestMetrics.UpdateLatencyArgsForCall(0)
	Expect(latency).To(BeNumerically(">=", 0))
}

func metricsUseCorrectCallTags(fakeRequestMetrics *helpersfakes.FakeRequestMetrics, expectedRequestType string) {

	if fakeRequestMetrics.IncrementRequestsStartedCounterCallCount() > 0 {
		requestType, _ := fakeRequestMetrics.IncrementRequestsStartedCounterArgsForCall(0)
		Expect(requestType).To(Equal(expectedRequestType))
	}

	if fakeRequestMetrics.IncrementRequestsSucceededCounterCallCount() > 0 {
		requestType, _ := fakeRequestMetrics.IncrementRequestsSucceededCounterArgsForCall(0)
		Expect(requestType).To(Equal(expectedRequestType))
	}

	if fakeRequestMetrics.IncrementRequestsFailedCounterCallCount() > 0 {
		requestType, _ := fakeRequestMetrics.IncrementRequestsFailedCounterArgsForCall(0)
		Expect(requestType).To(Equal(expectedRequestType))
	}

	if fakeRequestMetrics.IncrementRequestsInFlightCounterCallCount() > 0 {
		requestType, _ := fakeRequestMetrics.IncrementRequestsInFlightCounterArgsForCall(0)
		Expect(requestType).To(Equal(expectedRequestType))
	}

	if fakeRequestMetrics.DecrementRequestsInFlightCounterCallCount() > 0 {
		requestType, _ := fakeRequestMetrics.DecrementRequestsInFlightCounterArgsForCall(0)
		Expect(requestType).To(Equal(expectedRequestType))
	}

	if fakeRequestMetrics.UpdateLatencyCallCount() > 0 {
		requestType, _ := fakeRequestMetrics.UpdateLatencyArgsForCall(0)
		Expect(requestType).To(Equal(expectedRequestType))
	}
}
