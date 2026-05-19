package tests

import (
	"strings"
	"testing"

	"github.com/OryaHub/agent-os-instance/libs/bdd"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

// ---------------------------------------------------------------------------
// ExtractTagKey
// ---------------------------------------------------------------------------

func TestExtractTagKey(t *testing.T) {
	bdd.Feature(t, "TestExtractTagKey", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("tag com separador key:value", func(s *bdd.Scenario) {
			var key string
			s.When("ExtractTagKey é chamado com 'env:prod'", func() {
				key = utils.ExtractTagKey("env:prod")
			})
			s.Then("deve retornar apenas 'env'", func(t *testing.T) {
				bdd.AssertEqual(t, "env", key, "chave deve ser 'env'")
			})
		})

		Scenario("tag sem separador", func(s *bdd.Scenario) {
			var key string
			s.When("ExtractTagKey é chamado com 'standalone'", func() {
				key = utils.ExtractTagKey("standalone")
			})
			s.Then("deve retornar a tag inteira", func(t *testing.T) {
				bdd.AssertEqual(t, "standalone", key, "chave deve ser a tag inteira")
			})
		})
	})
}

// ---------------------------------------------------------------------------
// MergeTagsByKeyPriority
// ---------------------------------------------------------------------------

func TestMergeTagsByKeyPriority(t *testing.T) {
	bdd.Feature(t, "TestMergeTagsByKeyPriority", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("tag existente tem prioridade sobre tag nova com mesma chave", func(s *bdd.Scenario) {
			var merged []string
			s.When("existing tem 'env:prod' e incoming tem 'env:stage'", func() {
				merged = utils.MergeTagsByKeyPriority([]string{"env:prod"}, []string{"env:stage"})
			})
			s.Then("merged deve conter apenas 'env:prod'", func(t *testing.T) {
				bdd.AssertEqual(t, 1, len(merged), "deve ter exatamente 1 tag")
				bdd.AssertEqual(t, "env:prod", merged[0], "tag existente deve ser mantida")
			})
		})

		Scenario("nova tag com chave diferente é adicionada", func(s *bdd.Scenario) {
			var merged []string
			s.When("existing tem 'env:prod' e incoming tem 'service:web'", func() {
				merged = utils.MergeTagsByKeyPriority([]string{"env:prod"}, []string{"service:web"})
			})
			s.Then("merged deve conter as duas tags", func(t *testing.T) {
				bdd.AssertEqual(t, 2, len(merged), "deve ter 2 tags")
			})
		})

		Scenario("incoming vazio não altera existing", func(s *bdd.Scenario) {
			var merged []string
			s.When("incoming é vazio", func() {
				merged = utils.MergeTagsByKeyPriority([]string{"env:prod"}, []string{})
			})
			s.Then("merged deve ser igual ao existing", func(t *testing.T) {
				bdd.AssertEqual(t, 1, len(merged), "deve ter 1 tag")
				bdd.AssertEqual(t, "env:prod", merged[0], "tag original deve ser mantida")
			})
		})

		Scenario("existing vazio resulta apenas nas incoming tags", func(s *bdd.Scenario) {
			var merged []string
			s.When("existing é vazio e incoming tem 'env:prod'", func() {
				merged = utils.MergeTagsByKeyPriority([]string{}, []string{"env:prod"})
			})
			s.Then("merged deve conter 'env:prod'", func(t *testing.T) {
				bdd.AssertEqual(t, 1, len(merged), "deve ter 1 tag")
				bdd.AssertEqual(t, "env:prod", merged[0], "incoming tag deve ser adicionada")
			})
		})
	})
}

// ---------------------------------------------------------------------------
// ParseInlineTagLine
// ---------------------------------------------------------------------------

func TestParseInlineTagLine(t *testing.T) {
	bdd.Feature(t, "TestParseInlineTagLine", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("linha ativa com duas tags", func(s *bdd.Scenario) {
			var tags []string
			s.When("ParseInlineTagLine é chamado com linha ativa", func() {
				tags = utils.ParseInlineTagLine(`tags: ["env:prod", "service:web"]`)
			})
			s.Then("deve retornar as duas tags", func(t *testing.T) {
				bdd.AssertEqual(t, 2, len(tags), "deve ter 2 tags")
				bdd.AssertEqual(t, "env:prod", tags[0], "primeira tag deve ser 'env:prod'")
				bdd.AssertEqual(t, "service:web", tags[1], "segunda tag deve ser 'service:web'")
			})
		})

		Scenario("linha comentada com uma tag", func(s *bdd.Scenario) {
			var tags []string
			s.When("ParseInlineTagLine é chamado com linha comentada", func() {
				tags = utils.ParseInlineTagLine(`# tags: ["env:prod"]`)
			})
			s.Then("deve retornar a tag", func(t *testing.T) {
				bdd.AssertEqual(t, 1, len(tags), "deve ter 1 tag")
				bdd.AssertEqual(t, "env:prod", tags[0], "tag deve ser 'env:prod'")
			})
		})

		Scenario("lista vazia retorna nil", func(s *bdd.Scenario) {
			var tags []string
			s.When("ParseInlineTagLine é chamado com lista vazia", func() {
				tags = utils.ParseInlineTagLine(`tags: []`)
			})
			s.Then("deve retornar lista vazia", func(t *testing.T) {
				bdd.AssertEqual(t, 0, len(tags), "deve retornar lista vazia")
			})
		})
	})
}

