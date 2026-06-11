#!/bin/bash

sudo_cmd=

KERNEL_NAME=$(uname -s)
ARCHITECTURE=$(uname -m)
FILE_INDEX_URL="https://docp-agent.s3.us-east-1.amazonaws.com/index_os_instance.json"
BINARY_URL="https://github.com/OryaHub/agent-os-instance/releases/download"
VERSION="${VERSION:-${VERSION:-latest}}"
AGENT_IS_RUNNING=$($sudo_cmd systemctl is-active orya-agent)
ORYA_FILES_PATH=/opt/orya-agent
USER_GROUP_NAME=orya-agent

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
printf "Checking system architecture...\n"
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
      return 1
    fi
  fi
  echo "$version"
}

#add permission workdir
function add_perm_work_dir(){
  printf "Adding permissions to work directory...\n"
  $sudo_cmd chown -R $USER_GROUP_NAME:$USER_GROUP_NAME /opt/orya-agent/
}

#get binary arm64
function get_binary_arch64(){
  printf "Downloading binary for arm64 architecture...\n"
  sudo curl -s -L -o $ORYA_FILES_PATH/bin/releases/$VERSION/agent "$BINARY_URL/$VERSION/agent-linux-arm64"
  sudo chmod +x $ORYA_FILES_PATH/bin/releases/$VERSION/agent
}
#get binary amd64
function get_binary_amd64(){
  printf "Downloading binary for amd64 architecture...\n"
  sudo curl -s -L -o $ORYA_FILES_PATH/bin/releases/$VERSION/agent "$BINARY_URL/$VERSION/agent-linux-amd64"
  sudo chmod +x $ORYA_FILES_PATH/bin/releases/$VERSION/agent
}
# create symbolic link
function create_link_simbolic(){
  printf "Creating symbolic link for agent...\n"
  sudo ln -sfn $ORYA_FILES_PATH/bin/releases/$VERSION/agent $ORYA_FILES_PATH/bin/current/agent
}

#set content service
function set_content_service() {
  printf "Creating systemd service file...\n"
  printf "[Unit]\nDescription=Orya Agent\nAfter=network.target\n\n[Service]\nType=simple\nPIDFile=/opt/orya-agent/run/agent.pid\nUser=orya-agent\nRestart=on-failure\nEnvironmentFile=-/opt/orya-agent/environments\nRuntimeDirectory=orya\nExecStart=/opt/orya-agent/bin/current/agent run -p /opt/orya-agent/run/agent.pid\nStartLimitInterval=10\nStartLimitBurst=5\nStandardOutput=journal\nStandardError=journal\n\n[Install]\nWantedBy=multi-user.target\n" | sudo tee /etc/systemd/system/orya-agent.service > /dev/null
}

#prepare systemd
function prepare_systemd() {
  printf "Setting up systemd service...\n"
  $sudo_cmd systemctl daemon-reload
  $sudo_cmd systemctl start orya-agent.service
  $sudo_cmd systemctl enable orya-agent.service
}
#actions
printf "Resolving version...\n"
VERSION=$(resolve_version "$VERSION") || exit 1
setup
verify_architecture
create_link_simbolic
add_perm_work_dir
set_content_service
prepare_systemd
printf "\033[32mAgent installed successfully\033[0m\n"
