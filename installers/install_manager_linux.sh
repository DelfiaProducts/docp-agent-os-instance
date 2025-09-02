#!/bin/bash

sudo_cmd=

KERNEL_NAME=$(uname -s)
ARCHITECTURE=$(uname -m)
FILE_INDEX_URL="https://test-docp-agent-data.s3.amazonaws.com/index.json"
BINARY_URL="https://test-docp-agent-data.s3.amazonaws.com/manager"
VERSION="latest"
MANAGER_IS_RUNNING=$($sudo_cmd systemctl is-active docp-manager)
DOCP_FILES_PATH=/opt/docp-agent
USER_GROUP_NAME=docp-agent

# Root user detection
if [ "$UID" == "0" ]; then
    sudo_cmd=''
else
    sudo_cmd='sudo'
fi

#initalize options
apiKey=""
tags=""
noGroupAssociation="false"
#usage show default usage mode 
function usage() {
    echo "USAGE: $0 --apiKey <apikey> --tags <tag:1,tag:2>"
    echo "Example: $0 --apiKey \"xpto\" --tags \"group:app,machine:dev\" "
    exit 1
}

#analize options
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in --apiKey)
        apiKey="$2"
        shift 
        shift 
        ;;
        --tags)
        tags="$2"
        shift 
        shift 
        ;;
      --version)
        VERSION="$2"
        shift
        shift
        ;;
      --no-group-association)
        noGroupAssociation="true"
        shift
        shift
        ;;
        *)
        usage
        ;;
    esac
done

# Verify required options
function verify_usage(){
if [[ -z "$apiKey" || -z "$tags" ]]; then
   usage
    exit 0
fi
}

#verify script execute verification the script runnin
function verify_script(){
  verify_usage
}

