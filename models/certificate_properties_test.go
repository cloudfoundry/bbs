package models_test

import (
	"encoding/json"
	"strings"
	"testing"

	"code.cloudfoundry.org/bbs/models"
	"github.com/gogo/protobuf/proto"
)

func TestServiceAccountCertificatePropertiesRoundTrip(t *testing.T) {
	var properties models.CertificateProperties
	if err := json.Unmarshal([]byte(`{"organizational_unit":["app:app-guid"],"service_account":{"name":"payments-worker"}}`), &properties); err != nil {
		t.Fatal(err)
	}
	wire, err := proto.Marshal(&properties)
	if err != nil {
		t.Fatal(err)
	}
	var decoded models.CertificateProperties
	if err := proto.Unmarshal(wire, &decoded); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(&decoded)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatal(err)
	}
	identity, ok := result["service_account"].(map[string]interface{})
	if !ok || identity["name"] != "payments-worker" {
		t.Fatalf("service account lost in protobuf round trip: %s", payload)
	}
}

func TestServiceAccountCertificatePropertiesValidation(t *testing.T) {
	for _, name := range []string{"", "ab", "Payments-worker", "-payments", "payments-", "payments.reader", "payments.svc.identity", "payments_worker", strings.Repeat("a", 64)} {
		t.Run(name, func(t *testing.T) {
			properties := &models.CertificateProperties{ServiceAccount: &models.ServiceAccount{Name: name}}
			if err := properties.Validate(); err == nil {
				t.Fatal("noncanonical service account accepted")
			}
			runInfo := models.DesiredLRPRunInfo{DesiredLRPKey: models.DesiredLRPKey{ProcessGuid: "process"}, Action: models.WrapAction(&models.RunAction{Path: "/bin/true", User: "vcap"}), CertificateProperties: properties}
			if err := runInfo.Validate(); err == nil || !strings.Contains(err.Error(), "service_account") {
				t.Fatalf("LRP validation lost service account error: %v", err)
			}
			task := models.TaskDefinition{RootFs: "preloaded:stack", Action: runInfo.Action, CertificateProperties: properties}
			if err := task.Validate(); err == nil || !strings.Contains(err.Error(), "service_account") {
				t.Fatalf("task validation lost service account error: %v", err)
			}
		})
	}
	for _, properties := range []*models.CertificateProperties{{}, {ServiceAccount: &models.ServiceAccount{Name: "payments-worker"}}, {ServiceAccount: &models.ServiceAccount{Name: strings.Repeat("a", 63)}}} {
		if err := properties.Validate(); err != nil {
			t.Fatalf("valid certificate properties rejected: %v", err)
		}
	}
}
