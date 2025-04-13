package main

import (
	"fmt"

	"github.com/sshturbo/GoTeleMD"
	"github.com/sshturbo/GoTeleMD/pkg/types"
)

func main() {
	// Create a new converter with custom configuration
	converter := GoTeleMD.NewConverter(
		types.WithSafetyLevel(GoTeleMD.SAFETYLEVELBASIC),
		types.WithTableAlignment(true),
		types.WithMaxMessageLength(4096),
		types.WithDebugLogs(true),
	)

	// Sample Markdown text with various elements
	markdown := `# Header 1
## Header 2
### Header 3

Normal text with **bold** and _italic_ formatting.

* Unordered list item 1
* Unordered list item 2
  * Nested item
* Unordered list item 3

1. Ordered list item 1
2. Ordered list item 2
3. Ordered list item 3

> This is a blockquote
> With multiple lines

` + "```go" + `
package main

func main() {
    fmt.Println("Hello World!")
}
` + "```" + `

Sem problemas, aqui está a redação novamente:

**Título: A Urgência da Transição para Energias Renováveis**

A crescente preocupação com as mudanças climáticas e a exaustão dos recursos naturais tem impulsionado a busca por alternativas energéticas mais sustentáveis. As energias renováveis, como solar, eólica e hidrelétrica, emergem como pilares essenciais para um futuro energético mais limpo e resiliente.

A transição para fontes renováveis não é apenas uma questão ambiental, mas também econômica e social. A dependência de combustíveis fósseis expõe países a flutuações de preços e instabilidades geopolíticas. Ao investir em energias renováveis, as nações podem fortalecer sua segurança energética, gerar empregos e impulsionar a inovação tecnológica.

Para ilustrar o potencial de cada fonte de energia renovável, observe a seguinte tabela:

| Fonte de Energia | Vantagens                                                                 | Desvantagens                                                                 |
| ---------------- | ------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| Solar            | Abundante, limpa, custo de manutenção relativamente baixo.              | Dependente das condições climáticas, custo inicial elevado.                  |
| Eólica           | Limpa, grande potencial em áreas com ventos constantes.                 | Ruído, impacto visual, pode afetar a vida selvagem.                         |
| Hidrelétrica     | Fonte de energia estabelecida, alta capacidade de geração.                | Impacto ambiental significativo (desmatamento, alteração de ecossistemas). |
| Biomassa         | Utiliza resíduos orgânicos, reduz a emissão de metano.                   | Pode competir com a produção de alimentos, emissão de poluentes na queima. |

Além disso, a implementação de políticas de incentivo e regulamentação é crucial para acelerar a transição energética. Governos e empresas devem trabalhar em conjunto para promover a pesquisa e o desenvolvimento de tecnologias renováveis, bem como para criar um ambiente favorável a investimentos e à criação de empregos.

A tabela abaixo mostra exemplos de incentivos governamentais em diferentes países:

| País     | Incentivo                                                                  | Impacto                                                                      |
| -------- | -------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| Alemanha | Tarifas de alimentação para energia solar e eólica.                        | Impulsionou a instalação de painéis solares e turbinas eólicas.               |
| Brasil   | Leilões de energia para projetos de energia renovável.                     | Aumentou a capacidade instalada de energia eólica e solar.                    |
| China    | Subsídios para a fabricação de equipamentos de energia renovável.          | Tornou a China líder mundial na produção de equipamentos de energia limpa. |

Em conclusão, a transição para energias renováveis é um desafio complexo, mas essencial para garantir um futuro sustentável. Ao investir em tecnologias limpas, implementar políticas de incentivo e conscientizar a sociedade, podemos construir um mundo mais justo, próspero e resiliente. 😉

Se precisar de algo mais, é só me avisar! 😊

| Column 1 | Column 2 | Column 3 |
|:---------|:--------:|----------:|
| Left aqui tem um pornto.    | Center aqui tem um pornto.  | Right aqui tem um pornto.   |
| align aqui tem um pornto.   | align aqui tem um pornto.   | align aqui tem um pornto.  |`

	// Convert the Markdown text
	response, err := converter.Convert(markdown)
	if err != nil {
		panic(err)
	}

	// Print the converted messages
	for _, part := range response.Parts {
		fmt.Printf("Part %d: %s\n", part.Part, part.Content)
	}
}