# verify is already running manager
function already_running(){
if [[ "$MANAGER_IS_RUNNING" == "active" ]]; then
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

#create group 
function create_group(){
  $sudo_cmd groupadd $USER_GROUP_NAME > /dev/null 2>&1
}

#add user to group 
function add_user_to_group(){
  $sudo_cmd useradd -m -g $USER_GROUP_NAME -s /bin/bash $USER_GROUP_NAME > /dev/null 2>&1
}

# add perm sudoers file
function add_perm_sudoers_file(){
  $sudo_cmd echo 'docp-agent ALL=(ALL) NOPASSWD: ALL' | sudo tee -a /etc/sudoers > /dev/null 2>&1
}

#add permission workdir
function add_perm_work_dir(){
  $sudo_cmd chown -R $USER_GROUP_NAME:$USER_GROUP_NAME /opt/docp-agent/
}

#setup configure e verify machine
function setup(){
  verify_kernel
  already_running
  create_group
  add_user_to_group
  add_perm_sudoers_file
}

#create directories for docp agent
function create_workdir(){
  [[ ! -d $DOCP_FILES_PATH ]] && $sudo_cmd mkdir $DOCP_FILES_PATH 
  [[ ! -d $DOCP_FILES_PATH/bin ]] && $sudo_cmd mkdir $DOCP_FILES_PATH/bin 
  [[ ! -d $DOCP_FILES_PATH/bin/current ]] && $sudo_cmd mkdir $DOCP_FILES_PATH/bin/current 
  [[ ! -d $DOCP_FILES_PATH/bin/releases ]] && $sudo_cmd mkdir $DOCP_FILES_PATH/bin/releases 
  [[ ! -d $DOCP_FILES_PATH/bin/releases/$VERSION ]] && $sudo_cmd mkdir $DOCP_FILES_PATH/bin/releases/$VERSION 
  [[ ! -d $DOCP_FILES_PATH/logs ]] && $sudo_cmd mkdir $DOCP_FILES_PATH/logs 
  [[ ! -d $DOCP_FILES_PATH/state ]] && $sudo_cmd mkdir $DOCP_FILES_PATH/state 
  [[ ! -f $DOCP_FILES_PATH/environments ]] && $sudo_cmd touch $DOCP_FILES_PATH/environments 
  [[ ! -f $DOCP_FILES_PATH/state/current ]] && $sudo_cmd touch $DOCP_FILES_PATH/state/current 
  [[ ! -f $DOCP_FILES_PATH/state/received ]] && $sudo_cmd touch $DOCP_FILES_PATH/state/received 
}

#save apiKey and tags in directory
function save_api_key_and_tags(){
  api_key=$1
  tgs=$2
  printf "DOCP_API_KEY=$api_key\nDOCP_TAGS=$tgs\nDOCP_DOMAIN=https://msapi.sandbox.docphq.tech\nDOCP_AGENT_PORT=12012\n" | sudo tee $DOCP_FILES_PATH/environments > /dev/null
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


#get binary arm64
function get_binary_arch64(){
  sudo curl -s -o $DOCP_FILES_PATH/bin/releases/$VERSION/manager "$BINARY_URL/$VERSION/linux_arm64"
  sudo chmod +x $DOCP_FILES_PATH/bin/releases/$VERSION/manager
}
#get binary amd64
function get_binary_amd64(){
  sudo curl -s -o $DOCP_FILES_PATH/bin/releases/$VERSION/manager "$BINARY_URL/$VERSION/linux_amd64"
  sudo chmod +x $DOCP_FILES_PATH/bin/releases/$VERSION/manager
}

# create symbolic link
function create_link_simbolic(){
  sudo ln -sfn $DOCP_FILES_PATH/bin/releases/$VERSION/manager $DOCP_FILES_PATH/bin/current/manager
}

#set content service
function set_content_service() {
  printf "[Unit]\nDescription=Docp Manager\nAfter=network.target\n\n[Service]\nType=simple\nPIDFile=/opt/docp-agent/run/manager.pid\nUser=docp-agent\nRestart=on-failure\nEnvironmentFile=-/opt/docp-agent/environments\nRuntimeDirectory=docp\nExecStart=/opt/docp-agent/bin/current/manager run -p /opt/docp-agent/run/manager.pid\nStartLimitInterval=10\nStartLimitBurst=5\nStandardOutput=journal\nStandardError=journal\n\n[Install]\nWantedBy=multi-user.target\n" | sudo tee /etc/systemd/system/docp-manager.service > /dev/null
}

#prepare systemd
function prepare_systemd() {
  sudo systemctl daemon-reload
  sudo systemctl start docp-manager.service
  sudo systemctl enable docp-manager.service
}

#create config yml
function create_config_yml(){
  $sudo_cmd touch $DOCP_FILES_PATH/config.yml
}

#process tags
function process_tags() {
  local tag_string="$1"
  local IFS=',' 
  local processed_tags=""

  for tag in $tag_string; do
    if [[ ! $tag =~ ":" ]]; then
      tag="$tag: default" 
    fi
    if [[ $tag =~ ":" ]]; then
      tag=$(echo "$tag" | sed 's/:\([^ ]\)/: \1/g')
    fi
    processed_tags="${processed_tags}${tag},"
  done

  # Remove a vírgula final
  echo "${processed_tags%,}"
}

#add content config yml
function add_content_config_yml(){
  api_key="$1"
  tags=$(process_tags "$2")
  version="$3"
  noGroupAssociation="$4"
$sudo_cmd tee $DOCP_FILES_PATH/config.yml > /dev/null <<EOF  
# Docp file for agent configuration that contains
# information used to perform configuration and service.

no_group_association: $noGroupAssociation
version: $version 

agent:
  apiKey: $api_key 
  tags: 
    $(echo $tags | sed 's/,/\n    /g')
EOF
}
#actions
VERSION=$(resolve_version "$VERSION")
verify_script
setup
create_workdir
create_config_yml
save_api_key_and_tags $apiKey $tags
verify_architecture
create_link_simbolic
add_content_config_yml $apiKey $tags $VERSION $noGroupAssociation
add_perm_work_dir
set_content_service
prepare_systemd
