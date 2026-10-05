package models_test

import (
	"encoding/json"
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
