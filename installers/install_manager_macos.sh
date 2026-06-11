
#!/bin/bash

sudo_cmd=

KERNEL_NAME=$(uname -s)
ARCHITECTURE=$(uname -m)
FILE_INDEX_URL="https://docp-agent.s3.us-east-1.amazonaws.com/index_os_instance.json"
BINARY_URL="https://github.com/OryaHub/agent-os-instance/releases/download"
VERSION="latest"
MANAGER_IS_RUNNING=$(ps aux | grep -v grep | grep orya-agent/bin/manager)
ORYA_FILES_PATH=/opt/orya-agent
USER_GROUP_NAME=orya-agent

# Root user detection
if [ "$UID" == "0" ]; then
    sudo_cmd=''
else
    sudo_cmd='sudo'
fi

#initalize options
apiKey=""
tags=""
vmName=""
noGroupAssociation="false"
oryaSite="https://msapi.orya.tech"
#usage show default usage mode 
function usage() {
    echo "USAGE: $0 --api_key <apikey> --tags <tag:1,tag:2>"
    echo "Exemplo: $0 --api_key \"xpto\" --tags \"grupo:app,maquina:dev\" "
    exit 1
}

#analize options
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in --api_key)
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
      --orya_site)
        oryaSite="$2"
        shift
        shift
        ;;
      --vm_name)
        vmName="$2"
        shift
        shift
        ;;
      --no_group_association)
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
  printf "\033[31mAgent already running\033[0m\n"
  exit 0
fi
}

#verify is macOS kernel
function verify_kernel(){
if [[ "$KERNEL_NAME" != "Darwin" ]]; then
  printf "\033[31mInvalid installer for this machine\033[0m\n"
  exit 0
fi
}

# -------------------------------------------------------------------
# verify_orya_site: checks if Orya Site is reachable (DNS + connectivity)
#                   scenario 3 (network/DNS / wrong domain)
# -------------------------------------------------------------------
function verify_orya_site() {
    printf "Checking access to Orya Site...\n"

    curl -s -o /dev/null --connect-timeout 10 --max-time 15 "$oryaSite" 2>&1
    local exit_code=$?

    if [[ $exit_code -eq 6 ]]; then
        printf "\033[31mError: Domain not found - %s\n" "$oryaSite"
        printf "Check if the --orya_site URL is correct.\033[0m\n"
        exit 1
    elif [[ $exit_code -eq 7 ]]; then
        printf "\033[31mError: Connection refused by %s\n" "$oryaSite"
        printf "The server may be down or the address is incorrect.\033[0m\n"
        exit 1
    elif [[ $exit_code -eq 28 ]]; then
        printf "\033[31mError: Connection timeout for %s\n" "$oryaSite"
        printf "Check your network connection.\033[0m\n"
        exit 1
    elif [[ $exit_code -ne 0 ]]; then
        printf "\033[31mError: Failed to access %s (curl exit code: %d).\n" "$oryaSite" $exit_code
        printf "Check the address and your network connection.\033[0m\n"
        exit 1
    fi

    printf "Orya Site reachable.\n"
}

