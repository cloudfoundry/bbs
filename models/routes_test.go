package models_test

import (
	"encoding/json"

	"code.cloudfoundry.org/bbs/models"
	"google.golang.org/protobuf/proto"
	protocmp "google.golang.org/protobuf/testing/protocmp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// var _ = Describe("Routes", func() {
// 	var update models.DesiredLRPUpdate
// 	var aJson models.DesiredLRPUpdate
// 	var aProto models.DesiredLRPUpdate

// 	itSerializes := func(routes *models.Routes) {
// 		BeforeEach(func() {
// 			update = models.DesiredLRPUpdate{
// 				Routes: routes,
// 			}

// 			b, err := json.Marshal(update)
// 			Expect(err).NotTo(HaveOccurred())
// 			err = json.Unmarshal(b, &aJson)
// 			Expect(err).NotTo(HaveOccurred())

// 			b, err = proto.Marshal(&update)
// 			Expect(err).NotTo(HaveOccurred())
// 			err = proto.Unmarshal(b, &aProto)
// 			Expect(err).NotTo(HaveOccurred())
// 		})

// 		It("marshals JSON properly", func() {
// 			Expect(update.Equal(&aJson)).To(BeTrue())
// 			Expect(update).To(Equal(aJson))
// 		})

// 		It("marshals Proto properly", func() {
// 			Expect(update.Equal(&aProto)).To(BeTrue())
// 			Expect(update).To(Equal(aProto))
// 		})
// 	}

// 	itSerializes(nil)
// 	itSerializes(&models.Routes{
// 		"abc": &(json.RawMessage{'"', 'd', '"'}),
// 		"def": &(json.RawMessage{'"', 'g', '"'}),
// 	})
// })

var _ = Describe("ProtoRoutes", func() {
	// var beforeProtoRoutes models.ProtoRoutes
	// var beforeRoutes models.Routes

	// BeforeEach(func() {
	// })

	It("marshals properly", func() {
		prStruct := models.ProtoRoutes{
			Routes: map[string][]byte{"cf-router": []byte(`[
        {
          "hostnames": [
            "some-route.example.com"
          ],
          "port": 8080
        }
	  ]`),
				"diego-ssh": []byte(`{
					"container_port": 2222,
					"host_fingerprint": "ac:99:67:20:7e:c2:7c:2c:d2:22:37:bc:9f:14:01:ec",
					"private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQDuOfcUnfiXE6g6Cvgur3Om6t8cEx27FAoVrDrxMzy+q2NTJaQF\nNYqG2DDDHZCLG2mJasryKZfDyK30c48ITpecBkCux429aZN2gEJCEsyYgsZheI+5\neNYs1vzl68KQ1LdxlgNOqFZijyVjTOD60GMPCVlDICqGNUFH4aPTHA0fVwIDAQAB\nAoGBAO1Ak19YGHy1mgP8asFsAT1KitrV+vUW9xgwiB8xjRzDac8kHJ8HfKfg5Wdc\nqViw+0FdNzNH0xqsYPqkn92BECDqdWOzhlEYNj/AFSHTdRPrs9w82b7h/LhrX0H/\nRUrU2QrcI2uSV/SQfQvFwC6YaYugCo35noljJEcD8EYQTcRxAkEA+jfjumM6da8O\n8u8Rc58Tih1C5mumeIfJMPKRz3FBLQEylyMWtGlr1XT6ppqiHkAAkQRUBgKi+Ffi\nYedQOvE0/wJBAPO7I+brmrknzOGtSK2tvVKnMqBY6F8cqmG4ZUm0W9tMLKiR7JWO\nAsjSlQfEEnpOr/AmuONwTsNg+g93IILv3akCQQDnrKfmA8o0/IlS1ZfK/hcRYlZ3\nEmVoZBEciPwInkxCZ0F4Prze/l0hntYVPEeuyoO7wc4qYnaSiozJKWtXp83xAkBo\nk+ubsYv51jH6wzdkDiAlzsfSNVO/O7V/qHcNYO3o8o5W5gX1RbG8KV74rhCfmhOz\nn2nFbPLeskWZTSwOAo3BAkBWHBjvCj1sBgsIG4v6Tn2ig21akbmssJezmZRjiqeh\nqt0sAzMVixAwIFM0GsW3vQ8Hr/eBTb5EBQVZ/doRqUzf\n-----END RSA PRIVATE KEY-----\n"
				}`),
			},
		}

		jsonResult := []byte(`{
      "cf-router": [
        {
          "hostnames": [
            "some-route.example.com"
          ],
          "port": 8080
        }
      ],
      "diego-ssh": {
        "container_port": 2222,
        "host_fingerprint": "ac:99:67:20:7e:c2:7c:2c:d2:22:37:bc:9f:14:01:ec",
        "private_key": "-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQDuOfcUnfiXE6g6Cvgur3Om6t8cEx27FAoVrDrxMzy+q2NTJaQF\nNYqG2DDDHZCLG2mJasryKZfDyK30c48ITpecBkCux429aZN2gEJCEsyYgsZheI+5\neNYs1vzl68KQ1LdxlgNOqFZijyVjTOD60GMPCVlDICqGNUFH4aPTHA0fVwIDAQAB\nAoGBAO1Ak19YGHy1mgP8asFsAT1KitrV+vUW9xgwiB8xjRzDac8kHJ8HfKfg5Wdc\nqViw+0FdNzNH0xqsYPqkn92BECDqdWOzhlEYNj/AFSHTdRPrs9w82b7h/LhrX0H/\nRUrU2QrcI2uSV/SQfQvFwC6YaYugCo35noljJEcD8EYQTcRxAkEA+jfjumM6da8O\n8u8Rc58Tih1C5mumeIfJMPKRz3FBLQEylyMWtGlr1XT6ppqiHkAAkQRUBgKi+Ffi\nYedQOvE0/wJBAPO7I+brmrknzOGtSK2tvVKnMqBY6F8cqmG4ZUm0W9tMLKiR7JWO\nAsjSlQfEEnpOr/AmuONwTsNg+g93IILv3akCQQDnrKfmA8o0/IlS1ZfK/hcRYlZ3\nEmVoZBEciPwInkxCZ0F4Prze/l0hntYVPEeuyoO7wc4qYnaSiozJKWtXp83xAkBo\nk+ubsYv51jH6wzdkDiAlzsfSNVO/O7V/qHcNYO3o8o5W5gX1RbG8KV74rhCfmhOz\nn2nFbPLeskWZTSwOAo3BAkBWHBjvCj1sBgsIG4v6Tn2ig21akbmssJezmZRjiqeh\nqt0sAzMVixAwIFM0GsW3vQ8Hr/eBTb5EBQVZ/doRqUzf\n-----END RSA PRIVATE KEY-----\n"
      }
    }`)
		prJson, err := json.Marshal(&prStruct)
		Expect(err).NotTo(HaveOccurred())

		Expect(prJson).To(MatchJSON(jsonResult))

		// I don't know if we need any of the following, but it does work
		b, err := proto.Marshal(&prStruct)
		Expect(err).NotTo(HaveOccurred())

		remarshaledProto := models.ProtoRoutes{}
		err = proto.Unmarshal(b, &remarshaledProto)
		Expect(err).NotTo(HaveOccurred())
		Expect(remarshaledProto).To(BeComparableTo(prStruct, protocmp.Transform()))
	})

	It("unmarshals properly", func() {
		protoRoutes := models.ProtoRoutes{}
		jsonSource := `{
			"cf-router": [{"hostnames":["some-route.example.com"],"port":8080}],
			"diego-ssh": {"container_port":2222,"host_fingerprint":"ac:99","private_key":"--RSA PRIVATE KEY--"}
		}`

		err := json.Unmarshal([]byte(jsonSource), &protoRoutes)
		Expect(err).NotTo(HaveOccurred())

		expectedProtoRoutes := models.ProtoRoutes{
			Routes: map[string][]byte{
				"cf-router": []byte(`[{"hostnames":["some-route.example.com"],"port":8080}]`),
				"diego-ssh": []byte(`{"container_port":2222,"host_fingerprint":"ac:99","private_key":"--RSA PRIVATE KEY--"}`),
			},
		}
		Expect(protoRoutes.Routes).To(Equal(expectedProtoRoutes.Routes))

		// for k, v := range expectedProtoRoutes.Routes {
		// 	Expect(protoRoutes.Routes[k]).NotTo(BeNil())
		// 	Expect(json.RawMessage(protoRoutes.Routes[k])).To(MatchJSON(v))
		// }
	})
})
