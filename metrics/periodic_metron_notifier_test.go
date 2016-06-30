package metrics_test

import (
	"net/http"
	"os"
	"time"

	"code.cloudfoundry.org/bbs/db/etcd"
	"code.cloudfoundry.org/bbs/metrics"
	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/cloudfoundry/dropsonde/metric_sender/fake"
	dropsonde_metrics "github.com/cloudfoundry/dropsonde/metrics"
	"github.com/tedsuo/ifrit"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

// a bit of grace time for eventuallys
const aBit = 50 * time.Millisecond

var _ = Describe("PeriodicMetronNotifier", func() {
	var (
		sender *fake.FakeMetricSender

		etcdOptions    etcd.ETCDOptions
		reportInterval time.Duration
		fakeClock      *fakeclock.FakeClock

		pmn ifrit.Process
	)

	BeforeEach(func() {
		reportInterval = 100 * time.Millisecond

		fakeClock = fakeclock.NewFakeClock(time.Unix(123, 456))

		sender = fake.NewFakeMetricSender()
		dropsonde_metrics.Initialize(sender, nil)
		etcdOptions.IsConfigured = true
	})

	JustBeforeEach(func() {
		pmn = ifrit.Invoke(metrics.NewPeriodicMetronNotifier(
			lagertest.NewTestLogger("test"),
			reportInterval,
			&etcdOptions,
			fakeClock,
		))
	})

	AfterEach(func() {
		pmn.Signal(os.Interrupt)
		Eventually(pmn.Wait(), 2*time.Second).Should(Receive())
	})

	Context("when the metron notifier starts up", func() {
		It("should emit an event that BBS has started", func() {
			Eventually(func() uint64 {
				return sender.GetCounter("BBSMasterElected")
			}).Should(Equal(uint64(1)))
		})
	})

	Context("when the report interval elapses", func() {
		var (
			etcd1 *ghttp.Server
			etcd2 *ghttp.Server
			etcd3 *ghttp.Server
		)

		BeforeEach(func() {
			etcd1 = ghttp.NewServer()
			etcd2 = ghttp.NewServer()
			etcd3 = ghttp.NewServer()

			etcd1.RouteToHandler("GET", "/v2/stats/self", ghttp.RespondWith(200, `
            {
              "name": "node1",
							"id": "node1-id",
              "state": "StateFollower",

              "leaderInfo": {
                "leader": "node2-id",
								"uptime": "17h41m45.103057785s",
							  "startTime": "2015-02-13T01:28:26.657389108Z"
              },

              "recvAppendRequestCnt": 1234,
              "recvPkgRate": 2.0,
              "recvBandwidthRate": 1.2,

              "sendAppendRequestCnt": 4321
            }
	        `))

			etcd1.RouteToHandler("GET", "/v2/stats/leader", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, etcd2.URL(), 302)
			})

			etcd2.RouteToHandler("GET", "/v2/stats/self", ghttp.RespondWith(200, `
            {
              "name": "node2",
							"id": "node2-id",
              "state": "StateLeader",

              "leaderInfo": {
                "leader": "node2-id",
								"uptime": "17h41m45.103057785s",
							  "startTime": "2015-02-13T01:28:26.657389108Z"
              },

              "recvAppendRequestCnt": 1234,

              "sendAppendRequestCnt": 4321,
              "sendPkgRate": 5.0,
              "sendBandwidthRate": 3.0
            }
	        `))

			etcd2.RouteToHandler("GET", "/v2/stats/leader", ghttp.RespondWith(200, `
						{
						  "leader": "node2-id",
						  "followers": {
						    "node1-id": {
						      "latency": {
						        "current": 0.153507,
						        "average": 0.14636559394884047,
						        "standardDeviation": 0.15477392607571758,
						        "minimum": 8.4e-05,
						        "maximum": 6.78157
						      },
						      "counts": {
						        "fail": 4,
						        "success": 215000
						      }
						    },
						    "node3-id": {
						      "latency": {
						        "current": 0.052932,
						        "average": 0.13533593782359846,
						        "standardDeviation": 0.18151611603344037,
						        "minimum": 7.3e-05,
						        "maximum": 16.432439
						      },
						      "counts": {
						        "fail": 4,
						        "success": 214969
						      }
						    }
						  }
						}
	        `))

			etcd2.RouteToHandler("GET", "/v2/stats/store", ghttp.RespondWith(200, `
						{
							"getsSuccess": 10195,
							"getsFail": 26705,
							"setsSuccess": 2540,
							"setsFail": 0,
							"deleteSuccess": 0,
							"deleteFail": 0,
							"updateSuccess": 0,
							"updateFail": 0,
							"createSuccess": 18,
							"createFail": 15252,
							"compareAndSwapSuccess": 50350,
							"compareAndSwapFail": 22,
							"compareAndDeleteSuccess": 4,
							"compareAndDeleteFail": 0,
							"expireCount": 1,
							"watchers": 12
						}
					`))

			etcd2.RouteToHandler("GET", "/v2/keys", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Raft-Term", "123")
			})

			etcd3.RouteToHandler("GET", "/v2/stats/self", ghttp.RespondWith(200, `
            {
              "name": "node3",
							"id": "node3-id",
              "state": "StateFollower",

              "leaderInfo": {
                "leader": "node2-id",
								"uptime": "17h41m45.103057785s",
							  "startTime": "2015-02-13T01:28:26.657389108Z"
              },

              "recvAppendRequestCnt": 1234,
              "recvPkgRate": 2.0,
              "recvBandwidthRate": 0.8,

              "sendAppendRequestCnt": 4321
            }
	        `))

			etcd3.RouteToHandler("GET", "/v2/stats/leader", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, etcd2.URL(), 302)
			})

			etcdOptions.ClusterUrls = []string{
				etcd1.URL(),
				etcd2.URL(),
				etcd3.URL(),
			}
		})

		JustBeforeEach(func() {
			fakeClock.Increment(reportInterval)
		})

		AfterEach(func() {
			etcd1.Close()
			etcd2.Close()
			etcd3.Close()
		})

		Context("when the etcd cluster is around", func() {
			It("should emit them", func() {
				Eventually(func() fake.Metric {
					return sender.GetValue("ETCDLeader")
				}).Should(Equal(fake.Metric{
					Value: 1,
					Unit:  "Metric",
				}))

				Eventually(func() fake.Metric {
					return sender.GetValue("ETCDReceivedBandwidthRate")
				}).Should(Equal(fake.Metric{
					Value: 2,
					Unit:  "B/s",
				}))

				Eventually(func() fake.Metric {
					return sender.GetValue("ETCDSentBandwidthRate")
				}).Should(Equal(fake.Metric{
					Value: 3,
					Unit:  "B/s",
				}))

				Eventually(func() fake.Metric {
					return sender.GetValue("ETCDReceivedRequestRate")
				}).Should(Equal(fake.Metric{
					Value: 4,
					Unit:  "Req/s",
				}))

				Eventually(func() fake.Metric {
					return sender.GetValue("ETCDSentRequestRate")
				}).Should(Equal(fake.Metric{
					Value: 5,
					Unit:  "Req/s",
				}))

				Eventually(func() fake.Metric {
					return sender.GetValue("ETCDRaftTerm")
				}).Should(Equal(fake.Metric{
					Value: 123,
					Unit:  "Metric",
				}))

				Eventually(func() fake.Metric {
					return sender.GetValue("ETCDWatchers")
				}).Should(Equal(fake.Metric{
					Value: 12,
					Unit:  "Metric",
				}))
			})
		})

		Context("when the etcd cluster is not around", func() {
			BeforeEach(func() {
				etcdOptions.IsConfigured = false
			})

			It("does not emit any metrics", func() {
				Consistently(func() fake.Metric {
					return sender.GetValue("ETCDLeader")
				}).Should(Equal(fake.Metric{}))
			})
		})
	})
})
