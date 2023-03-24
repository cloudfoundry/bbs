package models_test

import (
	"encoding/json"

	"code.cloudfoundry.org/bbs/models"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SecurityGroupRule", func() {
	var rule models.SecurityGroupRule

	BeforeEach(func() {
		rule = models.SecurityGroupRule{
			Protocol:     models.TCPProtocol,
			Destinations: []string{"1.2.3.4/16"},
			PortRange: &models.PortRange{
				Start: 1,
				End:   1024,
			},
			Log: false,
		}
	})

	Describe("Validation", func() {
		var (
			validationErr error

			protocol    string
			destination string

			ports     []uint32
			portRange *models.PortRange

			icmpInfo *models.ICMPInfo

			log bool
		)

		BeforeEach(func() {
			protocol = "tcp"
			destination = "8.8.8.8/16"

			ports = nil
			portRange = nil
			icmpInfo = nil
			log = false
		})

		JustBeforeEach(func() {
			rule = models.SecurityGroupRule{
				Protocol:     protocol,
				Destinations: []string{destination},
				Ports:        ports,
				PortRange:    portRange,
				IcmpInfo:     icmpInfo,
				Log:          log,
			}

			validationErr = rule.Validate()
		})

		itAllowsPorts := func() {
			Describe("ports", func() {
				Context("with a valid port", func() {
					BeforeEach(func() {
						portRange = nil
						ports = []uint32{1}
					})

					It("passes validation and does not return an error", func() {
						Expect(validationErr).NotTo(HaveOccurred())
					})
				})

				Context("with an empty ports list", func() {
					BeforeEach(func() {
						portRange = nil
						ports = []uint32{}
					})

					It("returns an error", func() {
						Expect(validationErr).To(MatchError(ContainSubstring("ports")))
					})
				})

				Context("with an invalid port", func() {
					BeforeEach(func() {
						portRange = nil
						ports = []uint32{0}
					})

					It("returns an error", func() {
						Expect(validationErr).To(MatchError(ContainSubstring("ports")))
					})
				})

			})

			Describe("port range", func() {
				Context("when it is a valid port range", func() {
					BeforeEach(func() {
						ports = nil
						portRange = &models.PortRange{1, 65535}
					})

					It("passes validation and does not return an error", func() {
						Expect(validationErr).NotTo(HaveOccurred())
					})
				})

				Context("when port range has a start value greater than the end value", func() {
					BeforeEach(func() {
						ports = nil
						portRange = &models.PortRange{1024, 1}
					})

					It("returns an error", func() {
						Expect(validationErr).To(MatchError(ContainSubstring("port_range")))
					})
				})
			})

			Context("when ports and port range are provided", func() {
				BeforeEach(func() {
					portRange = &models.PortRange{1, 65535}
					ports = []uint32{1}
				})

				It("returns an error", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("Invalid: ports and port_range provided")))
				})
			})

			Context("when ports and port range are not provided", func() {
				BeforeEach(func() {
					portRange = nil
					ports = nil
				})

				It("returns an error", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("Missing required field: ports or port_range")))
				})
			})
		}

		itExpectsADestination := func() {
			Describe("destination", func() {
				Context("when the destination is valid", func() {
					BeforeEach(func() {
						destination = "1.2.3.4/32"
					})

					It("passes validation and does not return an error", func() {
						Expect(validationErr).NotTo(HaveOccurred())
					})
				})

				Context("when the destination is invalid", func() {
					BeforeEach(func() {
						destination = "garbage/32"
					})

					It("returns an error", func() {
						Expect(validationErr).To(MatchError(ContainSubstring("destination")))
					})
				})
			})
		}

		itFailsWithPorts := func() {
			Context("when Port range is provided", func() {
				BeforeEach(func() {
					ports = nil
					portRange = &models.PortRange{1, 65535}
				})

				It("fails", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("port_range")))
				})
			})

			Context("when Ports are provided", func() {
				BeforeEach(func() {
					portRange = nil
					ports = []uint32{1}
				})

				It("fails", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("ports")))
				})
			})
		}

		itFailsWithICMPInfo := func() {
			Context("when ICMP info is provided", func() {
				BeforeEach(func() {
					icmpInfo = &models.ICMPInfo{}
				})
				It("fails", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("icmp_info")))
				})
			})
		}

		itAllowsLogging := func() {
			Context("when log is true", func() {
				BeforeEach(func() {
					log = true
				})

				It("succeeds", func() {
					Expect(validationErr).NotTo(HaveOccurred())
				})
			})
		}

		Describe("destination", func() {
			BeforeEach(func() {
				ports = []uint32{1}
			})

			Context("when its an IP Address", func() {
				BeforeEach(func() {
					destination = "8.8.8.8"
				})

				It("passes validation and does not return an error", func() {
					Expect(validationErr).NotTo(HaveOccurred())
				})
			})

			Context("when its a range of IP Addresses", func() {
				BeforeEach(func() {
					destination = "8.8.8.8-8.8.8.9"
				})

				It("passes validation and does not return an error", func() {
					Expect(validationErr).NotTo(HaveOccurred())
				})

				Context("and the range is not valid", func() {
					BeforeEach(func() {
						destination = "1.2.3.4 - 1.2.1.3"
					})

					It("fails", func() {
						Expect(validationErr).To(MatchError(ContainSubstring("destination")))
					})
				})
			})

			Context("when its a CIDR", func() {
				BeforeEach(func() {
					destination = "8.8.8.8/16"
				})

				It("passes validation and does not return an error", func() {
					Expect(validationErr).NotTo(HaveOccurred())
				})
			})

			Context("when its not valid", func() {
				BeforeEach(func() {
					destination = "8.8"
				})

				It("fails", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("destination")))
				})
			})
		})

		Describe("protocol", func() {
			Context("when the protocol is tcp", func() {
				BeforeEach(func() {
					protocol = "tcp"
					ports = []uint32{1}
				})

				itFailsWithICMPInfo()
				itAllowsPorts()
				itExpectsADestination()
				itAllowsLogging()
			})

			Context("when the protocol is udp", func() {
				BeforeEach(func() {
					protocol = "udp"
					ports = []uint32{1}
				})

				itFailsWithICMPInfo()
				itAllowsPorts()
				itExpectsADestination()
				itAllowsLogging()
			})

			Context("when the protocol is icmp", func() {
				BeforeEach(func() {
					protocol = "icmp"
					icmpInfo = &models.ICMPInfo{}
				})

				itExpectsADestination()
				itFailsWithPorts()
				itAllowsLogging()

				Context("when no ICMPInfo is provided", func() {
					BeforeEach(func() {
						icmpInfo = nil
					})

					It("fails", func() {
						Expect(validationErr).To(HaveOccurred())
					})
				})
			})

			Context("when the protocol is all", func() {
				BeforeEach(func() {
					protocol = "all"
				})

				itFailsWithICMPInfo()
				itExpectsADestination()
				itFailsWithPorts()
				itAllowsLogging()
			})

			Context("when the protocol is invalid", func() {
				BeforeEach(func() {
					protocol = "foo"
				})

				It("returns an error", func() {
					Expect(validationErr).To(MatchError(ContainSubstring("protocol")))
				})
			})
		})

		Context("when thre are multiple field validations", func() {
			BeforeEach(func() {
				protocol = "tcp"
				destination = "garbage"
				portRange = &models.PortRange{443, 80}
			})

			It("aggregates validation errors", func() {
				Expect(validationErr).To(MatchError(ContainSubstring("port_range")))
				Expect(validationErr).To(MatchError(ContainSubstring("destination")))
			})
		})
	})

	Describe("serialization", func() {
		var securityGroupJson string
		var securityGroup models.SecurityGroupRule

		BeforeEach(func() {
			securityGroupJson = `{
        "protocol": "all",
        "destinations": [
          "0.0.0.0-9.255.255.255"
        ],
        "log": false,
				"annotations":["quack"]
      }`

			securityGroup = models.SecurityGroupRule{
				Protocol:     "all",
				Destinations: []string{"0.0.0.0-9.255.255.255"},
				Log:          false,
				Annotations:  []string{"quack"},
			}
		})

		It("successfully round trips through json and protobuf", func() {
			jsonSerialization, err := json.Marshal(securityGroup)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonSerialization).To(MatchJSON(securityGroupJson))

			protoSerialization, err := proto.Marshal(&securityGroup)
			Expect(err).NotTo(HaveOccurred())

			var protoDeserialization models.SecurityGroupRule
			err = proto.Unmarshal(protoSerialization, &protoDeserialization)
			Expect(err).NotTo(HaveOccurred())

			Expect(protoDeserialization).To(Equal(securityGroup))
		})

		Context("when annotations are empty", func() {
			BeforeEach(func() {
				securityGroupJson = `{
					"protocol": "all",
					"destinations": [
						"0.0.0.0-9.255.255.255"
					],
					"log": false
				}`

				securityGroup.Annotations = []string{}
			})

			It("successfully json serializes empty arrays to nil", func() {
				jsonSerialization, err := json.Marshal(securityGroup)
				Expect(err).NotTo(HaveOccurred())
				Expect(jsonSerialization).To(MatchJSON(securityGroupJson))
			})
		})
	})
})
