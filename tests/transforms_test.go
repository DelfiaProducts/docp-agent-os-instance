package tests

import (
	"log"
	"testing"

	"github.com/OryaHub/agent-os-instance/libs/bdd"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

func TestTransformMapToSlice(t *testing.T) {
	bdd.Feature(t, "TestTransformMapToSlice", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve transformar map em slice sem erro", func(s *bdd.Scenario) {
			var mapp map[string]interface{}
			s.When("map é definido", func() {
				mapp = map[string]interface{}{
					"group":   "dev",
					"app":     1,
					"machine": nil,
				}
			})
			s.Then("não deve retornar erro", func(t *testing.T) {
				slc, err := utils.TransformMapToSlice(mapp)
				if err != nil {
					t.Errorf("TestTransformMapToSlice: expect(nil) - got(%s)\n", err.Error())
				}
				log.Printf("slc: %v\n", slc)
			})
		})
	})
}

func TestTransformIsVersionGreater(t *testing.T) {
	bdd.Feature(t, "TestTransformIsVersionGreater", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("Deve verificar se a versão é maior", func(s *bdd.Scenario) {
			var version string
			var result bool
			var err error
			s.When("a versão é definida", func() {
				version = "7.46.0"
			})
			s.Then("deve retornar true se a nova versão for maior", func(t *testing.T) {
				result, err = utils.IsVersionGreater(version, "8.2.4")
				if err != nil {
					t.Errorf("TestTransformIsVersionGreater: expect(nil) - got(%s)\n", err.Error())
				}
			})
			bdd.Printf("RESULT: %t\n", result)
		})
	})
}
