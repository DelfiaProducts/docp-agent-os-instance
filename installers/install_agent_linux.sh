#!/bin/bash

sudo_cmd=

KERNEL_NAME=$(uname -s)
ARCHITECTURE=$(uname -m)
FILE_INDEX_URL="https://docp-agent-k8s.s3.us-east-1.amazonaws.com/index.json"
BINARY_URL="https://test-docp-agent-data.s3.amazonaws.com/agent"
VERSION="${VERSION:-${VERSION:-latest}}"
AGENT_IS_RUNNING=$($sudo_cmd systemctl is-active docp-agent)
DOCP_FILES_PATH=/opt/docp-agent
USER_GROUP_NAME=docp-agent

# Root user detection
if [ "$UID" == "0" ]; then
    sudo_cmd=''
else
    sudo_cmd='sudo'
fi

#analize options
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in --version)
        VERSION="$2"
        shift 
        shift 
        ;;
    esac
done

# verify is already running manager
function already_running(){
if [[ "$AGENT_IS_RUNNING" == "active" ]]; then
  printf "\033[31mAlready running agent\033[0m\n"
  exit 0
fi
}

#verify is linux kernel
function verify_kernel(){
if [[ "$KERNEL_NAME" != "Linux" ]]; then
  printf "\033[31mInvalid installer for machine\033[0m\n"
  exit 0
fi
}

#verify is architecture and get binary
function verify_architecture(){
if [[ "$ARCHITECTURE" == "aarch64" ]]; then
  get_binary_arch64
fi
if [[ "$ARCHITECTURE" == "x86_64" ]]; then
  get_binary_amd64
fi
}

#setup configure e verify machine
function setup(){
  verify_kernel
  already_running
}

# Corrige a função resolve_version para extrair corretamente o campo "latest" do JSON
function resolve_version() {
  local version="$1"
  if [[ "$version" == "latest" ]]; then
    version=$(curl -s "$FILE_INDEX_URL" | sed -n 's/.*"latest"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
    if [[ -z "$version" ]]; then
      echo "Failed to fetch latest version." >&2
      exit 1
    fi
  fi
  echo "$version"
}

#add permission workdir
function add_perm_work_dir(){
  $sudo_cmd chown -R $USER_GROUP_NAME:$USER_GROUP_NAME /opt/docp-agent/
}

#get binary arm64
function get_binary_arch64(){
  sudo curl -s -o $DOCP_FILES_PATH/bin/releases/$VERSION/agent "$BINARY_URL/$VERSION/linux_arm64"
  sudo chmod +x $DOCP_FILES_PATH/bin/releases/$VERSION/agent
}
#get binary amd64
function get_binary_amd64(){
  sudo curl -s -o $DOCP_FILES_PATH/bin/releases/$VERSION/agent "$BINARY_URL/$VERSION/linux_amd64"
  sudo chmod +x $DOCP_FILES_PATH/bin/releases/$VERSION/agent
}
# create symbolic link
function create_link_simbolic(){
  sudo ln -sfn $DOCP_FILES_PATH/bin/releases/$VERSION/agent $DOCP_FILES_PATH/bin/current/agent
}

#set content service
function set_content_service() {
  printf "[Unit]\nDescription=Docp Agent\nAfter=network.target\n\n[Service]\nType=simple\nPIDFile=/opt/docp-agent/run/agent.pid\nUser=docp-agent\nRestart=on-failure\nEnvironmentFile=-/opt/docp-agent/environments\nRuntimeDirectory=docp\nExecStart=/opt/docp-agent/bin/current/agent run -p /opt/docp-agent/run/agent.pid\nStartLimitInterval=10\nStartLimitBurst=5\nStandardOutput=journal\nStandardError=journal\n\n[Install]\nWantedBy=multi-user.target\n" | sudo tee /etc/systemd/system/docp-agent.service > /dev/null
}

#prepare systemd
function prepare_systemd() {
  $sudo_cmd systemctl daemon-reload
  $sudo_cmd systemctl start docp-agent.service
  $sudo_cmd systemctl enable docp-agent.service
}
#actions
VERSION=$(resolve_version "$VERSION")
setup
verify_architecture
create_link_simbolic
add_perm_work_dir
set_content_service
prepare_systemd
