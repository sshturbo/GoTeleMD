package formatter

import (
	"encoding/json"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/parser"
	"github.com/sshturbo/GoTeleMD/pkg/types"
	"github.com/sshturbo/GoTeleMD/pkg/utils"
)

type blockProcessor struct {
	input    chan processTask
	output   chan processResult
	wg       sync.WaitGroup
	errChan  chan error
	config   *types.Config
	stopChan chan struct{}
}

type processTask struct {
	block  internal.Block
	config *types.Config
	index  int
	part   int
	total  int
}

type processResult struct {
	content string
	index   int
	part    int
}

func newBlockProcessor(config *types.Config) *blockProcessor {
	numWorkers := config.NumWorkers
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	p := &blockProcessor{
		input:    make(chan processTask, config.WorkerQueueSize),
		output:   make(chan processResult, config.WorkerQueueSize),
		errChan:  make(chan error, numWorkers),
		config:   config,
		stopChan: make(chan struct{}),
	}

	for i := 0; i < numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	return p
}

func (p *blockProcessor) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.input:
			if !ok {
				return
			}

			startTime := time.Now()

			// Recuperação de pânico
			defer func() {
				if r := recover(); r != nil {
					p.errChan <- types.NewError(types.ErrProcessingFailed, "worker panic", r.(error))
				}
			}()

			rendered := RenderBlock(task.block, task.config)

			if task.config.EnableDebugLogs {
				utils.LogDebug("🔧 Worker %d - Bloco %d (Parte %d/%d): Tipo=%d, Tamanho=%d, Tempo=%v",
					id, task.index+1, task.part, task.total, task.block.Type, len(rendered), time.Since(startTime))
			}

			select {
			case p.output <- processResult{content: rendered, index: task.index, part: task.part}:
			case <-p.stopChan:
				return
			}

		case <-p.stopChan:
			return
		}
	}
}

func (p *blockProcessor) close() {
	close(p.stopChan)
	close(p.input)
	p.wg.Wait()
	close(p.output)
	close(p.errChan)
}

func ConvertMarkdown(input string, config *types.Config) (string, error) {
	if input == "" {
		return "", types.NewError(types.ErrInvalidInput, "input cannot be empty", nil)
	}

	startTime := time.Now()
	defer func() {
		utils.LogPerformance("Convert total", time.Since(startTime))
	}()

	utils.LogDebug("🔄 Iniciando conversão do Markdown")
	utils.LogDebug("📊 Configurações:")
	utils.LogDebug("   - Nível de segurança: %d", config.SafetyLevel)
	utils.LogDebug("   - Tamanho máximo: %d", config.MaxMessageLength)
	utils.LogDebug("   - Workers: %d", config.NumWorkers)
	utils.LogDebug("   - Fila por worker: %d", config.WorkerQueueSize)
	utils.LogDebug("   - Partes simultâneas: %d", config.MaxConcurrentParts)

	utils.LogDebug("📝 Texto original:\n%s", input)

	response, err := parser.BreakLongText(strings.TrimSpace(input), config.MaxMessageLength)
	if err != nil {
		return "", types.NewError(types.ErrProcessingFailed, "failed to break text", err)
	}

	if config.EnableDebugLogs {
		jsonResponse, _ := json.MarshalIndent(response, "", "  ")
		utils.LogDebug("📦 Divisão em partes:")
		utils.LogDebug("   - Total de partes: %d", response.TotalParts)
		utils.LogDebug("   - ID da mensagem: %s", response.MessageID)
		utils.LogDebug("   - Estrutura JSON:\n%s", string(jsonResponse))
	}

	processor := newBlockProcessor(config)
	defer processor.close()

	outputParts := make([]string, len(response.Parts))

	semaphore := make(chan struct{}, config.MaxConcurrentParts)
	var errMutex sync.Mutex
	var firstErr error

	var wg sync.WaitGroup
	for partIdx, part := range response.Parts {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(partIdx int, part types.MessagePart) {
			defer wg.Done()
			defer func() { <-semaphore }()

			blocks := parser.Tokenize(strings.TrimSpace(part.Content))
			results := make(map[int]string)
			pendingBlocks := len(blocks)

			utils.LogDebug("🔍 Processando parte %d/%d", part.Part, response.TotalParts)
			utils.LogDebug("   - Blocos encontrados: %d", len(blocks))

			for i, block := range blocks {
				select {
				case processor.input <- processTask{
					block:  block,
					config: config,
					index:  i,
					part:   part.Part,
					total:  response.TotalParts,
				}:
				case err := <-processor.errChan:
					errMutex.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMutex.Unlock()
					return
				}
			}

			for pendingBlocks > 0 {
				select {
				case result := <-processor.output:
					results[result.index] = result.content
					pendingBlocks--
				case err := <-processor.errChan:
					errMutex.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMutex.Unlock()
					return
				}
			}

			
			var output strings.Builder
			output.Grow(len(part.Content))

			for i := 0; i < len(blocks); i++ {
				if i > 0 {
					if blocks[i].Type == internal.BlockTitle || blocks[i-1].Type == internal.BlockTitle {
						output.WriteString("\n\n")
					} else if blocks[i].Type == internal.BlockCode || blocks[i-1].Type == internal.BlockCode {
						output.WriteString("\n\n")
					} else if blocks[i].Type == internal.BlockList || blocks[i-1].Type == internal.BlockList {
						output.WriteString("\n\n")
					} else {
						output.WriteString("\n\n")
					}
				}
				output.WriteString(results[i])
			}

			formattedContent := strings.TrimSpace(output.String())
			outputParts[partIdx] = formattedContent

			if config.EnableDebugLogs {
				utils.LogDebug("📤 Parte %d formatada:", part.Part)
				utils.LogDebug("   - Tamanho: %d caracteres", len(formattedContent))
				utils.LogDebug("   - Conteúdo:\n%s", formattedContent)
			}
		}(partIdx, part)
	}

	wg.Wait()

	if firstErr != nil {
		return "", firstErr
	}

	result := strings.TrimSpace(strings.Join(outputParts, "\n\n"))
	utils.LogDebug("✅ Conversão finalizada")
	utils.LogDebug("   - Tamanho final: %d caracteres", len(result))
	utils.LogDebug("   - Partes geradas: %d", len(outputParts))

	return result, nil
}