// ---------------------------------------------------------------------------
// ParseBlockTagItemLine
// ---------------------------------------------------------------------------

func TestParseBlockTagItemLine(t *testing.T) {
	bdd.Feature(t, "TestParseBlockTagItemLine", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {
		Scenario("item de bloco ativo", func(s *bdd.Scenario) {
			var tag string
			s.When("ParseBlockTagItemLine é chamado com item ativo", func() {
				tag = utils.ParseBlockTagItemLine(`  - "env:prod"`)
			})
			s.Then("deve retornar 'env:prod'", func(t *testing.T) {
				bdd.AssertEqual(t, "env:prod", tag, "tag deve ser 'env:prod'")
			})
		})

		Scenario("item de bloco comentado", func(s *bdd.Scenario) {
			var tag string
			s.When("ParseBlockTagItemLine é chamado com item comentado", func() {
				tag = utils.ParseBlockTagItemLine(`#   - "env:prod"`)
			})
			s.Then("deve retornar 'env:prod'", func(t *testing.T) {
				bdd.AssertEqual(t, "env:prod", tag, "tag deve ser 'env:prod'")
			})
		})
	})
}

// ---------------------------------------------------------------------------
// MergeTagsInDatadogConfig
// ---------------------------------------------------------------------------

func TestMergeTagsInDatadogConfig(t *testing.T) {
	bdd.Feature(t, "TestMergeTagsInDatadogConfig", func(t *testing.T, Scenario func(description string, steps func(s *bdd.Scenario))) {

		Scenario("formato inline ativo — tag existente tem prioridade", func(s *bdd.Scenario) {
			var result string
			var err error
			content := `hostname: myhost
tags: ["env:prod", "service:web"]
api_key: abc123`
			s.When("MergeTagsInDatadogConfig é chamado com nova tag 'env:stage'", func() {
				result, err = utils.MergeTagsInDatadogConfig(content, []string{"env:stage", "region:us-east-1"})
			})
			s.Then("nenhum erro deve ocorrer", func(t *testing.T) {
				bdd.AssertNoError(t, err, "não deve retornar erro")
			})
			s.Then("tag 'env:prod' deve ser preservada", func(t *testing.T) {
				bdd.AssertTrue(t, strings.Contains(result, `"env:prod"`), "deve conter env:prod")
			})
			s.Then("tag 'env:stage' não deve estar presente", func(t *testing.T) {
				bdd.AssertFalse(t, strings.Contains(result, "env:stage"), "não deve conter env:stage")
			})
			s.Then("nova tag 'region:us-east-1' deve ser adicionada", func(t *testing.T) {
				bdd.AssertTrue(t, strings.Contains(result, "region:us-east-1"), "deve conter region:us-east-1")
			})
			s.Then("o restante do arquivo deve ser preservado", func(t *testing.T) {
				bdd.AssertTrue(t, strings.Contains(result, "hostname: myhost"), "deve conter hostname")
				bdd.AssertTrue(t, strings.Contains(result, "api_key: abc123"), "deve conter api_key")
			})
		})

		Scenario("formato inline comentado — tags comentadas ignoradas, newTags usadas como estão", func(s *bdd.Scenario) {
			var result string
			var err error
			content := `hostname: myhost
# tags: ["env:prod"]
api_key: abc123`
			s.When("MergeTagsInDatadogConfig é chamado com nova tag 'service:web'", func() {
				result, err = utils.MergeTagsInDatadogConfig(content, []string{"service:web"})
			})
			s.Then("nenhum erro deve ocorrer", func(t *testing.T) {
				bdd.AssertNoError(t, err, "não deve retornar erro")
			})
			s.Then("linha tags não deve estar comentada", func(t *testing.T) {
				bdd.AssertFalse(t, strings.Contains(result, "# tags:"), "tags não deve estar comentada")
			})
			s.Then("tag 'env:prod' comentada não deve estar presente", func(t *testing.T) {
				bdd.AssertFalse(t, strings.Contains(result, "env:prod"), "não deve conter env:prod — estava comentada")
			})
			s.Then("tag 'service:web' deve ser usada", func(t *testing.T) {
				bdd.AssertTrue(t, strings.Contains(result, "service:web"), "deve conter service:web")
			})
		})

		Scenario("formato bloco ativo — tag existente tem prioridade", func(s *bdd.Scenario) {
			var result string
			var err error
			content := `hostname: myhost
tags:
  - "env:prod"
  - "service:web"
api_key: abc123`
			s.When("MergeTagsInDatadogConfig é chamado com 'env:stage' e 'region:us-east-1'", func() {
				result, err = utils.MergeTagsInDatadogConfig(content, []string{"env:stage", "region:us-east-1"})
			})
			s.Then("nenhum erro deve ocorrer", func(t *testing.T) {
				bdd.AssertNoError(t, err, "não deve retornar erro")
			})
			s.Then("tag 'env:prod' deve ser preservada", func(t *testing.T) {
				bdd.AssertTrue(t, strings.Contains(result, `"env:prod"`), "deve conter env:prod")
			})
			s.Then("tag 'env:stage' não deve estar presente", func(t *testing.T) {
				bdd.AssertFalse(t, strings.Contains(result, "env:stage"), "não deve conter env:stage")
			})
			s.Then("nova tag 'region:us-east-1' deve ser adicionada", func(t *testing.T) {
				bdd.AssertTrue(t, strings.Contains(result, "region:us-east-1"), "deve conter region:us-east-1")
			})
		})

		Scenario("formato bloco comentado — tags comentadas ignoradas, newTags usadas como estão", func(s *bdd.Scenario) {
			var result string
			var err error
			content := `hostname: myhost
# tags:
#   - "env:prod"
api_key: abc123`
			s.When("MergeTagsInDatadogConfig é chamado com 'service:web'", func() {
				result, err = utils.MergeTagsInDatadogConfig(content, []string{"service:web"})
			})
			s.Then("nenhum erro deve ocorrer", func(t *testing.T) {
				bdd.AssertNoError(t, err, "não deve retornar erro")
			})
			s.Then("linha tags não deve estar comentada", func(t *testing.T) {
				bdd.AssertFalse(t, strings.Contains(result, "# tags"), "tags não deve estar comentada")
			})
			s.Then("tag 'env:prod' comentada não deve estar presente", func(t *testing.T) {
				bdd.AssertFalse(t, strings.Contains(result, "env:prod"), "não deve conter env:prod — estava comentada")
			})
			s.Then("tag 'service:web' deve ser usada", func(t *testing.T) {
				bdd.AssertTrue(t, strings.Contains(result, "service:web"), "deve conter service:web")
			})
		})

		Scenario("sem seção tags — conteúdo retornado inalterado", func(s *bdd.Scenario) {
			var result string
			var err error
			content := `hostname: myhost
api_key: abc123`
			s.When("MergeTagsInDatadogConfig é chamado em arquivo sem seção tags", func() {
				result, err = utils.MergeTagsInDatadogConfig(content, []string{"env:prod"})
			})
			s.Then("nenhum erro deve ocorrer", func(t *testing.T) {
				bdd.AssertNoError(t, err, "não deve retornar erro")
			})
			s.Then("conteúdo deve ser idêntico ao original", func(t *testing.T) {
				bdd.AssertEqual(t, content, result, "conteúdo não deve ser alterado")
			})
		})

		Scenario("newTags vazio com seção comentada — seção descomentada e vazia", func(s *bdd.Scenario) {
			var result string
			var err error
			content := `hostname: myhost
# tags: ["env:prod"]
api_key: abc123`
			s.When("MergeTagsInDatadogConfig é chamado com newTags vazio", func() {
				result, err = utils.MergeTagsInDatadogConfig(content, []string{})
			})
			s.Then("nenhum erro deve ocorrer", func(t *testing.T) {
				bdd.AssertNoError(t, err, "não deve retornar erro")
			})
			s.Then("linha tags deve estar descomentada", func(t *testing.T) {
				bdd.AssertFalse(t, strings.Contains(result, "# tags:"), "tags não deve estar comentada")
			})
			s.Then("tag 'env:prod' comentada não deve estar presente", func(t *testing.T) {
				bdd.AssertFalse(t, strings.Contains(result, "env:prod"), "não deve conter env:prod — estava comentada")
			})
		})
	})
}
