#!/bin/bash

sudo_cmd=

KERNEL_NAME=$(uname -s)
ARCHITECTURE=$(uname -m)
ORYA_FILES_PATH=/opt/orya-agent
USER_GROUP_NAME=orya-agent

# Root user detection
if [ "$UID" == "0" ]; then
    sudo_cmd=''
else
    sudo_cmd='sudo'
fi

#stop and disable systemd
function stop_and_disable() {
  printf "Stopping and disabling systemd service...\n"
  $sudo_cmd systemctl stop orya-manager
  $sudo_cmd systemctl disable orya-manager
}

#remove file service
function remove_file_service(){
  printf "Removing systemd service file...\n"
  $sudo_cmd rm /etc/systemd/system/orya-manager.service
}

#remove work dir
function remove_work_dir() {
  printf "Removing work directory...\n"
  $sudo_cmd rm -rf $ORYA_FILES_PATH 
}

# remove user and group
function remove_user_and_group(){
  printf "Removing user and group...\n"
  $sudo_cmd userdel $USER_GROUP_NAME > /dev/null 2>&1
  $sudo_cmd groupdel $USER_GROUP_NAME > /dev/null 2>&1
}

# remove perm sudoers file
function remove_perm_sudoers_file(){
  printf "Removing sudoers file entry...\n"
  $sudo_cmd sed -i '/^orya-agent ALL=(ALL) NOPASSWD: ALL/d' /etc/sudoers
}

#setup configure e verify machine
function setup(){
  verify_kernel
}

# uninstaller the manager
function _uninstaller(){
  stop_and_disable
  remove_file_service
  remove_work_dir
  remove_user_and_group
  remove_perm_sudoers_file
}

#uninstall manager
function uninstall(){
  setup
 _uninstaller 
}

#verify is linux kernel
function verify_kernel(){
if [[ "$KERNEL_NAME" != "Linux" ]]; then
  printf "\033[31mInvalid uninstaller for machine\033[0m\n"
  exit 0
fi
}


#actions
uninstall
