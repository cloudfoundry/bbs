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
	if err := json.Unmarshal([]byte(`{"organizational_unit":["app:app-guid"],"service_account":{"name":"payments-worker","prefix":"payments-team"}}`), &properties); err != nil {
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
	if !ok || identity["name"] != "payments-worker" || identity["prefix"] != "payments-team" {
		t.Fatalf("service account lost in protobuf round trip: %s", payload)
	}
}

func TestQualifiedServiceAccountPrefixRoundTrip(t *testing.T) {
	var properties models.CertificateProperties
	if err := json.Unmarshal([]byte(`{"service_account":{"name":"payments-worker","prefix":"payments-team"}}`), &properties); err != nil {
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
	if !strings.Contains(string(payload), `"prefix":"payments-team"`) {
		t.Fatalf("qualified prefix lost: %s", payload)
	}
}

func TestQualifiedPrefixRejectsNoncanonicalLabels(t *testing.T) {
	for _, prefix := range []string{"", "ab", "Upper", "a.b", "-abc", "abc-", "abc_def", strings.Repeat("a", 64)} {
		t.Run(prefix, func(t *testing.T) {
			properties := &models.CertificateProperties{ServiceAccount: &models.ServiceAccount{Name: "worker-app", Prefix: prefix}}
			runInfo := models.DesiredLRPRunInfo{DesiredLRPKey: models.DesiredLRPKey{ProcessGuid: "process", Domain: "cf-apps"}, Action: models.WrapAction(&models.RunAction{Path: "/bin/true", User: "vcap"}), CertificateProperties: properties}
			task := models.TaskDefinition{RootFs: "preloaded:stack", Action: runInfo.Action, CertificateProperties: properties}
			for name, validate := range map[string]func() error{
				"certificate properties": properties.Validate,
				"LRP":                    runInfo.Validate,
				"task":                   task.Validate,
			} {
				t.Run(name, func(t *testing.T) {
					if err := validate(); err == nil || !strings.Contains(err.Error(), "certificate_properties.service_account.prefix") {
						t.Fatalf("expected invalid prefix error for %q, got: %v", prefix, err)
					}
				})
			}
		})
	}
}

func TestServiceAccountCertificatePropertiesValidation(t *testing.T) {
	for _, name := range []string{"", "ab", "Payments-worker", "-payments", "payments-", "payments.reader", "payments.svc.identity", "payments_worker", strings.Repeat("a", 64)} {
		t.Run(name, func(t *testing.T) {
			properties := &models.CertificateProperties{ServiceAccount: &models.ServiceAccount{Name: name, Prefix: "payments-team"}}
			if err := properties.Validate(); err == nil || !strings.Contains(err.Error(), "certificate_properties.service_account.name") {
				t.Fatal("noncanonical service account accepted")
			}
			runInfo := models.DesiredLRPRunInfo{DesiredLRPKey: models.DesiredLRPKey{ProcessGuid: "process", Domain: "cf-apps"}, Action: models.WrapAction(&models.RunAction{Path: "/bin/true", User: "vcap"}), CertificateProperties: properties}
			if err := runInfo.Validate(); err == nil || !strings.Contains(err.Error(), "service_account") {
				t.Fatalf("LRP validation lost service account error: %v", err)
			}
			task := models.TaskDefinition{RootFs: "preloaded:stack", Action: runInfo.Action, CertificateProperties: properties}
			if err := task.Validate(); err == nil || !strings.Contains(err.Error(), "service_account") {
				t.Fatalf("task validation lost service account error: %v", err)
			}
		})
	}
	for _, properties := range []*models.CertificateProperties{nil, {}, {ServiceAccount: &models.ServiceAccount{Name: "payments-worker", Prefix: "payments-team"}}, {ServiceAccount: &models.ServiceAccount{Name: "a-0", Prefix: "0-a"}}, {ServiceAccount: &models.ServiceAccount{Name: strings.Repeat("a", 63), Prefix: strings.Repeat("b", 63)}}} {
		if properties != nil {
			if err := properties.Validate(); err != nil {
				t.Fatalf("valid certificate properties rejected: %v", err)
			}
		}
		runInfo := models.DesiredLRPRunInfo{DesiredLRPKey: models.DesiredLRPKey{ProcessGuid: "process", Domain: "cf-apps"}, Action: models.WrapAction(&models.RunAction{Path: "/bin/true", User: "vcap"}), CertificateProperties: properties}
		if err := runInfo.Validate(); err != nil {
			t.Fatalf("valid LRP rejected: %v", err)
		}
		task := models.TaskDefinition{RootFs: "preloaded:stack", Action: runInfo.Action, CertificateProperties: properties}
		if err := task.Validate(); err != nil {
			t.Fatalf("valid task rejected: %v", err)
		}
	}
}

func TestAssembledDesiredLRPRetainsServiceAccountIdentity(t *testing.T) {
	properties := &models.CertificateProperties{ServiceAccount: &models.ServiceAccount{Name: "payments-worker", Prefix: "payments-team"}}
	desired := &models.DesiredLRP{}
	desired.AddRunInfo(models.DesiredLRPRunInfo{CertificateProperties: properties})
	if desired.CertificateProperties == nil || desired.CertificateProperties.ServiceAccount == nil || desired.CertificateProperties.ServiceAccount.Name != "payments-worker" || desired.CertificateProperties.ServiceAccount.Prefix != "payments-team" {
		t.Fatal("assembling stored LRP run info dropped account identity")
	}
}