function verify_usage_limit(){
    local url="$oryaSite/usage-tracking/usage-limit/orya/check"
    printf "Checking usage limit...\n"

    local tmpfile=$(mktemp)
    local http_code=$(curl -s -o "$tmpfile" -w "%{http_code}" --connect-timeout 10 --max-time 30 "$url" -H "docp-api-key: $apiKey" 2>&1)
    local exit_code=$?
    local body=$(cat "$tmpfile" 2>/dev/null)
    rm -f "$tmpfile"

    # Scenario 3 (fallback): network error, if connectivity dropped between steps
    if [[ $exit_code -ne 0 || "$http_code" == "000" ]]; then
        printf "\033[31mError: Could not connect to service at %s.\n" "$url"
        printf "Check your network connection and try again.\033[0m\n"
        exit 1
    fi

    # Scenario 4: backend error
    if [[ "$http_code" -ge 500 ]]; then
        printf "\033[31mError: The Orya backend is temporarily unavailable (HTTP %s).\n" "$http_code"
        printf "Please wait a few moments and try again.\033[0m\n"
        exit 1
    fi

    # Scenario 1: invalid API Key — API returns 404/401/403 with {"detail":{"message":"Api key ... not found"}}
    if [[ "$http_code" == "404" || "$http_code" == "401" || "$http_code" == "403" ]]; then
        printf "\033[31mError: Invalid API Key or Orya Site.\n"
        printf "Check the --api_key or --orya_site parameters.\033[0m\n"
        exit 1
    fi

    if [[ "$http_code" != "200" ]]; then
        printf "\033[31mError: Unexpected response while checking usage limit (HTTP %s).\n" "$http_code"
        printf "Please try again.\033[0m\n"
        exit 1
    fi

    # Scenario 2: limit exceeded
    has_limit=$(echo "$body" | grep -o '"has_limit":[^,}]*' | cut -d: -f2)
    limit=$(echo "$body" | grep -o '"limit":[^,}]*' | cut -d: -f2)

    if [[ "$has_limit" == "false" ]]; then
        printf "\033[31mUsage limit exceeded.\n"
        printf "You have reached the maximum number of agents for your plan.\n"
        if [[ -n "$limit" && "$limit" != "0" ]]; then
            printf "Current limit: %s\n" "$limit"
        fi
        printf "Increase your limits on the Orya platform.\033[0m\n"
        exit 1
    fi

    printf "Usage limit OK.\n"
}

#get binary arm64
function get_binary_arch64(){
  printf "Downloading binary for arm64...\n"
  sudo curl -sS -L -o $ORYA_FILES_PATH/bin/releases/$VERSION/manager "$BINARY_URL/$VERSION/manager-macos-arm64" 2>&1
  local exit_code=$?
  if [[ $exit_code -ne 0 ]]; then
    printf "\033[31mError: Failed to download binary (curl exit code: %d).\n" $exit_code
    printf "Check your network connection and try again.\033[0m\n"
    exit 1
  fi
  sudo chmod +x $ORYA_FILES_PATH/bin/releases/$VERSION/manager || {
    printf "\033[31mLocal system error: failed to set execute permission on binary.\033[0m\n"
    exit 1
  }
}

#get binary amd64
function get_binary_amd64(){
  printf "Downloading binary for amd64...\n"
  sudo curl -sS -L -o $ORYA_FILES_PATH/bin/releases/$VERSION/manager "$BINARY_URL/$VERSION/manager-macos-amd64" 2>&1
  local exit_code=$?
  if [[ $exit_code -ne 0 ]]; then
    printf "\033[31mError: Failed to download binary (curl exit code: %d).\n" $exit_code
    printf "Check your network connection and try again.\033[0m\n"
    exit 1
  fi
  sudo chmod +x $ORYA_FILES_PATH/bin/releases/$VERSION/manager || {
    printf "\033[31mLocal system error: failed to set execute permission on binary.\033[0m\n"
    exit 1
  }
}

#verify is architecture and get binary
function verify_architecture(){
printf "Checking system architecture...\n"
if [[ "$ARCHITECTURE" == "arm64" ]]; then
  get_binary_arch64
fi
if [[ "$ARCHITECTURE" == "x86_64" ]]; then
  get_binary_amd64
fi
}


#setup configure and verify machine
function setup(){
  verify_kernel
  verify_architecture
}

#create directories for orya agent
function create_workdir(){
  printf "Creating necessary directories and files...\n"
  [[ ! -d $ORYA_FILES_PATH ]] && $sudo_cmd mkdir $ORYA_FILES_PATH 
  [[ ! -d $ORYA_FILES_PATH/bin ]] && $sudo_cmd mkdir $ORYA_FILES_PATH/bin 
  [[ ! -d $ORYA_FILES_PATH/logs ]] && $sudo_cmd mkdir $ORYA_FILES_PATH/logs 
  [[ ! -d $ORYA_FILES_PATH/state ]] && $sudo_cmd mkdir $ORYA_FILES_PATH/state 
  [[ ! -f $ORYA_FILES_PATH/environments ]] && $sudo_cmd touch $ORYA_FILES_PATH/environments 
  [[ ! -f $ORYA_FILES_PATH/state/current ]] && $sudo_cmd touch $ORYA_FILES_PATH/state/current 
  [[ ! -f $ORYA_FILES_PATH/state/received ]] && $sudo_cmd touch $ORYA_FILES_PATH/state/received 
}

