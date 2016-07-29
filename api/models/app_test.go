package models_test

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/convox/rack/api/models"
	"github.com/convox/rack/manifest"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("RACK", "convox-test")
}

func TestRackStackName(t *testing.T) {
	r := models.App{
		Name: "convox-test",
	}

	assert.Equal(t, "convox-test", r.StackName())
}

func TestAppStackName(t *testing.T) {
	// unbound app (no rack prefix)
	a := models.App{
		Name: "httpd",
		Tags: map[string]string{
			"Type":   "app",
			"System": "convox",
			"Rack":   "convox-test",
		},
	}

	assert.Equal(t, "httpd", a.StackName())

	// bound app (rack prefix, and Name tag)
	a = models.App{
		Name: "httpd",
		Tags: map[string]string{
			"Name":   "httpd",
			"Type":   "app",
			"System": "convox",
			"Rack":   "convox-test",
		},
	}

	assert.Equal(t, "convox-test-httpd", a.StackName())
}

func TestAppFormation(t *testing.T) {
	a := models.App{
		Name: "testerness",
	}

	m := manifest.Manifest{
		Services: map[string]manifest.Service{
			"api": manifest.Service{
				Name: "api",
				Ports: []manifest.Port{
					manifest.Port{
						Balancer:  80,
						Container: 3000,
						Public:    true,
					},
				},
			},
		},
	}

	res := map[string]interface{}{}

	f, err := a.Formation(m)
	if err != nil {
		t.Error(err)
		return
	}

	err = json.Unmarshal([]byte(f), &res)
	if err != nil {
		t.Error(err)
		return
	}

	resources := res["Resources"].(map[string]interface{})
	// outputs := res["Outputs"].(map[string]interface{})
	// parameters := res["Parameters"].(map[string]interface{})
	// conditions := res["Conditions"].(map[string]interface{})
	// mappings := res["Mappings"].(map[string]interface{})

	balancer := resources["BalancerApi"].(map[string]interface{})
	balancerProps := balancer["Properties"].(map[string]interface{})
	listeners := balancerProps["Listeners"].([]interface{})

	assert.Equal(t, balancer["Type"], "AWS::ElasticLoadBalancing::LoadBalancer")
	assert.Equal(t, balancer["Condition"], "EnabledApi")
	assert.Equal(t, len(listeners), 2)

	apiBalancer := listeners[0].(map[string]interface{})
	byts, err := json.Marshal(apiBalancer)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(string(byts))

	for k, v := range listeners[0].(map[string]interface{})["Fn::If"].([]interface{}) {
		log.Print(k)
		log.Print(v)
	}
}
