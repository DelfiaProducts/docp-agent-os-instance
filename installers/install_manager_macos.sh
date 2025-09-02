
#!/bin/bash

sudo_cmd=

KERNEL_NAME=$(uname -s)
ARCHITECTURE=$(uname -m)
BINARY_URL="https://test-docp-agent-data.s3.amazonaws.com/manager"
VERSION="latest"
MANAGER_IS_RUNNING=$(ps aux | grep -v grep | grep docp-agent/bin/manager)
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
if [[ "$MANAGER_IS_RUNNING" ]]; then
  printf "\033[31mAlready running agent\033[0m\n"
  exit 0
fi
}

#verify is linux kernel
function verify_kernel(){
if [[ "$KERNEL_NAME" != "Darwin" ]]; then
  printf "\033[31mInvalid installer for machine\033[0m\n"
  exit 0
fi
}

#get binary arm64
function get_binary_arch64(){
  sudo curl -s -o $DOCP_FILES_PATH/bin/manager "$BINARY_URL/$VERSION/macos_arm64"
  sudo chmod +x $DOCP_FILES_PATH/bin/manager
}

#get binary amd64
function get_binary_amd64(){
  sudo curl -s -o $DOCP_FILES_PATH/bin/manager "$BINARY_URL/$VERSION/macos_amd64"
  sudo chmod +x $DOCP_FILES_PATH/bin/manager
}

#verify is architecture and get binary
function verify_architecture(){
if [[ "$ARCHITECTURE" == "arm64" ]]; then
  get_binary_arch64
fi
if [[ "$ARCHITECTURE" == "x86_64" ]]; then
  get_binary_amd64
fi
}


#setup configure e verify machine
function setup(){
  verify_kernel
  verify_architecture
  already_running
}

#create directories for docp agent
function create_workdir(){
  [[ ! -d $DOCP_FILES_PATH ]] && $sudo_cmd mkdir $DOCP_FILES_PATH 
  [[ ! -d $DOCP_FILES_PATH/bin ]] && $sudo_cmd mkdir $DOCP_FILES_PATH/bin 
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
  printf "DOCP_API_KEY=$api_key\nDOCP_TAGS=$tgs\nDOCP_REGISTER_URL=https://msapi.sandbox.docphq.tech/agents\nDOCP_STATE_CHECK_URL=https://msapi.sandbox.docphq.tech/agents\nDOCP_AGENT_PORT=12012\n" | sudo tee $DOCP_FILES_PATH/environments > /dev/null
}


#set content service
function set_content_service() {
  printf '
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
    <dict>
        <key>KeepAlive</key>
        <dict>
            <key>SuccessfulExit</key>
            <false/>
        </dict>
        <key>Label</key>
        <string>com.docp.manager</string>
        <key>EnvironmentVariables</key>
        <dict>
            <key>DOCP_REGISTER_URL</key>
            <string>https://msapi.sandbox.docphq.tech/agents</string>
            <key>DOCP_STATE_CHECK_URL</key>
            <string>https://msapi.sandbox.docphq.tech/agents</string>
            <key>DOCP_AGENT_PORT</key>
            <string>12012</string>
            <key>LOG_LEVEL</key>
            <string>info</string>
            <key>ERROR_LEVEL</key>
            <string>high</string>
        </dict>
        <key>ProgramArguments</key>
        <array>
            <string>/opt/docp-agent/bin/manager</string>
        </array>
        <key>StandardOutPath</key>
        <string>/opt/docp-agent/logs/launchd.log</string>
        <key>StandardErrorPath</key>
        <string>/opt/docp-agent/logs/launchd.log</string>
        <key>ExitTimeOut</key>
        <integer>10</integer>
    </dict>
    </plist>' | sudo tee ~/Library/LaunchAgents/com.docp.manager.plist > /dev/null
}

#prepare launchd
function prepare_launchd() {
  launchctl load ~/Library/LaunchAgents/com.docp.manager.plist
  launchctl start gui/$(id -u)/com.docp.manager
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
verify_script
setup
create_workdir
create_config_yml
save_api_key_and_tags $apiKey $tags
add_content_config_yml $apiKey $tags $VERSION $noGroupAssociation
set_content_service
prepare_launchd