#save_environments in directory
function save_environments(){
  printf "Saving environment variables...\n"
  api_key=$1
  tgs=$2
  orya_site=$3
  printf "ORYA_API_KEY=$api_key\nORYA_TAGS=$tgs\nORYA_DOMAIN=$orya_site\nORYA_AGENT_PORT=12012\n" | sudo tee $ORYA_FILES_PATH/environments > /dev/null
}

# Resolve version, extracting the "latest" field from JSON
function resolve_version() {
  local version="$1"
  if [[ "$version" == "latest" ]]; then
    version=$(curl -s "$FILE_INDEX_URL" | sed -n 's/.*"latest"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
    if [[ -z "$version" ]]; then
      echo "Failed to fetch the latest version." >&2
      return 1
    fi
  fi
  echo "$version"
}

#set content service
function set_content_service() {
  printf "Setting up launchd service...\n"
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
        <string>tech.orya.manager</string>
        <key>EnvironmentVariables</key>
        <dict>
            <key>ORYA_AGENT_PORT</key>
            <string>12012</string>
            <key>LOG_LEVEL</key>
            <string>info</string>
            <key>ERROR_LEVEL</key>
            <string>high</string>
        </dict>
        <key>ProgramArguments</key>
        <array>
            <string>/opt/orya-agent/bin/current/manager</string>
        </array>
        <key>StandardOutPath</key>
        <string>/opt/orya-agent/logs/launchd.log</string>
        <key>StandardErrorPath</key>
        <string>/opt/orya-agent/logs/launchd.log</string>
        <key>ExitTimeOut</key>
        <integer>10</integer>
    </dict>
    </plist>' | sudo tee ~/Library/LaunchAgents/tech.orya.manager.plist > /dev/null
}

#prepare launchd (scenario 5 - OS errors)
function prepare_launchd() {
  printf "Configuring and running service...\n"
  launchctl load ~/Library/LaunchAgents/tech.orya.manager.plist || {
    printf "\033[31mLocal system error: failed to load launchctl.\n"
    printf "Check if launchd is available and try again.\033[0m\n"
    exit 1
  }
  launchctl start gui/$(id -u)/tech.orya.manager || {
    printf "\033[31mLocal system error: failed to start orya-manager service.\n"
    printf "Check ~/Library/LaunchAgents/tech.orya.manager.plist for details.\033[0m\n"
    exit 1
  }
}

#create config yml
function create_config_yml(){
  $sudo_cmd touch $ORYA_FILES_PATH/config.yml
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

  # Remove trailing comma
  echo "${processed_tags%,}"
}

#add content config yml
function add_content_config_yml(){
  printf "Adding content to config.yml...\n"
  api_key="$1"
  tags=$(process_tags "$2")
  version="$3"
  noGroupAssociation="$4"
$sudo_cmd tee $ORYA_FILES_PATH/config.yml > /dev/null <<EOF  
# Orya file for agent configuration that contains
# information used to perform configuration and service.

no_group_association: $noGroupAssociation
version: $version 

agent:
  api_key: $api_key 
  tags: 
    $(echo $tags | sed 's/,/\n    /g')
EOF
}
#actions
printf "Resolving version...\n"
VERSION=$(resolve_version "$VERSION") || exit 1
verify_script
printf "Checking if the agent is already running...\n"
already_running
verify_orya_site
verify_usage_limit
setup
create_workdir
create_config_yml
save_environments $apiKey $tags $oryaSite
add_content_config_yml $apiKey $tags $VERSION $noGroupAssociation
set_content_service
prepare_launchd
printf "\033[32mInstallation completed successfully.\033[0m\n"

