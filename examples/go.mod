module examples

go 1.24.0

require (
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/sshturbo/GoTeleMD v0.0.0-00010101000000-000000000000
)

// Use o módulo local para desenvolvimento
replace github.com/sshturbo/GoTeleMD => ../
