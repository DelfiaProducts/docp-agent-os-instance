## Instalação

Processo de instalação dos agents(manager/agent) docp.

### Requerimentos

- Golang
- WixToolset

## Manager

Processo do Manager

Todos os arquivos necessarios pro build precisam estar na mesma pasta.

- config_manager.exe
- manager.exe
- install_manager_windows.xml

### Processo

Build do arquivo de configuração pro manager windows.

```
GOOS=windows GOARCH=amd64 go build -o config_manager.exe install_manager_windows.go
```

Navegar ate o repositorio do manager e realizar o build.

```
GOOS=windows GOARCH=amd64 go build -o manager.exe main.go
```

### Build

Construção com wix toolset

```
candle.exe install_manager_windows.xml -o install_manager_windows.wixobj
light.exe install_manager_windows.wixobj -o install_manager_windows.msi

```

### Instalação

Instalar o msi gerado na maquina local

```
Start-Process -Wait msiexec -ArgumentList '/qn /i install_manager_windows.msi API_KEY="xpto" TAGS="app:dev"'

```

## Agent

Processo do Agent

Todos os arquivos necessarios pro build precisam estar na mesma pasta.

- config_agent.exe
- agent.exe
- install_agent_windows.xml

### Processo

Build do arquivo de configuração pro agent windows.

```
GOOS=windows GOARCH=amd64 go build -o config_agent.exe install_agent_windows.go
```

Navegar ate o repositorio do agent e realizar o build.

```
SCM=agent GOOS=windows GOARCH=amd64 go build -o agent.exe main.go
```

### Build

Construção com wix toolset

```
candle.exe install_agent_windows.xml -o install_agent_windows.wixobj
light.exe install_agent_windows.wixobj -o install_agent_windows.msi

```

Mover o msi do agent pro bucket do s3
