package tests

import (
	"log"
	"testing"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/bdd"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
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
