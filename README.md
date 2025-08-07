# Auction System - Go Expert Labs

Sistema de leilões desenvolvido em Go com funcionalidade de fechamento automático baseado em tempo configurável.

## Funcionalidades

- ✅ Criação de leilões
- ✅ Criação de lances (bids)
- ✅ **Fechamento automático de leilões** (Nova funcionalidade)
- ✅ Validação de tempo para lances
- ✅ API REST para todas as operações

## Nova Funcionalidade: Fechamento Automático

O sistema agora inclui uma funcionalidade de fechamento automático de leilões que:

1. **Calcula o tempo de expiração** baseado na variável de ambiente `AUCTION_INTERVAL`
2. **Executa uma goroutine em background** que verifica leilões vencidos a cada 30 segundos
3. **Atualiza automaticamente o status** de leilões expirados para `Completed`
4. **Usa concorrência segura** com mutexes para evitar race conditions

### Arquivos Modificados

- `internal/infra/database/auction/create_auction.go`: Implementação principal da funcionalidade

### Implementação Técnica

```go
// Goroutine que executa em background
func (ar *AuctionRepository) startAuctionClosingRoutine() {
    ticker := time.NewTicker(time.Second * 30)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            ar.checkAndCloseExpiredAuctions()
        }
    }
}
```

## Pré-requisitos

- Go 1.20+
- Docker e Docker Compose
- MongoDB

## Configuração de Ambiente

Crie um arquivo `.env` em `cmd/auction/.env` com as seguintes variáveis:

```env
# Configurações do MongoDB
MONGODB_URL=mongodb://mongodb:27017
MONGODB_DB=auction_db

# Configurações de leilão
AUCTION_INTERVAL=5m

# Configurações de batch (para bids)
BATCH_INSERT_INTERVAL=20s
MAX_BATCH_SIZE=4
```

### Variável AUCTION_INTERVAL

A variável `AUCTION_INTERVAL` define quanto tempo um leilão fica ativo. Exemplos:
- `30s` - 30 segundos
- `5m` - 5 minutos (padrão)
- `1h` - 1 hora
- `24h` - 24 horas

## Como Executar o Projeto

### 1. Ambiente de Desenvolvimento

```bash
# Clone o repositório
git clone <repository-url>
cd labs-auction-goexpert

# Baixe as dependências
go mod download

# Inicie o MongoDB
docker-compose up mongodb -d

# Configure as variáveis de ambiente
cp cmd/auction/.env.example cmd/auction/.env

# Execute a aplicação
go run cmd/auction/main.go
```

### 2. Com Docker Compose (Recomendado)

```bash
# Inicie todos os serviços
docker-compose up -d

# Verifique os logs
docker-compose logs -f app
```

### 3. Apenas MongoDB para desenvolvimento local

```bash
# Inicie apenas o MongoDB
docker-compose up mongodb -d

# Execute a aplicação localmente
MONGODB_URL=mongodb://localhost:27017 go run cmd/auction/main.go
```

## Executando os Testes

### 1. Testes da Funcionalidade de Fechamento Automático

```bash
# Certifique-se que o MongoDB está rodando (usar o mesmo do desenvolvimento)
docker-compose up mongodb -d

# Execute os testes
go test ./internal/infra/database/auction/ -v

# Ou execute um teste específico
go test ./internal/infra/database/auction/ -v -run TestAuctionAutoClose
```

### 2. Testes Disponíveis

- `TestAuctionAutoClose`: Verifica se leilões expirados são fechados automaticamente
- `TestAuctionStaysActive`: Verifica se leilões ainda válidos permanecem ativos
- `TestUpdateAuctionStatus`: Testa a atualização manual de status
- `TestGetAuctionInterval`: Testa o parsing da variável de ambiente

### 3. Limpeza após os testes

```bash
# Os testes usam o mesmo MongoDB do desenvolvimento
# Não é necessária limpeza especial, os testes criam/removem dados automaticamente
```

## API Endpoints

### Leilões

- `POST /auction` - Criar novo leilão
- `GET /auction?status={0|1}&category={categoria}&product_name={nome}` - Listar leilões (status: 0=Ativo, 1=Completo)
- `GET /auction/{id}` - Buscar leilão por ID
- `GET /auction/winner/{auctionId}` - Buscar lance vencedor de um leilão

### Lances

- `POST /bid` - Criar novo lance
- `GET /bid/{auctionId}` - Listar lances de um leilão

### Usuários

- `GET /user/{id}` - Buscar usuário por ID

## Exemplos de Uso da API

### Criar um leilão

```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 15 Pro",
    "category": "Electronics", 
    "description": "Brand new iPhone 15 Pro 128GB",
    "condition": 1
  }'
```

### Listar leilões ativos

```bash
curl -X GET "http://localhost:8080/auction?status=0"
```

### Listar leilões completos

```bash
curl -X GET "http://localhost:8080/auction?status=1"
```

### Buscar leilão específico

```bash
curl -X GET "http://localhost:8080/auction/{auction-id}"
```

## Estrutura do Projeto

```
.
├── cmd/auction/                 # Ponto de entrada da aplicação
├── configuration/               # Configurações (DB, Logger, etc)
├── internal/
│   ├── entity/                 # Entidades de domínio
│   ├── infra/
│   │   ├── api/web/            # Controllers e validações
│   │   └── database/           # Repositórios e acesso a dados
│   ├── usecase/                # Casos de uso/regras de negócio
│   └── internal_error/         # Tratamento de erros
├── docker-compose.yml          # Configuração para produção
├── .gitignore                  # Arquivos ignorados pelo git
└── README.md
```

## Logs e Monitoramento

O sistema gera logs automáticos quando:
- Um leilão é fechado automaticamente
- Ocorrem erros na verificação de leilões
- Há problemas na atualização de status

Exemplo de log:
```
INFO: Auction closed automatically: auction-id-123
```

## Tecnologias Utilizadas

- **Go 1.20**: Linguagem principal
- **Gin**: Framework web
- **MongoDB**: Banco de dados
- **Go Routines**: Concorrência para fechamento automático
- **Zap**: Sistema de logs
- **Testify**: Framework de testes
- **Docker**: Containerização

## Contribuição

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-funcionalidade`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova funcionalidade'`)
4. Push para a branch (`git push origin feature/nova-funcionalidade`)
5. Abra um Pull Request

## Licença

Este projeto está sob a licença MIT. Veja o arquivo LICENSE para mais detalhes.